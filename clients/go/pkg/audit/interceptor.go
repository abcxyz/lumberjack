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
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/genproto/googleapis/logging/v2"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/abcxyz/lumberjack/clients/go/pkg/zlogger"
	"github.com/google/uuid"
)

type auditLogReqKey struct{}

// Interceptor contains the fields required for an interceptor
// to autofill and emit audit logs.
type Interceptor struct {
	*Client
	SecurityContext security.GRPCContext
	Rules           []*alpb.AuditRule
}

// UnaryInterceptor is a gRPC unary interceptor that automatically emits application audit logs.
// The interceptor is currently implemented in fail-close mode.
// TODO(#95): add support for fail-close/best-effort logging.
func (i *Interceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	zlogger := zlogger.FromContext(ctx)
	mostRelevantRule := mostRelevantRule(info.FullMethod, i.Rules)
	if mostRelevantRule == nil {
		zlogger.Debug("no audit rule matching the method name", zap.String("method_name", info.FullMethod), zap.Any("audit_rules", i.Rules))
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
		return nil, status.Errorf(codes.FailedPrecondition, "audit interceptor: %v", err)
	}
	logReq.Payload.ServiceName = serviceName

	// Autofill `Payload.AuthenticationInfo.PrincipalEmail`.
	principal, err := i.SecurityContext.RequestPrincipal(ctx)
	if err != nil {
		zlogger.Warn("audit interceptor failed to get request principal; this is likely a result of misconfiguration of audit client (security_context)",
			zap.Any("security_context", i.SecurityContext), zap.Error(err))
		return handler(ctx, req)
	}
	logReq.Payload.AuthenticationInfo = &calpb.AuthenticationInfo{PrincipalEmail: principal}

	// Autofill `Payload.Request`.
	d := mostRelevantRule.Directive
	if d == alpb.AuditRuleDirectiveRequestAndResponse || d == alpb.AuditRuleDirectiveRequestOnly {
		if reqStruct, err := toProtoStruct(req); err != nil {
			return nil, status.Errorf(codes.Internal, "audit interceptor failed converting req into a proto struct: %v", err)
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
	if handlerErr != nil {
		// TODO(#96): Consider emitting an audit log when the RPC call fails.
		return handlerResp, handlerErr
	}

	// Autofill `Payload.Response`.
	if d == alpb.AuditRuleDirectiveRequestAndResponse {
		if respStruct, err := toProtoStruct(handlerResp); err != nil {
			return nil, status.Errorf(codes.Internal, "audit interceptor failed converting resp into a proto struct: %v", err)
		} else {
			logReq.Payload.Response = respStruct
		}
	}

	// Emits the log in best-effort logging mode.
	if err := i.Log(ctx, logReq); err != nil {
		return nil, status.Errorf(codes.Internal, "audit interceptor failed to emit log: %v", err)
	}

	return handlerResp, handlerErr
}

// StreamInterceptor intercepts gRPC stream calls to inject audit logging capability.
func (i *Interceptor) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	zlogger := zlogger.FromContext(ctx)

	r := mostRelevantRule(info.FullMethod, i.Rules)
	if r == nil {
		zlogger.Debug("no audit rule matching the method name", zap.String("method_name", info.FullMethod), zap.Any("audit_rules", i.Rules))
		return handler(srv, ss)
	}

	// Build a baseline log request to be shared by all stream calls.
	logReq := &alpb.AuditLogRequest{
		Payload: &calpb.AuditLog{},
		// Set operation to associate logs from the same stream.
		Operation: &logging.LogEntryOperation{
			Producer: info.FullMethod,
			Id:       uuid.New().String(),
		},
	}

	// Set log type.
	logReq.Type = alpb.AuditLogRequest_UNSPECIFIED
	if t, ok := alpb.AuditLogRequest_LogType_value[r.LogType]; ok {
		logReq.Type = alpb.AuditLogRequest_LogType(t)
	}

	// Autofill `Payload.ServiceName` and `Payload.MethodName`.
	logReq.Payload.MethodName = info.FullMethod
	serviceName, err := serviceName(info.FullMethod)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "audit interceptor: %v", err)
	}
	logReq.Payload.ServiceName = serviceName

	// Autofill `Payload.AuthenticationInfo.PrincipalEmail`.
	principal, err := i.SecurityContext.RequestPrincipal(ctx)
	if err != nil {
		zlogger.Warn("audit interceptor failed to get request principal; this is likely a result of misconfiguration of audit client (security_context)",
			zap.Any("security_context", i.SecurityContext), zap.Error(err))
		return handler(srv, ss)
	}
	logReq.Payload.AuthenticationInfo = &calpb.AuthenticationInfo{PrincipalEmail: principal}

	return handler(srv, &serverStreamWrapper{
		c:              i.Client,
		baselineLogReq: logReq,
		rule:           r,
		ServerStream:   ss,
	})
}

