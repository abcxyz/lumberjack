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
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeServer struct {
	alpb.UnimplementedAuditLogAgentServer
	gotReq *alpb.AuditLogRequest
}

func (s *fakeServer) ProcessLog(_ context.Context, logReq *alpb.AuditLogRequest) (*alpb.AuditLogResponse, error) {
	s.gotReq = logReq
	return &alpb.AuditLogResponse{Result: logReq}, nil
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
			wantErrSubstr: "audit interceptor failed converting req into a Google struct proto",
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

			if diff := cmp.Diff(tc.wantLogReq, r.gotReq, protocmp.Transform()); diff != "" {
				t.Errorf("UnaryInterceptor(...) got diff in automatically emitted LogReq (-want, +got): %v", diff)
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
