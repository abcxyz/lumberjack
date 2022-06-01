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
	"errors"
	"testing"

	"github.com/abcxyz/lumberjack/clients/go/pkg/remote"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
	pkgtest "github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	capi "google.golang.org/genproto/googleapis/cloud/audit"
)

type fakeServer struct {
	api.UnimplementedAuditLogAgentServer
	gotReqs []*api.AuditLogRequest
}

func (s *fakeServer) ProcessLog(_ context.Context, logReq *api.AuditLogRequest) (*api.AuditLogResponse, error) {
	s.gotReqs = append(s.gotReqs, logReq)
	return &api.AuditLogResponse{Result: logReq}, nil
}

type fakeServerStream struct {
	grpc.ServerStream

	incomingCtx context.Context //nolint:containedctx // Only for testing
	gotRecvMsgs []interface{}
	gotSendMsgs []interface{}
}

func (ss *fakeServerStream) Context() context.Context {
	return ss.incomingCtx
}

func (ss *fakeServerStream) SendMsg(m interface{}) error {
	// Not thread safe.
	ss.gotSendMsgs = append(ss.gotSendMsgs, m)
	return nil
}

func (ss *fakeServerStream) RecvMsg(m interface{}) error {
	// Not thread safe.
	ss.gotRecvMsgs = append(ss.gotRecvMsgs, m)
	return nil
}