type serverStreamWrapper struct {
	grpc.ServerStream

	c *Client

	baselineLogReq *alpb.AuditLogRequest
	rule           *alpb.AuditRule

	// We use a lock to guard the last received request.
	// This is OK because according to: https://pkg.go.dev/google.golang.org/grpc#ServerStream
	// It's not safe to call RecvMsg on the same stream in different goroutines.
	// As a result, (per stream) we will only have one last received request at a time.
	mu      sync.Mutex
	lastReq interface{}
}

func (ss *serverStreamWrapper) swapLastReq(m interface{}) interface{} {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	old := ss.lastReq
	ss.lastReq = m
	return old
}

func (ss *serverStreamWrapper) popLastReq() interface{} {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	m := ss.lastReq
	ss.lastReq = nil
	return m
}

// Context attaches the audit log request to the original context.
func (ss *serverStreamWrapper) Context() context.Context {
	return context.WithValue(ss.ServerStream.Context(), auditLogReqKey{}, ss.baselineLogReq)
}

// RecvMsg wraps the original ServerStream.RecvMsg to send audit logs
// for incoming requests. We first log the last request received if any.
// We keep the latest request with the hope it can be logged in the next response.
func (ss *serverStreamWrapper) RecvMsg(m interface{}) error {
	logReq := proto.Clone(ss.baselineLogReq).(*alpb.AuditLogRequest)

	lr := ss.swapLastReq(m)
	if lr != nil {
		if ss.rule.Directive == alpb.AuditRuleDirectiveRequestAndResponse || ss.rule.Directive == alpb.AuditRuleDirectiveRequestOnly {
			if ms, err := toProtoStruct(lr); err != nil {
				return status.Errorf(codes.Internal, "audit interceptor failed converting req into a proto struct: %v", err)
			} else {
				logReq.Payload.Request = ms
				ss.c.Log(ss.ServerStream.Context(), logReq)
			}
		}
	}

	// TODO(#96): Consider emitting an audit log when the RPC call fails.
	return ss.ServerStream.RecvMsg(m)
}

func (ss *serverStreamWrapper) SendMsg(m interface{}) error {
	logReq := proto.Clone(ss.baselineLogReq).(*alpb.AuditLogRequest)

	// If there is a last request, we log it with the response in the same log entry.
	// Otherwise, this log entry will only contain the response.
	lr := ss.popLastReq()
	if lr != nil {
		if ss.rule.Directive == alpb.AuditRuleDirectiveRequestAndResponse || ss.rule.Directive == alpb.AuditRuleDirectiveRequestOnly {
			if ms, err := toProtoStruct(lr); err != nil {
				return status.Errorf(codes.Internal, "audit interceptor failed converting req into a proto struct: %v", err)
			} else {
				logReq.Payload.Request = ms
			}
		}
	}

	if ss.rule.Directive == alpb.AuditRuleDirectiveRequestAndResponse {
		if ms, err := toProtoStruct(m); err != nil {
			return status.Errorf(codes.Internal, "audit interceptor failed converting resp into a proto struct: %v", err)
		} else {
			logReq.Payload.Response = ms
		}
	}

	ss.c.Log(ss.ServerStream.Context(), logReq)

	// TODO(#96): Consider emitting an audit log when the RPC call fails.
	return ss.ServerStream.SendMsg(m)
}

var serviceNameRegexp = regexp.MustCompile("^/{1,2}(.*?)/")

// serviceName extracts the name of a service from the string `info.FullMethod`.
// In `info.FullMethod`, the service name is preceded by one or two slashes, and
// followed by one slash. For example:
//   - /$SERVICE_NAME/foo"
//   - //$SERVICE_NAME/foo/bar
func serviceName(methodName string) (string, error) {
	groups := serviceNameRegexp.FindStringSubmatch(methodName)
	if len(groups) < 2 || groups[1] == "" {
		return "", fmt.Errorf("failed capturing non-nil service name with regexp %q from %q", serviceNameRegexp.String(), methodName)
	}
	return groups[1], nil
}

// LogReqFromCtx returns the AuditLogRequest stored in the context.
// If the AuditLogRequest doesn't exist, we return an empty one.
func LogReqFromCtx(ctx context.Context) (*alpb.AuditLogRequest, bool) {
	if r, ok := ctx.Value(auditLogReqKey{}).(*alpb.AuditLogRequest); ok {
		return r, true
	}
	return &alpb.AuditLogRequest{Payload: &calpb.AuditLog{}}, false
}

// toProtoStruct converts v, which must marshal into a JSON object,
// into a proto struct.
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
