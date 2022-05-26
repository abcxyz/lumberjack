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

package remote

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
	"google.golang.org/protobuf/testing/protocmp"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/internal/testutil"
)

type fakeServer struct {
	api.UnimplementedAuditLogAgentServer

	gotReq    *api.AuditLogRequest
	resp      *api.AuditLogResponse
	returnErr error
}

func (s *fakeServer) ProcessLog(_ context.Context, logReq *api.AuditLogRequest) (*api.AuditLogResponse, error) {
	s.gotReq = logReq
	return s.resp, s.returnErr
}

func TestProcessor_Process_Insecure(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		server        *fakeServer
		req           *api.AuditLogRequest
		wantSentReq   *api.AuditLogRequest
		wantResultReq *api.AuditLogRequest
		wantErr       error
	}{{
		name: "success_log_req_no_change",
		server: &fakeServer{
			resp: &api.AuditLogResponse{
				Result: testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
			},
		},
		req:           testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
		wantSentReq:   testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
		wantResultReq: testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
	}, {
		name: "success_log_req_updated",
		server: &fakeServer{
			resp: &api.AuditLogResponse{
				Result: testutil.NewRequest(testutil.WithLabels(
					map[string]string{
						"foo": "bar",
						"abc": "123",
					})),
			},
		},
		req:         testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
		wantSentReq: testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
		wantResultReq: testutil.NewRequest(testutil.WithLabels(
			map[string]string{
				"foo": "bar",
				"abc": "123",
			})),
	}, {
		name: "server_error",
		server: &fakeServer{
			returnErr: status.Error(codes.FailedPrecondition, "injected err"),
		},
		req:           testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
		wantSentReq:   testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
		wantResultReq: testutil.NewRequest(testutil.WithLabels(map[string]string{"foo": "bar"})),
		wantErr:       status.Error(codes.FailedPrecondition, "injected err"),
	}}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := grpc.NewServer()
			defer s.Stop()
			api.RegisterAuditLogAgentServer(s, tc.server)

			lis, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatalf("net.Listen(tcp, localhost:0) failed: %v", err)
			}
			go func() {
				err := s.Serve(lis)
				if err != nil {
					// TODO: it may be worth validating this scenario. #47
					fmt.Printf("net.Listen(tcp, localhost:0) serve failed: %v\n", err)
				}
			}()

			addr := lis.Addr().String()
			p, err := NewProcessor(addr)
			if err != nil {
				t.Fatalf("NewProcessor() failed: %v", err)
			}
			t.Cleanup(func() {
				err := p.Stop()
				if err != nil {
					t.Errorf("failed to stop processor: %v", err)
				}
			})

			gotErr := p.Process(context.Background(), tc.req)
			if !cmp.Equal(tc.wantErr, gotErr, cmpopts.EquateErrors()) {
				t.Errorf("Process() error got=%v, want=%v", gotErr, tc.wantErr)
			}
			if diff := cmp.Diff(tc.wantResultReq, tc.req, protocmp.Transform()); diff != "" {
				t.Errorf("Process() request (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantSentReq, tc.server.gotReq, protocmp.Transform()); diff != "" {
				t.Errorf("Server received request (-want,+got):\n%s", diff)
			}
		})
	}
}