func TestUnaryInterceptor(t *testing.T) {
	t.Parallel()

	jwt := "Bearer " + testutil.JWTFromClaims(t, map[string]interface{}{
		"email": "user@example.com",
	})

	tests := []struct {
		name          string
		ctx           context.Context //nolint:containedctx // Only for testing
		auditRules    []*api.AuditRule
		req           interface{}
		logMode       api.AuditLogRequest_LogMode
		info          *grpc.UnaryServerInfo
		handler       grpc.UnaryHandler
		wantLogReq    *api.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "interceptor_autofills_successful_rpc",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*api.AuditRule{{
				Selector:  "/ExampleService/ExampleMethod",
				Directive: api.AuditRuleDirectiveRequestAndResponse,
				LogType:   "ADMIN_ACTIVITY",
			}},
			logMode: api.AuditLogRequest_BEST_EFFORT,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			wantLogReq: &api.AuditLogRequest{
				Type: api.AuditLogRequest_ADMIN_ACTIVITY,
				Payload: &capi.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "/ExampleService/ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &capi.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
					Request:  &structpb.Struct{},
					Response: &structpb.Struct{},
				},
				Mode: api.AuditLogRequest_BEST_EFFORT,
			},
		},
		{
			name: "interceptor_autofills_failed_rpc",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*api.AuditRule{{
				Selector:  "*",
				Directive: api.AuditRuleDirectiveRequestAndResponse,
				LogType:   "DATA_ACCESS",
			}},
			logMode: api.AuditLogRequest_BEST_EFFORT,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, grpcstatus.Error(codes.FailedPrecondition, "fake error")
			},
			wantErrSubstr: "fake error",
			wantLogReq: &api.AuditLogRequest{
				Type: api.AuditLogRequest_DATA_ACCESS,
				Payload: &capi.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "/ExampleService/ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &capi.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
					Request: &structpb.Struct{},
					Status: &rpcstatus.Status{
						Code:    int32(codes.FailedPrecondition),
						Message: "fake error",
					},
				},
				Mode: api.AuditLogRequest_BEST_EFFORT,
			},
		},
		{
			name: "interceptor_autofills_failed_rpc_unknown_err",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*api.AuditRule{{
				Selector:  "*",
				Directive: api.AuditRuleDirectiveRequestAndResponse,
				LogType:   "DATA_ACCESS",
			}},
			logMode: api.AuditLogRequest_BEST_EFFORT,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, errors.New("fake error")
			},
			wantErrSubstr: "fake error",
			wantLogReq: &api.AuditLogRequest{
				Type: api.AuditLogRequest_DATA_ACCESS,
				Payload: &capi.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "/ExampleService/ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &capi.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
					Request: &structpb.Struct{},
					Status: &rpcstatus.Status{
						Code:    int32(codes.Internal),
						Message: "fake error",
					},
				},
				Mode: api.AuditLogRequest_BEST_EFFORT,
			},
		},
		{
			name: "default_audit_rule_directive_omits_req_and_resp",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*api.AuditRule{{
				Selector:  "/ExampleService/ExampleMethod",
				Directive: api.AuditRuleDirectiveDefault,
			}},
			logMode: api.AuditLogRequest_BEST_EFFORT,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			wantLogReq: &api.AuditLogRequest{
				Payload: &capi.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "/ExampleService/ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &capi.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
				},
				Mode: api.AuditLogRequest_BEST_EFFORT,
			},
		},
		{
			name: "audit_rule_directive_omits_resp",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*api.AuditRule{{
				Selector:  "/ExampleService/ExampleMethod",
				Directive: api.AuditRuleDirectiveRequestOnly,
			}},
			logMode: api.AuditLogRequest_FAIL_CLOSE,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			wantLogReq: &api.AuditLogRequest{
				Payload: &capi.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "/ExampleService/ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &capi.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
					Request: &structpb.Struct{},
				},
				Mode: api.AuditLogRequest_FAIL_CLOSE,
			},
		},
		{
			name: "audit_rule_is_inapplicable",
			ctx:  context.Background(),
			auditRules: []*api.AuditRule{{
				Selector: "/ExampleService/Inapplicable",
			}},
			logMode: api.AuditLogRequest_BEST_EFFORT,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name: "malformed_method_info_fail_close",
			ctx:  context.Background(),
			auditRules: []*api.AuditRule{{
				Selector: "*",
			}},
			logMode: api.AuditLogRequest_FAIL_CLOSE,
			info: &grpc.UnaryServerInfo{
				FullMethod: "bananas",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
			wantErrSubstr: `audit interceptor: failed capturing non-nil service name with regexp "^/{1,2}(.*?)/" from "bananas"`,
		},
		{
			name: "malformed_method_info_best_effort",
			ctx:  context.Background(),
			auditRules: []*api.AuditRule{{
				Selector: "*",
			}},
			logMode: api.AuditLogRequest_BEST_EFFORT,
			info: &grpc.UnaryServerInfo{
				FullMethod: "bananas",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name: "unable_to_extract_principal_best_effort",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "bananas",
			})),
			auditRules: []*api.AuditRule{{
				Selector: "*",
			}},
			logMode: api.AuditLogRequest_BEST_EFFORT,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name: "unable_to_extract_principal_fail_close",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "bananas",
			})),
			auditRules: []*api.AuditRule{{
				Selector: "*",
			}},
			logMode: api.AuditLogRequest_FAIL_CLOSE,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
			wantErrSubstr: `audit interceptor failed to get request principal;`,
		},
		{
			name: "unable_to_extract_principal_fail_close",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "bananas",
			})),
			auditRules: []*api.AuditRule{{
				Selector: "*",
			}},
			logMode: api.AuditLogRequest_FAIL_CLOSE,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
			wantErrSubstr: `audit interceptor failed to get request principal;`,
		},
		{
			name: "unable_to_convert_req_to_proto_struct_fail_close",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*api.AuditRule{{
				Selector:  "*",
				Directive: api.AuditRuleDirectiveRequestAndResponse,
			}},
			logMode: api.AuditLogRequest_FAIL_CLOSE,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			req:           "bananas",
			wantErrSubstr: "audit interceptor failed converting req into a Google struct",
		},
		{
			name: "unable_to_convert_req_to_proto_struct_best_effort",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*api.AuditRule{{
				Selector:  "*",
				Directive: api.AuditRuleDirectiveRequestAndResponse,
			}},
			logMode: api.AuditLogRequest_BEST_EFFORT,
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			req: "bananas",
		},
	}
	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Interceptor{rules: tc.auditRules, logMode: tc.logMode}

			r := &fakeServer{}

			addr, _ := testutil.TestFakeGRPCServer(t, func(s *grpc.Server) {
				api.RegisterAuditLogAgentServer(s, r)
			})

			p, err := remote.NewProcessor(addr)
			if err != nil {
				t.Fatal(err)
			}
			c, err := NewClient(WithBackend(p))
			if err != nil {
				t.Fatal(err)
			}
			i.Client = c

			fromRawJWT := &security.FromRawJWT{
				FromRawJWT: []*api.FromRawJWT{{
					Key:    "authorization",
					Prefix: "Bearer ",
				}},
			}
			i.sc = fromRawJWT

			_, gotErr := i.UnaryInterceptor(tc.ctx, tc.req, tc.info, tc.handler)
			if diff := pkgtest.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("UnaryInterceptor(...) got unexpected error substring: %v", diff)
			}

			var gotReq *api.AuditLogRequest
			if len(r.gotReqs) > 0 {
				gotReq = r.gotReqs[0]
			}
			if tc.wantLogReq != nil && gotReq.Timestamp == nil {
				t.Error("UnaryInterceptor(...) gotReq missing timestamp")
			}
			if diff := cmp.Diff(tc.wantLogReq, gotReq, protocmp.Transform(), protocmp.IgnoreFields(&api.AuditLogRequest{}, "timestamp")); diff != "" {
				t.Errorf("UnaryInterceptor(...) got diff in automatically emitted LogReq (-want, +got): %v", diff)
			}
			if err := c.Stop(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestStreamInterceptor(t *testing.T) {
	t.Parallel()

	jwt := "Bearer " + testutil.JWTFromClaims(t, map[string]interface{}{
		"email": "user@example.com",
	})

	type msg struct {
		Val string
	}

	cases := []struct {
		name          string
		ss            *fakeServerStream
		info          *grpc.StreamServerInfo
		handler       grpc.StreamHandler
		auditRules    []*api.AuditRule
		wantLogReqs   []*api.AuditLogRequest
		wantErrSubstr string
	}{{
		name: "client_stream_multiple_reqs_single_resp",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*api.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: api.AuditRuleDirectiveRequestAndResponse,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			for _, m := range []*msg{{Val: "req1"}, {Val: "req2"}, {Val: "req3"}} {
				if err := ss.RecvMsg(m); err != nil {
					return err
				}
			}
			return ss.SendMsg(&msg{Val: "resp1"})
		},
		wantLogReqs: []*api.AuditLogRequest{{
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req1"),
				}},
			},
		}, {
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req2"),
				}},
			},
		}, {
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req3"),
				}},
				Response: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("resp1"),
				}},
			},
		}},
	}, {
		name: "server_stream_single_req_multiple_resps",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*api.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: api.AuditRuleDirectiveRequestAndResponse,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			if err := ss.RecvMsg(&msg{Val: "req1"}); err != nil {
				return err
			}
			for _, m := range []*msg{{Val: "resp1"}, {Val: "resp2"}, {Val: "resp3"}} {
				if err := ss.SendMsg(m); err != nil {
					return err
				}
			}
			return nil
		},
		wantLogReqs: []*api.AuditLogRequest{{
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req1"),
				}},
				Response: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("resp1"),
				}},
			},
		}, {
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Response: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("resp2"),
				}},
			},
		}, {
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Response: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("resp3"),
				}},
			},
		}},
	}, {
		name: "bidirection_stream_multiple_reqs_resps",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*api.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: api.AuditRuleDirectiveRequestAndResponse,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			if err := ss.RecvMsg(&msg{Val: "req1"}); err != nil {
				return err
			}
			if err := ss.SendMsg(&msg{Val: "resp1"}); err != nil {
				return err
			}
			if err := ss.RecvMsg(&msg{Val: "req2"}); err != nil {
				return err
			}
			if err := ss.SendMsg(&msg{Val: "resp2"}); err != nil {
				return err
			}
			return nil
		},
		wantLogReqs: []*api.AuditLogRequest{{
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req1"),
				}},
				Response: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("resp1"),
				}},
			},
		}, {
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req2"),
				}},
				Response: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("resp2"),
				}},
			},
		}},
	}, {
		name: "stream_without_logging_req_resp",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*api.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: api.AuditRuleDirectiveDefault,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			if err := ss.RecvMsg(&msg{Val: "req1"}); err != nil {
				return err
			}
			if err := ss.SendMsg(&msg{Val: "resp1"}); err != nil {
				return err
			}
			if err := ss.RecvMsg(&msg{Val: "req2"}); err != nil {
				return err
			}
			if err := ss.SendMsg(&msg{Val: "resp2"}); err != nil {
				return err
			}
			return nil
		},
		wantLogReqs: []*api.AuditLogRequest{{
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
			},
		}, {
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
			},
		}},
	}, {
		name: "stream_with_logging_req_only",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*api.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: api.AuditRuleDirectiveRequestOnly,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			if err := ss.RecvMsg(&msg{Val: "req1"}); err != nil {
				return err
			}
			if err := ss.SendMsg(&msg{Val: "resp1"}); err != nil {
				return err
			}
			if err := ss.RecvMsg(&msg{Val: "req2"}); err != nil {
				return err
			}
			if err := ss.SendMsg(&msg{Val: "resp2"}); err != nil {
				return err
			}
			return nil
		},
		wantLogReqs: []*api.AuditLogRequest{{
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req1"),
				}},
			},
		}, {
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req2"),
				}},
			},
		}},
	}, {
		name: "audit_rule_is_inapplicable",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*api.AuditRule{{
			Selector:  "/ExampleService/OtherMethod",
			Directive: api.AuditRuleDirectiveDefault,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			if err := ss.RecvMsg(&msg{Val: "req1"}); err != nil {
				return err
			}
			if err := ss.SendMsg(&msg{Val: "resp1"}); err != nil {
				return err
			}
			return nil
		},
	}, {
		name: "fail_to_retrieve_principal",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "banana",
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*api.AuditRule{{
			Selector:  "/ExampleService/OtherMethod",
			Directive: api.AuditRuleDirectiveDefault,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			if err := ss.RecvMsg(&msg{Val: "req1"}); err != nil {
				return err
			}
			if err := ss.SendMsg(&msg{Val: "resp1"}); err != nil {
				return err
			}
			return nil
		},
	}, {
		name: "handler_error",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*api.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: api.AuditRuleDirectiveRequestAndResponse,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			return grpcstatus.Error(codes.Internal, "something is wrong")
		},
		wantErrSubstr: "something is wrong",
		wantLogReqs: []*api.AuditLogRequest{{
			Type: api.AuditLogRequest_DATA_ACCESS,
			Payload: &capi.AuditLog{
				ServiceName: "ExampleService",
				MethodName:  "/ExampleService/ExampleMethod",
				AuthenticationInfo: &capi.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Status: &rpcstatus.Status{
					Code:    int32(codes.Internal),
					Message: "something is wrong",
				},
			},
		}},
	}}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Interceptor{rules: tc.auditRules}

			r := &fakeServer{}

			addr, _ := testutil.TestFakeGRPCServer(t, func(s *grpc.Server) {
				api.RegisterAuditLogAgentServer(s, r)
			})

			p, err := remote.NewProcessor(addr)
			if err != nil {
				t.Fatal(err)
			}
			c, err := NewClient(WithBackend(p))
			if err != nil {
				t.Fatal(err)
			}
			i.Client = c

			fromRawJWT := &security.FromRawJWT{
				FromRawJWT: []*api.FromRawJWT{{
					Key:    "authorization",
					Prefix: "Bearer ",
				}},
			}
			i.sc = fromRawJWT

			gotErr := i.StreamInterceptor(nil, tc.ss, tc.info, tc.handler)
			if diff := pkgtest.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("UnaryInterceptor(...) got unexpected error substring: %v", diff)
			}

			for i, lr := range r.gotReqs {
				if lr.Operation == nil || lr.Operation.Id == "" {
					t.Errorf("StreamInterceptor(...) gotReqs[%d] missing operation id", i)
				}
				if lr.Timestamp == nil {
					t.Errorf("StreamInterceptor(...) gotReqs[%d] missing timestamp", i)
				}
			}

			if diff := cmp.Diff(tc.wantLogReqs, r.gotReqs, protocmp.Transform(), protocmp.IgnoreFields(&api.AuditLogRequest{}, "timestamp", "operation")); diff != "" {
				t.Errorf("StreamInterceptor(...) got diff in automatically emitted log requests (-want, +got): %v", diff)
			}

			if err := c.Stop(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fullMethodName string
		want           string
		wantErrSubstr  string
	}{
		{
			name:           "service_name_with_one_leading_slash",
			fullMethodName: "/foo/bar",
			want:           "foo",
		},
		{
			name:           "service_name_with_two_leading_slash",
			fullMethodName: "//foo/bar",
			want:           "foo",
		},
		{
			name:           "service_name_from_string_with_many_elements",
			fullMethodName: "/foo/bar/baz",
			want:           "foo",
		},
		{
			name:           "error_due_to_nil_service_name",
			fullMethodName: "///bar",
			wantErrSubstr:  "failed capturing non-nil service name",
		},
		{
			name:           "error_due_to_missing_leading_slash",
			fullMethodName: "bar/foo",
			wantErrSubstr:  "failed capturing non-nil service name",
		},
		{
			name:           "error_due_to_empty_fullMethodName",
			fullMethodName: "",
			wantErrSubstr:  "failed capturing non-nil service name",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := serviceName(tc.fullMethodName)
			if diff := pkgtest.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("serviceName(%v) got unexpected error substring: %v", tc.fullMethodName, diff)
			}

			if got != tc.want {
				t.Errorf("serviceName(%v) = %v, want %v", tc.fullMethodName, got, tc.want)
			}
		})
	}
}

