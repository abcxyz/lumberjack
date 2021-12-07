// Copyright 2021 Lumberjack authors (see AUTHORS file)
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

package server

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
)

type fakeLogProcessor struct {
	gotReq    *alpb.AuditLogRequest
	updateReq *alpb.AuditLogRequest
	returnErr error
}

func (p *fakeLogProcessor) Process(_ context.Context, logReq *alpb.AuditLogRequest) error {
	reqClone := proto.Clone(logReq)
	p.gotReq = reqClone.(*alpb.AuditLogRequest)
	if p.updateReq != nil {
		logReq.Labels = p.updateReq.Labels
		logReq.Payload = p.updateReq.Payload
	}
	return p.returnErr
}

func TestAuditLogAgent_ProcessLog(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		req         *alpb.AuditLogRequest
		p           *fakeLogProcessor
		wantSentReq *alpb.AuditLogRequest
		wantResp    *alpb.AuditLogResponse
		wantErr     error
	}{{
		name:        "success_no_update",
		req:         testutil.ReqBuilder().Build(),
		p:           &fakeLogProcessor{},
		wantSentReq: testutil.ReqBuilder().Build(),
		wantResp: &alpb.AuditLogResponse{
			Result: testutil.ReqBuilder().Build(),
		},
	}, {
		name: "success_with_update",
		req:  testutil.ReqBuilder().WithServiceName("test-service").Build(),
		p: &fakeLogProcessor{
			updateReq: testutil.ReqBuilder().WithServiceName("bar").WithMethodName("Do").Build(),
		},
		wantSentReq: testutil.ReqBuilder().WithServiceName("test-service").Build(),
		wantResp: &alpb.AuditLogResponse{
			Result: testutil.ReqBuilder().WithServiceName("bar").WithMethodName("Do").Build(),
		},
	}, {
		name: "internal_failure",
		req:  testutil.ReqBuilder().WithServiceName("test-service").Build(),
		p: &fakeLogProcessor{
			returnErr: fmt.Errorf("injected err"),
		},
		wantSentReq: testutil.ReqBuilder().WithServiceName("test-service").Build(),
		wantErr:     status.Error(codes.Internal, "failed to execute backend *server.fakeLogProcessor: injected err"),
	}, {
		name: "invaid_argument_failure",
		req:  testutil.ReqBuilder().WithServiceName("test-service").Build(),
		p: &fakeLogProcessor{
			returnErr: fmt.Errorf("injected: %w", audit.ErrInvalidRequest),
		},
		wantSentReq: testutil.ReqBuilder().WithServiceName("test-service").Build(),
		wantErr:     status.Error(codes.InvalidArgument, "failed to execute backend *server.fakeLogProcessor: injected: invalid audit log request"),
	}}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := grpc.NewServer()
			defer s.Stop()

			ac, err := audit.NewClient(audit.WithBackend(tc.p))
			if err != nil {
				t.Fatalf("Failed to create audit client: %v", err)
			}
			server, err := NewAuditLogAgent(ac)
			if err != nil {
				t.Fatalf("Failed to create audit log agent server: %v", err)
			}
			alpb.RegisterAuditLogAgentServer(s, server)

			lis, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatalf("net.Listen(tcp, localhost:0) failed: %v", err)
			}
			go s.Serve(lis)

			conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
			if err != nil {
				t.Fatalf("Failed to establish gRPC conn: %v", err)
			}

			client := alpb.NewAuditLogAgentClient(conn)
			gotResp, err := client.ProcessLog(context.Background(), tc.req)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ProcessLog() error (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantResp, gotResp, protocmp.Transform()); diff != "" {
				t.Errorf("ProcessLog() response (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantSentReq, tc.p.gotReq, protocmp.Transform()); diff != "" {
				t.Errorf("ProcessLog() request sent (-want,+got):\n%s", diff)
			}
		})
	}
}
