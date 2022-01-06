package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt"
	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	grpcmetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

func (c Client) AuditInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	principal, err := principalFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, auditLogReqKey{}, &alpb.AuditLogRequest{
		Payload: &calpb.AuditLog{
			MethodName: info.FullMethod,
			AuthenticationInfo: &calpb.AuthenticationInfo{
				PrincipalEmail: principal,
			},
			Status: &spb.Status{},
		},
	})

	if reqStruct, err := toProtoStruct(req); err == nil {
		LogReqInCtx(ctx).Payload.Request = reqStruct
	}

	resp, handlerErr := handler(ctx, req)
	s, _ := status.FromError(handlerErr) // tested and working for both 200 and X00 responses
	LogReqInCtx(ctx).Payload.Status.Code = int32(s.Code())
	LogReqInCtx(ctx).Payload.Status.Message = s.Message()

	if respStruct, err := toProtoStruct(resp); err == nil {
		LogReqInCtx(ctx).Payload.Response = respStruct
	}

	c.Log(ctx, LogReqInCtx(ctx))

	return resp, handlerErr
}

func principalFromContext(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing gRPC incoming context")
	}
	// log.Printf("incomingContext metadata: \n" + prettyPrint(md))

	var principal string
	// Try to get the principal from Google ID JWT
	var idToken string
	if auths := md["authorization"]; len(auths) > 0 {
		idToken = auths[0][7:] // trim "Bearer: " prefix
	}
	if idToken != "" {
		// Retrieve the identity
		p := &jwt.Parser{}
		claims := jwt.MapClaims{}
		_, _, err := p.ParseUnverified(idToken, claims)
		if err != nil {
			return "", fmt.Errorf("invalid auth token: %w", err)
		}
		principal = claims["email"].(string)
	}
	// Try to get the principal from IAP header
	if emails := md["x-goog-authenticated-user-email"]; len(emails) > 0 {
		principal = emails[0]
	}
	if principal == "" {
		return "", fmt.Errorf("id token and IAP principal are nil in authorization metadata: %+v", md)
	}
	return principal, nil
}

type auditLogReqKey struct{}

func LogReqInCtx(ctx context.Context) *alpb.AuditLogRequest {
	r, ok := ctx.Value(auditLogReqKey{}).(*alpb.AuditLogRequest)
	if ok {
		return r
	}
	return &alpb.AuditLogRequest{
		Payload: &calpb.AuditLog{},
	}
}

// Copied from Log client.

// toProtoStruct converts v, which must marshal into a JSON object,
// into a Google Struct proto.
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