func TestHandleReturnUnary(t *testing.T) {
	t.Parallel()
	req := "test_request"
	ctx := context.Background()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		logReq, _ := LogReqFromCtx(ctx)
		logReq.Payload.ResourceName = "ExampleResourceName"
		return "test_response", nil
	}

	tests := []struct {
		name     string
		logMode  api.AuditLogRequest_LogMode
		err      error
		wantResp bool
		wantErr  bool
	}{
		{
			name:     "returns_response_no_err_best_effort",
			logMode:  api.AuditLogRequest_BEST_EFFORT,
			wantResp: true,
			wantErr:  false,
		},
		{
			name:     "returns_response_no_err_fail_close",
			logMode:  api.AuditLogRequest_FAIL_CLOSE,
			wantResp: true,
			wantErr:  false,
		},
		{
			name:     "returns_err_with_err_fail_close",
			logMode:  api.AuditLogRequest_FAIL_CLOSE,
			err:      errors.New("test error"),
			wantResp: false,
			wantErr:  true,
		},
		{
			name:     "returns_response_with_err_best_effort",
			logMode:  api.AuditLogRequest_BEST_EFFORT,
			err:      errors.New("test error"),
			wantResp: true,
			wantErr:  false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Interceptor{logMode: tc.logMode}

			got, gotErr := i.handleReturnUnary(ctx, req, handler, tc.err)

			if (gotErr != nil) != tc.wantErr {
				expected := "an error"
				if !tc.wantErr {
					expected = "nil"
				}
				t.Errorf("returned %v, but expected %v", gotErr, expected)
			}

			if (got != nil) != tc.wantResp {
				expected := "a response"
				if !tc.wantResp {
					expected = "nil"
				}
				t.Errorf("returned %v, but expected %v", got, expected)
			}
		})
	}
}

