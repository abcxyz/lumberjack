package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	protostatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	grpcmetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/zlogger"
)

// UnaryLogger is a unary interceptor that automatically emits application audit logs.
func (c Client) UnaryLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logReq := &alpb.AuditLogRequest{
		Payload: &calpb.AuditLog{
			AuthenticationInfo: &calpb.AuthenticationInfo{},
			Status:             &protostatus.Status{},
		},
	}
	// The interceptor begins by autofilling the following fields:
	//   - Payload.ServiceName
	//   - Payload.MethodName
	//   - Payload.Request
	//   - Payload.AuthenticationInfo.PrincipalEmail
	s := strings.Split(info.FullMethod, "/")
	if len(s) < 3 {
		return nil, fmt.Errorf("info.FullMethod should have format /$SERVICE_NAME/$METHOD_NAME")
	}
	logReq.Payload.ServiceName = s[1]
	logReq.Payload.MethodName = s[2]

	principal, err := principalFromContext(ctx)
	if err != nil {
		return nil, err
	}
	logReq.Payload.AuthenticationInfo.PrincipalEmail = principal

	var request *structpb.Struct
	if reqStruct, err := toProtoStruct(req); err == nil {
		request = reqStruct
	}
	logReq.Payload.Request = request

	// Store our log req in the context to make it accessible
	// to the handler source code.
	ctx = context.WithValue(ctx, auditLogReqKey{}, logReq)

	// Execute the handler. The handler can modify the log
	// req in the context. For example, the handler can:
	//   - overwrite a log req field we set previously
	//   - fill the field `Payload.ResourceName`
	resp, handlerErr := handler(ctx, req)

	// Then, the interceptor autofills the following fields:
	//   - Payload.Status
	//   - Payload.Response
	status, _ := status.FromError(handlerErr)
	logReq.Payload.Status.Code = int32(status.Code())
	logReq.Payload.Status.Message = status.Message()

	if respStruct, err := toProtoStruct(resp); err == nil {
		logReq.Payload.Response = respStruct
	}

	// Finally, the interceptor writes the log. We write the
	// error as a warning to the standard logger because we're
	// operating in best-effort logging mode, as opposed to
	// fail-closed logging mode.
	err = c.Log(ctx, logReq)
	if err != nil {
		zlogger.FromContext(ctx).Warnf("unary interceptor failed to emit log: %v", err)
	}

	return resp, handlerErr
}

// principalFromContext extracts the principal from the context.
// This method assumes that a JWT exists the grpcmetadata under
// the key `authorization` and with prefix `Bearer `. If that's
// not the case, we return an error.
func principalFromContext(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("cannot extract principal because of missing gRPC incoming context")
	}

	// Extract the JWT.
	var idToken string
	if auths := md["authorization"]; len(auths) > 0 {
		idToken = strings.TrimPrefix(auths[0], "Bearer ")
	}
	if idToken == "" {
		return "", fmt.Errorf("cannot extract principal because JWT under the key `authorization` is nil: +%v", md)
	}

	// Retrieve the principal from the JWT.
	p := &jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err := p.ParseUnverified(idToken, claims)
	if err != nil {
		return "", fmt.Errorf("unable to parse JWT: %w", err)
	}

	principal := claims["email"].(string)
	if principal == "" {
		return "", fmt.Errorf("cannot extract principal because it's nil")
	}

	return principal, nil
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
// This method is copied from the Google Cloud Logging Client.
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
	if raw, ok := v.(json.RawMessage); ok { // needed for Go 1.7 and below
		jb = []byte(raw)
	} else {
		jb, err = json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("logging: json.Marshal: %v", err)
		}
	}
	var m map[string]interface{}
	err = json.Unmarshal(jb, &m)
	if err != nil {
		return nil, fmt.Errorf("logging: json.Unmarshal: %v", err)
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
