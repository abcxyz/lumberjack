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
	"time"

	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	zlogger "github.com/abcxyz/pkg/logging"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/logging/v2"
	rpccode "google.golang.org/genproto/googleapis/rpc/code"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcmetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	capi "google.golang.org/genproto/googleapis/cloud/audit"
)

type auditLogReqKey struct{}

// InterceptorOption defines the option func to configure an interceptor.
type InterceptorOption func(i *Interceptor) error

// WithAuditClient configures the interceptor to use the given audit client
// to send audit logs.
func WithAuditClient(c *Client) InterceptorOption {
	return func(i *Interceptor) error {
		i.Client = c
		return nil
	}
}

// WithSecurityContext configures the interceptor to use the given security
// context to retrieve authentication info.
func WithSecurityContext(sc security.GRPCContext) InterceptorOption {
	return func(i *Interceptor) error {
		i.sc = sc
		return nil
	}
}

// WithAuditRules configures the interceptor to use the given rules to match
// methods and instruct audit logging.
func WithAuditRules(rs ...*api.AuditRule) InterceptorOption {
	return func(i *Interceptor) error {
		i.rules = rs
		return nil
	}
}

// WithInterceptorLogMode configures the interceptor to honor the given log mode.
func WithInterceptorLogMode(m api.AuditLogRequest_LogMode) InterceptorOption {
	return func(i *Interceptor) error {
		i.logMode = m
		return nil
	}
}

// WithJWTValidator configures the interceptor to use the JWTValidator client to add justifications to logs.
func WithJWTValidator(j JWTValidator) InterceptorOption {
	return func(i *Interceptor) error {
		i.jwtValidator = j
		return nil
	}
}

// JWTValidator is intended to provide a method for validating JWTs passed in using the
// justification_token header.
type JWTValidator interface {
	ValidateJWT(jwtStr string) (*jwt.Token, error)
}

// Interceptor contains the fields required for an interceptor
// to autofill and emit audit logs.
type Interceptor struct {
	*Client
	sc           security.GRPCContext
	rules        []*api.AuditRule
	logMode      api.AuditLogRequest_LogMode
	jwtValidator JWTValidator
}

// NewInterceptor creates a new interceptor with the given options.
func NewInterceptor(options ...InterceptorOption) (*Interceptor, error) {
	it := &Interceptor{}
	for _, o := range options {
		if err := o(it); err != nil {
			return nil, fmt.Errorf("failed to apply interceptor option: %w", err)
		}
	}
	return it, nil
}

// UnaryInterceptor is a gRPC unary interceptor that automatically emits application audit logs.
// The interceptor is currently implemented in fail-close mode.
func (i *Interceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger := zlogger.FromContext(ctx)
	r := mostRelevantRule(info.FullMethod, i.rules)
	if r == nil {
		logger.Debug("no audit rule matching the method name", zap.String("method_name", info.FullMethod), zap.Any("audit_rules", i.rules))
		// Interceptor not applied to this method, continue
		return handler(ctx, req)
	}

	serviceName, err := serviceName(info.FullMethod)
	if err != nil {
		return i.handleReturnUnary(ctx, req, handler, status.Errorf(codes.FailedPrecondition, "audit interceptor: %v", err))
	}

	logReq := &api.AuditLogRequest{
		Payload: &capi.AuditLog{
			ServiceName: serviceName,
			MethodName:  info.FullMethod,
		},
		Mode:      i.logMode,
		Timestamp: timestamppb.New(time.Now().UTC()),
	}

	// Set log type.
	logReq.Type = api.AuditLogRequest_UNSPECIFIED
	if t, ok := api.AuditLogRequest_LogType_value[r.LogType]; ok {
		logReq.Type = api.AuditLogRequest_LogType(t)
	}

	// Autofill `Payload.AuthenticationInfo.PrincipalEmail`.
	principal, err := i.sc.RequestPrincipal(ctx)
	if err != nil {
		return i.handleReturnUnary(ctx, req, handler, status.Errorf(codes.FailedPrecondition,
			"audit interceptor failed to get request principal; this is likely a result of misconfiguration of audit client (security_context): %v %v",
			zap.Any("security_context", i.sc), zap.Error(err)))
	}
	logReq.Payload.AuthenticationInfo = &capi.AuthenticationInfo{PrincipalEmail: principal}

	// Autofill `Payload.Request`.
	if shouldLogReq(r) {
		if err := setReq(logReq, req); err != nil {
			return i.handleReturnUnary(ctx, req, handler, status.Errorf(codes.Internal,
				"audit interceptor failed converting req into a Google struct proto: %v", err))
		}
	}

	// If JWT Validator hasn't been added, we don't handle justification JWTs, and just log
	// without the justification.
	if i.jwtValidator != nil {
		// JWT Validator added, look for justification info
		md, ok := grpcmetadata.FromIncomingContext(ctx)
		if !ok {
			return "", fmt.Errorf("gRPC metadata in incoming context is missing")
		}
		vals := md.Get("justification_token")
		if len(vals) == 0 {
			return i.handleReturnUnary(ctx, req, handler, status.Errorf(codes.Internal,
				"no justification token found."))
		}
		jwtRaw := vals[0]
		tok, err := i.jwtValidator.ValidateJWT(jwtRaw)
		if err != nil {
			return i.handleReturnUnary(ctx, req, handler, status.Errorf(codes.Internal,
				"audit interceptor failed converting parsing or validating justification token: %v", err))
		}
		buf, err := json.Marshal(*tok)
		if err != nil {
			return i.handleReturnUnary(ctx, req, handler, status.Errorf(codes.Internal,
				"couldn't convert token to json: %v", err))
		}
		// Note: We don't set metadata before here, so we can directly set it.
		// If we need to put other data in metadata, this should be modified to not
		// overwrite it.
		logReq.Payload.Metadata = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"justification_token": structpb.NewStringValue(string(buf)),
			},
		}
	}

	// Store our log req in the context to make it accessible
	// to the handler source code.
	ctx = context.WithValue(ctx, auditLogReqKey{}, logReq)

	// Execute the handler. The handler can modify the log
	// req in the context. For example, the handler can:
	//   - overwrite a log req field we set previously
	//   - fill the field `Payload.ResourceName`
	resp, handlerErr := handler(ctx, req)
	if handlerErr != nil {
		i.setErrorStatus(handlerErr, logReq)

		// Best effort log the error.
		if err := i.Log(ctx, logReq); err != nil {
			logger.Error("unable to audit log error", zap.Error(err))
		}
		return resp, handlerErr
	}

	// Autofill `Payload.Response`.
	if shouldLogResp(r) {
		if err := setResp(logReq, resp); err != nil {
			return i.handleReturnWithResponse(ctx, resp, status.Errorf(codes.Internal,
				"audit interceptor failed converting resp into a Google struct proto: %v", err))
		}
	}

	if err := i.Log(ctx, logReq); err != nil {
		return i.handleReturnWithResponse(ctx, resp, status.Errorf(codes.Internal, "audit interceptor failed to emit log: %v", err))
	}

	return resp, handlerErr
}