func TestHandleReturnStream(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ss := &fakeServerStream{}

	handler := func(srv interface{}, ss grpc.ServerStream) error {
		return nil
	}

	tests := []struct {
		name    string
		logMode api.AuditLogRequest_LogMode
		err     error
		wantErr bool
	}{
		{
			name:    "returns_nil_no_err_best_effort",
			logMode: api.AuditLogRequest_BEST_EFFORT,
			wantErr: false,
		},
		{
			name:    "returns_response_no_err_fail_close",
			logMode: api.AuditLogRequest_FAIL_CLOSE,
			wantErr: false,
		},
		{
			name:    "returns_err_with_err_fail_close",
			logMode: api.AuditLogRequest_FAIL_CLOSE,
			err:     errors.New("test error"),
			wantErr: true,
		},
		{
			name:    "returns_nil_with_err_best_effort",
			logMode: api.AuditLogRequest_BEST_EFFORT,
			err:     errors.New("test error"),
			wantErr: false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Interceptor{logMode: tc.logMode}

			gotErr := i.handleReturnStream(ctx, ss, handler, tc.err)

			if (gotErr != nil) != tc.wantErr {
				expected := "an error"
				if !tc.wantErr {
					expected = "nil"
				}
				t.Errorf("returned %v, but expected %v", gotErr, expected)
			}
		})
	}
}

