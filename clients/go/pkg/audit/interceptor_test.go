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

	"github.com/google/go-cmp/cmp"
	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	protostatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
	"github.com/abcxyz/lumberjack/clients/go/pkg/remote"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
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

	// Test JWT:
	// {
	// 	 "name": "user",
	// 	 "email": "user@example.com"
	// }
	jwt := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6InVzZXIiLCJpYXQiOjE1MTYyMzkwMjIsImVtYWlsIjoidXNlckBleGFtcGxlLmNvbSJ9.PXl-SJniWHMVLNYb77HmVFFqWTlu28xf9fou2GaT0Jc"

	tests := []struct {
		name          string
		ctx           context.Context
		req           interface{}
		info          *grpc.UnaryServerInfo
		handler       grpc.UnaryHandler
		wantLogReq    *alpb.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "interceptor_autofills_everything",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				LogReqInCtx(ctx).Payload.ResourceName = "ExampleResourceName"
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
					Request:  &structpb.Struct{},
					Response: &structpb.Struct{},
					Status:   &protostatus.Status{},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Interceptor{}

			r := &fakeServer{}
			s := grpc.NewServer()
			defer s.Stop()
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
			defer c.Stop()
			i.Client = c

			fromRawJWT := &security.FromRawJWT{
				FromRawJWT: &alpb.FromRawJWT{
					Key:    "authorization",
					Prefix: "Bearer ",
				},
			}
			i.SecurityContext = fromRawJWT

			_, gotErr := i.UnaryInterceptor(tc.ctx, tc.req, tc.info, tc.handler)
			if diff := errutil.DiffSubstring(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("UnaryInterceptor(...) got unexpected error substring: %v", diff)
			}

			if diff := cmp.Diff(tc.wantLogReq, r.gotReq, protocmp.Transform()); diff != "" {
				t.Errorf("UnaryInterceptor(...) got diff in automatically emitted LogReq (-want, +got): %v", diff)
			}
		})
	}
}
