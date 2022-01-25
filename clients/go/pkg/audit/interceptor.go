// Copyright 2022 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	calpb "google.golang.org/genproto/googleapis/cloud/audit"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/abcxyz/lumberjack/clients/go/pkg/zlogger"
)

// Interceptor contains the fields required for an interceptor
// to autofill and emit audit logs.
type Interceptor struct {
	*Client
	SecurityContext security.GRPCContext
	Rules           []*alpb.AuditRule
}

// UnaryInterceptor is a gRPC unary interceptor that automatically emits application audit logs.
// TODO(#95): add support for fail-close/best-effort logging.
func (i *Interceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	zlogger := zlogger.FromContext(ctx)
	mostRelevantRule := mostRelevantRule(info.FullMethod, i.Rules)
	if mostRelevantRule == nil {
		zlogger.Infow("no audit rule matching method name", "audit rules", i.Rules, "method name", info.FullMethod)
		return handler(ctx, req)
	}

	logReq := &alpb.AuditLogRequest{Payload: &calpb.AuditLog{}}

	// Set log type.
	logReq.Type = alpb.AuditLogRequest_UNSPECIFIED
	if t, ok := alpb.AuditLogRequest_LogType_value[mostRelevantRule.LogType]; ok {
		logReq.Type = alpb.AuditLogRequest_LogType(t)
	}

	// Autofill `Payload.ServiceName` and `Payload.MethodName`.
	logReq.Payload.MethodName = info.FullMethod
	serviceName, err := serviceName(info.FullMethod)
	if err != nil {
		zlogger.Warnw("unary interceptor failed extract service name", "error", err)
		return handler(ctx, req)
	}
	logReq.Payload.ServiceName = serviceName

	// Autofill `Payload.AuthenticationInfo.PrincipalEmail`.
	principal, err := i.SecurityContext.RequestPrincipal(ctx)
	if err != nil {
		zlogger.Warnw("unary interceptor failed getting principal from ctx", "ctx", ctx, "security context", i.SecurityContext, "error", err)
		return handler(ctx, req)
	}
	logReq.Payload.AuthenticationInfo = &calpb.AuthenticationInfo{PrincipalEmail: principal}

	// Autofill `Payload.Request`.
	d := mostRelevantRule.Directive
	if d == alpb.AuditRuleDirectiveRequestAndResponse || d == alpb.AuditRuleDirectiveRequestOnly {
		if reqStruct, err := toProtoStruct(req); err != nil {
			zlogger.Warnw("unary interceptor failed converting req into a Google struct proto", "error", err)
		} else {
			logReq.Payload.Request = reqStruct
		}
	}

	// Store our log req in the context to make it accessible
	// to the handler source code.
	ctx = context.WithValue(ctx, auditLogReqKey{}, logReq)

	// Execute the handler. The handler can modify the log
	// req in the context. For example, the handler can:
	//   - overwrite a log req field we set previously
	//   - fill the field `Payload.ResourceName`
	handlerResp, handlerErr := handler(ctx, req)
	if handler != nil {
		// TODO(#96): Consider emitting an audit log when the RPC call fails
		return handlerResp, handlerErr
	}

	// Autofill `Payload.Response`.
	if d == alpb.AuditRuleDirectiveRequestAndResponse {
		if respStruct, err := toProtoStruct(handlerResp); err != nil {
			zlogger.Warnw("unary interceptor failed converting resp into a Google struct proto", "error", err)
		} else {
			logReq.Payload.Response = respStruct
		}
	}

	// Emits the log in best-effort logging mode.
	if err := i.Log(ctx, logReq); err != nil {
		zlogger.Warnw("unary interceptor failed to emit log", "error", err)
	}

	return handlerResp, handlerErr
}

// serviceName extracts the name of a service from the string `info.FullMethod`.
// In `info.FullMethod`, the service name is preceded by one or two slashes, and
// followed by one slash. For example:
//   - /$SERVICE_NAME/foo"
//   - //$SERVICE_NAME/foo/bar
func serviceName(methodName string) (string, error) {
	re := regexp.MustCompile("^/{1,2}(.*)/")
	groups := re.FindStringSubmatch(methodName)
	if len(groups) < 2 || groups[1] == "" {
		return "", fmt.Errorf("failed capturing non-nil service name with regexp %q from %q", re.String(), methodName)
	}
	return groups[1], nil
}

type auditLogReqKey struct{}

// LogReqFromCtx returns the AuditLogRequest stored in the context.
// If the AuditLogRequest doesn't exist, we return an empty one.
func LogReqFromCtx(ctx context.Context) (*alpb.AuditLogRequest, bool) {
	if r, ok := ctx.Value(auditLogReqKey{}).(*alpb.AuditLogRequest); ok {
		return r, true
	}
	return &alpb.AuditLogRequest{Payload: &calpb.AuditLog{}}, false
}

// toProtoStruct converts v, which must marshal into a JSON object,
// into a Google Struct proto.
// This method is inspired from the Google Cloud Logging Client.
// https://github.com/googleapis/google-cloud-go/blob/main/logging/logging.go#L650
func toProtoStruct(v interface{}) (*structpb.Struct, error) {
	// Fast path: if v is already a *structpb.Struct, nothing to do.
	if s, ok := v.(*structpb.Struct); ok {
		return s, nil
	}
	// v is a Go value that supports JSON marshalling. We want a Struct
	// protobuf. Some day we may have a more direct way to get there, but right
	// now the only way is to marshal the Go value to JSON, unmarshal into a
	// map, and then build the Struct proto from the map.
	var jb []byte
	var err error
	jb, err = json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}
	var m map[string]interface{}
	err = json.Unmarshal(jb, &m)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return jsonMapToProtoStruct(m), nil
}

func jsonMapToProtoStruct(m map[string]interface{}) *structpb.Struct {
	fields := map[string]*structpb.Value{}
	for k, v := range m {
		fields[k] = jsonValueToStructValue(v)
	}
	return &structpb.Struct{Fields: fields}
}

func jsonValueToStructValue(v interface{}) *structpb.Value {
	switch x := v.(type) {
	case bool:
		return &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: x}}
	case float64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: x}}
	case string:
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: x}}
	case nil:
		return &structpb.Value{Kind: &structpb.Value_NullValue{}}
	case map[string]interface{}:
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: jsonMapToProtoStruct(x)}}
	case []interface{}:
		var vals []*structpb.Value
		for _, e := range x {
			vals = append(vals, jsonValueToStructValue(e))
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: vals}}}
	default:
		return &structpb.Value{Kind: &structpb.Value_NullValue{}}
	}
}