func TestHandleReturnWithResponse(t *testing.T) {
	t.Parallel()
	response := "test_response"
	ctx := context.Background()
	errStr := "test error"
	testErr := errors.New(errStr)

	tests := []struct {
		name       string
		logMode    api.AuditLogRequest_LogMode
		err        error
		wantResp   bool
		wantErrStr string
	}{
		{
			name:     "returns_response_no_err_best_effort",
			logMode:  api.AuditLogRequest_BEST_EFFORT,
			wantResp: true,
		},
		{
			name:     "returns_response_no_err_fail_close",
			logMode:  api.AuditLogRequest_FAIL_CLOSE,
			wantResp: true,
		},
		{
			name:       "returns_err_with_err_fail_close",
			logMode:    api.AuditLogRequest_FAIL_CLOSE,
			err:        testErr,
			wantResp:   true,
			wantErrStr: errStr,
		},
		{
			name:     "returns_response_with_err_best_effort",
			logMode:  api.AuditLogRequest_BEST_EFFORT,
			err:      testErr,
			wantResp: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Interceptor{logMode: tc.logMode}

			got, gotErr := i.handleReturnWithResponse(ctx, response, tc.err)

			if diff := pkgtest.DiffErrString(gotErr, tc.wantErrStr); diff != "" {
				t.Errorf("got unexpected error substring: %v", diff)
			}

			if (got != nil) != tc.wantResp {
				expected := "a response"
				if !tc.wantResp {
					expected = "nil"
				}
				t.Errorf("returned %v, but expected %v", got, expected)
			}
		})
	}
}
