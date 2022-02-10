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
	"fmt"
	"net"
	"testing"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
	"github.com/abcxyz/lumberjack/clients/go/pkg/remote"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeServer struct {
	alpb.UnimplementedAuditLogAgentServer
	gotReqs []*alpb.AuditLogRequest
}

func (s *fakeServer) ProcessLog(_ context.Context, logReq *alpb.AuditLogRequest) (*alpb.AuditLogResponse, error) {
	s.gotReqs = append(s.gotReqs, logReq)
	return &alpb.AuditLogResponse{Result: logReq}, nil
}

type fakeServerStream struct {
	grpc.ServerStream

	incomingCtx context.Context
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
		ctx           context.Context
		auditRules    []*alpb.AuditRule
		req           interface{}
		info          *grpc.UnaryServerInfo
		handler       grpc.UnaryHandler
		wantLogReq    *alpb.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "interceptor_autofills_successful_rpc",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*alpb.AuditRule{{
				Selector:  "/ExampleService/ExampleMethod",
				Directive: alpb.AuditRuleDirectiveRequestAndResponse,
				LogType:   "ADMIN_ACTIVITY",
			}},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			wantLogReq: &alpb.AuditLogRequest{
				Type: alpb.AuditLogRequest_ADMIN_ACTIVITY,
				Payload: &calpb.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "/ExampleService/ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &calpb.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
					Request:  &structpb.Struct{},
					Response: &structpb.Struct{},
				},
			},
		},
		{
			name: "interceptor_autofills_failed_rpc",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*alpb.AuditRule{{
				Selector:  "*",
				Directive: alpb.AuditRuleDirectiveRequestAndResponse,
				LogType:   "DATA_ACCESS",
			}},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, grpcstatus.Error(codes.Internal, "fake error")
			},
			wantErrSubstr: "fake error",
		},
		{
			name: "default_audit_rule_directive_omits_req_and_resp",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*alpb.AuditRule{{
				Selector:  "/ExampleService/ExampleMethod",
				Directive: alpb.AuditRuleDirectiveDefault,
			}},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			wantLogReq: &alpb.AuditLogRequest{
				Payload: &calpb.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "/ExampleService/ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &calpb.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
				},
			},
		},
		{
			name: "audit_rule_directive_omits_resp",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*alpb.AuditRule{{
				Selector:  "/ExampleService/ExampleMethod",
				Directive: alpb.AuditRuleDirectiveRequestOnly,
			}},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			wantLogReq: &alpb.AuditLogRequest{
				Payload: &calpb.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "/ExampleService/ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &calpb.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
					Request: &structpb.Struct{},
				},
			},
		},
		{
			name: "audit_rule_is_inapplicable",
			ctx:  context.Background(),
			auditRules: []*alpb.AuditRule{{
				Selector: "/ExampleService/Inapplicable",
			}},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name: "malformed_method_info",
			ctx:  context.Background(),
			auditRules: []*alpb.AuditRule{{
				Selector: "*",
			}},
			info: &grpc.UnaryServerInfo{
				FullMethod: "bananas",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
			wantErrSubstr: `audit interceptor: failed capturing non-nil service name with regexp "^/{1,2}(.*?)/" from "bananas"`,
		},
		{
			name: "unable_to_extract_principal",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "bananas",
			})),
			auditRules: []*alpb.AuditRule{{
				Selector: "*",
			}},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name: "unable_to_convert_req_to_proto_struct",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			auditRules: []*alpb.AuditRule{{
				Selector:  "*",
				Directive: alpb.AuditRuleDirectiveRequestAndResponse,
			}},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				logReq, _ := LogReqFromCtx(ctx)
				logReq.Payload.ResourceName = "ExampleResourceName"
				return nil, nil
			},
			req:           "bananas",
			wantErrSubstr: "audit interceptor failed converting req into a proto struct",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Interceptor{Rules: tc.auditRules}

			r := &fakeServer{}
			s := grpc.NewServer()
			t.Cleanup(s.Stop)

			alpb.RegisterAuditLogAgentServer(s, r)
			lis, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatal(err)
			}
			go func(t *testing.T, s *grpc.Server, lis net.Listener) {
				err := s.Serve(lis)
				if err != nil {
					// TODO: it may be worth validating this scenario. #47
					fmt.Printf("net.Listen(tcp, localhost:0) serve failed: %v", err)
				}
			}(t, s, lis)
			p, err := remote.NewProcessor(lis.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			c, err := NewClient(WithBackend(p))
			if err != nil {
				t.Fatal(err)
			}
			i.Client = c

			fromRawJWT := &security.FromRawJWT{
				FromRawJWT: []*alpb.FromRawJWT{{
					Key:    "authorization",
					Prefix: "Bearer ",
				}},
			}
			i.SecurityContext = fromRawJWT

			_, gotErr := i.UnaryInterceptor(tc.ctx, tc.req, tc.info, tc.handler)
			if diff := errutil.DiffSubstring(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("UnaryInterceptor(...) got unexpected error substring: %v", diff)
			}

			var gotReq *alpb.AuditLogRequest
			if len(r.gotReqs) > 0 {
				gotReq = r.gotReqs[0]
			}
			if diff := cmp.Diff(tc.wantLogReq, gotReq, protocmp.Transform()); diff != "" {
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
		auditRules    []*alpb.AuditRule
		wantLogReqs   []*alpb.AuditLogRequest
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
		auditRules: []*alpb.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: alpb.AuditRuleDirectiveRequestAndResponse,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			for _, m := range []*msg{{Val: "req1"}, {Val: "req2"}, {Val: "req3"}} {
				ss.RecvMsg(m)
			}
			ss.SendMsg(&msg{Val: "resp1"})
			return nil
		},
		wantLogReqs: []*alpb.AuditLogRequest{{
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req1"),
				}},
			},
		}, {
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req2"),
				}},
			},
		}, {
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
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
		auditRules: []*alpb.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: alpb.AuditRuleDirectiveRequestAndResponse,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			ss.RecvMsg(&msg{Val: "req1"})
			for _, m := range []*msg{{Val: "resp1"}, {Val: "resp2"}, {Val: "resp3"}} {
				ss.SendMsg(m)
			}
			return nil
		},
		wantLogReqs: []*alpb.AuditLogRequest{{
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
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
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Response: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("resp2"),
				}},
			},
		}, {
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
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
		auditRules: []*alpb.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: alpb.AuditRuleDirectiveRequestAndResponse,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			ss.RecvMsg(&msg{Val: "req1"})
			ss.SendMsg(&msg{Val: "resp1"})
			ss.RecvMsg(&msg{Val: "req2"})
			ss.SendMsg(&msg{Val: "resp2"})
			return nil
		},
		wantLogReqs: []*alpb.AuditLogRequest{{
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
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
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
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
		name: "stream_wihtout_logging_req_resp",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*alpb.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: alpb.AuditRuleDirectiveDefault,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			ss.RecvMsg(&msg{Val: "req1"})
			ss.SendMsg(&msg{Val: "resp1"})
			ss.RecvMsg(&msg{Val: "req2"})
			ss.SendMsg(&msg{Val: "resp2"})
			return nil
		},
		wantLogReqs: []*alpb.AuditLogRequest{{
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
			},
		}, {
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
			},
		}},
	}, {
		name: "stream_wiht_logging_req_only",
		ss: &fakeServerStream{
			incomingCtx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
		},
		info: &grpc.StreamServerInfo{
			FullMethod: "/ExampleService/ExampleMethod",
		},
		auditRules: []*alpb.AuditRule{{
			Selector:  "/ExampleService/ExampleMethod",
			Directive: alpb.AuditRuleDirectiveRequestOnly,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			ss.RecvMsg(&msg{Val: "req1"})
			ss.SendMsg(&msg{Val: "resp1"})
			ss.RecvMsg(&msg{Val: "req2"})
			ss.SendMsg(&msg{Val: "resp2"})
			return nil
		},
		wantLogReqs: []*alpb.AuditLogRequest{{
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
				Request: &structpb.Struct{Fields: map[string]*structpb.Value{
					"Val": structpb.NewStringValue("req1"),
				}},
			},
		}, {
			Type: alpb.AuditLogRequest_DATA_ACCESS,
			Payload: &calpb.AuditLog{
				ServiceName:  "ExampleService",
				MethodName:   "/ExampleService/ExampleMethod",
				ResourceName: "ExampleResourceName",
				AuthenticationInfo: &calpb.AuthenticationInfo{
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
		auditRules: []*alpb.AuditRule{{
			Selector:  "/ExampleService/OtherMethod",
			Directive: alpb.AuditRuleDirectiveDefault,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			ss.RecvMsg(&msg{Val: "req1"})
			ss.SendMsg(&msg{Val: "resp1"})
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
		auditRules: []*alpb.AuditRule{{
			Selector:  "/ExampleService/OtherMethod",
			Directive: alpb.AuditRuleDirectiveDefault,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			logReq, _ := LogReqFromCtx(ss.Context())
			logReq.Payload.ResourceName = "ExampleResourceName"
			ss.RecvMsg(&msg{Val: "req1"})
			ss.SendMsg(&msg{Val: "resp1"})
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
		auditRules: []*alpb.AuditRule{{
			Selector:  "/ExampleService/OtherMethod",
			Directive: alpb.AuditRuleDirectiveDefault,
			LogType:   "DATA_ACCESS",
		}},
		handler: func(srv interface{}, ss grpc.ServerStream) error {
			return status.Error(codes.Internal, "something is wrong")
		},
		wantErrSubstr: "something is wrong",
	}}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Interceptor{Rules: tc.auditRules}

			r := &fakeServer{}
			s := grpc.NewServer()
			t.Cleanup(s.Stop)

			alpb.RegisterAuditLogAgentServer(s, r)
			lis, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatal(err)
			}
			go func(t *testing.T, s *grpc.Server, lis net.Listener) {
				err := s.Serve(lis)
				if err != nil {
					// TODO: it may be worth validating this scenario. #47
					fmt.Printf("net.Listen(tcp, localhost:0) serve failed: %v", err)
				}
			}(t, s, lis)
			p, err := remote.NewProcessor(lis.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			c, err := NewClient(WithBackend(p))
			if err != nil {
				t.Fatal(err)
			}
			i.Client = c

			fromRawJWT := &security.FromRawJWT{
				FromRawJWT: []*alpb.FromRawJWT{{
					Key:    "authorization",
					Prefix: "Bearer ",
				}},
			}
			i.SecurityContext = fromRawJWT

			gotErr := i.StreamInterceptor(nil, tc.ss, tc.info, tc.handler)
			if diff := errutil.DiffSubstring(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("UnaryInterceptor(...) got unexpected error substring: %v", diff)
			}

			for i, lr := range r.gotReqs {
				if lr.Operation == nil || lr.Operation.Id == "" {
					t.Errorf("StreamInterceptor(...) gotReqs[%d] missing operation id", i)
				}
				// Nil the operation for easy comparison below.
				lr.Operation = nil
			}

			if diff := cmp.Diff(tc.wantLogReqs, r.gotReqs, protocmp.Transform()); diff != "" {
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
			if diff := errutil.DiffSubstring(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("serviceName(%v) got unexpected error substring: %v", tc.fullMethodName, diff)
			}

			if got != tc.want {
				t.Errorf("serviceName(%v) = %v, want %v", tc.fullMethodName, got, tc.want)
			}
		})
	}
}