// StreamInterceptor intercepts gRPC stream calls to inject audit logging capability.
func (i *Interceptor) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	logger := zlogger.FromContext(ctx)

	r := mostRelevantRule(info.FullMethod, i.rules)
	if r == nil {
		logger.Debug("no audit rule matching the method name", zap.String("method_name", info.FullMethod), zap.Any("audit_rules", i.rules))
		return handler(srv, ss)
	}

	serviceName, err := serviceName(info.FullMethod)
	if err != nil {
		return i.handleReturnStream(ctx, ss, handler, status.Errorf(codes.FailedPrecondition, "audit interceptor: %v", err))
	}

	// Build a baseline log request to be shared by all stream calls.
	logReq := &api.AuditLogRequest{
		Payload: &capi.AuditLog{
			ServiceName: serviceName,
			MethodName:  info.FullMethod,
		},
		// Set operation to associate logs from the same stream.
		Operation: &logging.LogEntryOperation{
			Producer: info.FullMethod,
			Id:       uuid.New().String(),
		},
		Timestamp: timestamppb.New(time.Now().UTC()),
	}

	// Set log type.
	logReq.Type = api.AuditLogRequest_UNSPECIFIED
	if t, ok := api.AuditLogRequest_LogType_value[r.LogType]; ok {
		logReq.Type = api.AuditLogRequest_LogType(t)
	}

	// Autofill `Payload.AuthenticationInfo.PrincipalEmail`.
	principal, err := i.sc.RequestPrincipal(ctx)
	if err != nil {
		return i.handleReturnStream(ctx, ss, handler, status.Errorf(codes.FailedPrecondition,
			"audit interceptor failed to get request principal; this is likely a result of misconfiguration of audit client (security_context): %v %v",
			zap.Any("security_context", i.sc), zap.Error(err)))
	}
	logReq.Payload.AuthenticationInfo = &capi.AuthenticationInfo{PrincipalEmail: principal}

	handlerErr := handler(srv, &serverStreamWrapper{
		c:              i.Client,
		baselineLogReq: logReq,
		rule:           r,
		ServerStream:   ss,
	})
	if handlerErr != nil {
		i.setErrorStatus(handlerErr, logReq)

		// Best effort log the error.
		if err := i.Log(ctx, logReq); err != nil {
			logger.Error("unable to audit log error", zap.Error(err))
		}
	}
	return handlerErr
}

type serverStreamWrapper struct {
	grpc.ServerStream

	c *Client

	baselineLogReq *api.AuditLogRequest
	rule           *api.AuditRule

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
	logReq, ok := proto.Clone(ss.baselineLogReq).(*api.AuditLogRequest)
	if !ok {
		return fmt.Errorf("expected *api.AuditLogRequest")
	}

	// RecvMsg is a blocking call until the next message is received into 'm'.
	if err := ss.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	lr := ss.swapLastReq(m)
	if lr != nil {
		if shouldLogReq(ss.rule) {
			if err := setReq(logReq, lr); err != nil {
				return err
			}
			if err := ss.c.Log(ss.ServerStream.Context(), logReq); err != nil {
				return status.Errorf(codes.Internal, "audit interceptor failed to emit log: %v", err)
			}
		}
	}

	return nil
}

