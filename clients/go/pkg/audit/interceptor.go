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
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	protostatus "google.golang.org/genproto/googleapis/rpc/status"

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

// UnaryInterceptor is a unary interceptor that automatically emits application audit logs.
func (i *Interceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fullMethodName := info.FullMethod
	mostRelevantRule := mostRelevantRule(fullMethodName, i.Rules)
	if mostRelevantRule == nil {
		return handler(ctx, req)
	}

	logReq := &alpb.AuditLogRequest{
		Payload: &calpb.AuditLog{
			AuthenticationInfo: &calpb.AuthenticationInfo{},
			Status:             &protostatus.Status{},
		},
	}

	// Autofill `Payload.ServiceName` and `Payload.MethodName`.
	s := strings.Split(fullMethodName, "/")
	if len(s) < 3 || len(s[0]) != 0 {
		return nil, fmt.Errorf("info.FullMethod should have format /$SERVICE_NAME/$METHOD_NAME")
	}
	logReq.Payload.ServiceName = s[1]
	logReq.Payload.MethodName = fullMethodName

	// Autofill `Payload.AuthenticationInfo.PrincipalEmail`.
	principal, err := i.SecurityContext.RequestPrincipal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting principal from ctx in %T: %w", i.SecurityContext, err)
	}
	logReq.Payload.AuthenticationInfo.PrincipalEmail = principal

	// Autofill `Payload.Request`.
	d := mostRelevantRule.Directive
	if d == alpb.AuditRuleDirectiveRequestAndResponse || d == alpb.AuditRuleDirectiveRequestOnly {
		reqStruct, err := toProtoStruct(req)
		if err != nil {
			return nil, fmt.Errorf("failed converting req %+v into a Google struct proto: %w", req, err)
		}
		logReq.Payload.Request = reqStruct
	}

	// Store our log req in the context to make it accessible
	// to the handler source code.
	ctx = context.WithValue(ctx, auditLogReqKey{}, logReq)

	// Execute the handler. The handler can modify the log
	// req in the context. For example, the handler can:
	//   - overwrite a log req field we set previously
	//   - fill the field `Payload.ResourceName`
	handlerResp, handlerErr := handler(ctx, req)

	// Autofill `Payload.Status` and `Payload.Response`.
	status, _ := status.FromError(handlerErr)
	logReq.Payload.Status.Code = int32(status.Code())
	logReq.Payload.Status.Message = status.Message()

	// Autofill `Payload.Response`.
	if d == alpb.AuditRuleDirectiveRequestAndResponse {
		respStruct, err := toProtoStruct(handlerResp)
		if err != nil {
			return nil, fmt.Errorf("failed converting resp %+v into a Google struct proto: %w", handlerResp, err)
		}
		logReq.Payload.Response = respStruct
	}

	// Set log type.
	logReq.Type = alpb.AuditLogRequest_UNSPECIFIED
	t, ok := alpb.AuditLogRequest_LogType_value[mostRelevantRule.LogType]
	if ok {
		logReq.Type = alpb.AuditLogRequest_LogType(t)
	}

	// Emits the log in best-effort logging mode.
	err = i.Log(ctx, logReq)
	if err != nil {
		zlogger.FromContext(ctx).Warnf("unary interceptor failed to emit log req %+v: %w", err)
	}

	return handlerResp, handlerErr
}

type auditLogReqKey struct{}

// LogReqInCtx returns the AuditLogRequest stored in the context.
// If the AuditLogRequest doesn't exist, we return an empty one.
func LogReqInCtx(ctx context.Context) *alpb.AuditLogRequest {
	r, ok := ctx.Value(auditLogReqKey{}).(*alpb.AuditLogRequest)
	if ok {
		return r
	}
	return &alpb.AuditLogRequest{
		Payload: &calpb.AuditLog{},
	}
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