// SendMsg wraps the original ServerStream.SendMsg to send audit logs
// for outgoing responses. If there is a request from last time, we log them
// together. Otherwise, only the response will be logged.
func (ss *serverStreamWrapper) SendMsg(m interface{}) error {
	logReq, ok := proto.Clone(ss.baselineLogReq).(*api.AuditLogRequest)
	if !ok {
		return fmt.Errorf("expected *api.AuditLogRequest")
	}

	// If there is a last request, we log it with the response in the same log entry.
	// Otherwise, this log entry will only contain the response.
	lr := ss.popLastReq()
	if lr != nil {
		if shouldLogReq(ss.rule) {
			if err := setReq(logReq, lr); err != nil {
				return err
			}
		}
	}

	if shouldLogResp(ss.rule) {
		if err := setResp(logReq, m); err != nil {
			return err
		}
	}

	if err := ss.c.Log(ss.ServerStream.Context(), logReq); err != nil {
		return status.Errorf(codes.Internal, "audit interceptor failed to emit log: %v", err)
	}

	return ss.ServerStream.SendMsg(m)
}

func setReq(logReq *api.AuditLogRequest, m interface{}) error {
	ms, err := toProtoStruct(m)
	if err != nil {
		return status.Errorf(codes.Internal, "audit interceptor failed converting req into a proto struct: %v", err)
	}
	logReq.Payload.Request = ms
	return nil
}

func setResp(logReq *api.AuditLogRequest, m interface{}) error {
	ms, err := toProtoStruct(m)
	if err != nil {
		return status.Errorf(codes.Internal, "audit interceptor failed converting resp into a proto struct: %v", err)
	}
	logReq.Payload.Response = ms
	return nil
}

func shouldLogReq(r *api.AuditRule) bool {
	return r.Directive == api.AuditRuleDirectiveRequestAndResponse || r.Directive == api.AuditRuleDirectiveRequestOnly
}

func shouldLogResp(r *api.AuditRule) bool {
	return r.Directive == api.AuditRuleDirectiveRequestAndResponse
}

// handleReturnUnary is intended to be a wrapper that handles the LogMode correctly, and returns errors or the handler
// depending on whether the config and has specified to fail close.
func (i *Interceptor) handleReturnUnary(ctx context.Context, req interface{}, handler grpc.UnaryHandler, err error) (interface{}, error) {
	if api.ShouldFailClose(i.logMode) && err != nil {
		return nil, err
	}
	if err != nil {
		// There was an error, but we are failing open.
		logger := zlogger.FromContext(ctx)
		logger.Warn("Error occurred while attempting to audit log, but continuing without audit logging or raising error.",
			zap.Error(err))
	}
	return handler(ctx, req)
}

func (i *Interceptor) handleReturnStream(ctx context.Context, ss grpc.ServerStream, handler grpc.StreamHandler, err error) error {
	if api.ShouldFailClose(i.logMode) && err != nil {
		return err
	}
	if err != nil {
		// There was an error, but we are failing open.
		logger := zlogger.FromContext(ctx)
		logger.Warn("Error occurred while attempting to audit log, but continuing without audit logging or raising error.",
			zap.Error(err))
	}
	return handler(ctx, ss)
}

// handleReturnWithResponse is intended to be a wrapper that handles the LogMode correctly, and returns errors or a response
// depending on whether the config and has specified to fail close. Differs from the above, as this is intended to be used
// after the next handler in the chain has returned, and so we have a response formed already.
func (i *Interceptor) handleReturnWithResponse(ctx context.Context, handlerResp interface{}, err error) (interface{}, error) {
	if api.ShouldFailClose(i.logMode) && err != nil {
		return handlerResp, err
	}
	if err != nil {
		// There was an error, but we are failing open.
		logger := zlogger.FromContext(ctx)
		logger.Warn("Error occurred while attempting to audit log, but continuing without audit logging or raising error.",
			zap.Error(err))
	}
	return handlerResp, nil
}

// logError attempts to emit an audit log for an error that has occurred. Errors are logged in
// rpc Status format, and if a grpc error has occurred, that grpc error is converted to rpc.
func (i *Interceptor) setErrorStatus(err error, logReq *api.AuditLogRequest) {
	grpcStatus, ok := status.FromError(err)
	if ok {
		logReq.Payload.Status = &rpcstatus.Status{
			Code:    int32(grpcStatus.Code()),
			Message: grpcStatus.Message(),
		}
	} else {
		logReq.Payload.Status = &rpcstatus.Status{
			Code:    int32(rpccode.Code_INTERNAL),
			Message: err.Error(),
		}
	}
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
func LogReqFromCtx(ctx context.Context) (*api.AuditLogRequest, bool) {
	if r, ok := ctx.Value(auditLogReqKey{}).(*api.AuditLogRequest); ok {
		return r, true
	}
	return &api.AuditLogRequest{Payload: &capi.AuditLog{}}, false
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
