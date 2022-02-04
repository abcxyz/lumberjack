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

package cloudlogging

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"cloud.google.com/go/logging"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/cloud/audit"
	logpb "google.golang.org/genproto/googleapis/logging/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
)

type fakeServer struct {
	logpb.UnimplementedLoggingServiceV2Server

	resp      *logpb.WriteLogEntriesResponse
	returnErr error
}

func (s *fakeServer) WriteLogEntries(_ context.Context, req *logpb.WriteLogEntriesRequest) (*logpb.WriteLogEntriesResponse, error) {
	return s.resp, s.returnErr
}

func TestNewProcessor(t *testing.T) {
	t.Parallel()

	opts := []Option{WithLoggingClient(&logging.Client{})}
	p, err := NewProcessor(context.Background(), opts...)
	if err != nil {
		t.Errorf("NewProcessor(%v) unexpected error: %v", opts, err)
	}
	if p.client == nil {
		t.Errorf("NewProcessor(%v) error: p.client should not be nil", opts)
	}
	if len(p.loggerByLogType) != len(alpb.AuditLogRequest_LogType_name) {
		t.Errorf("NewProcessor(%v) got len(p.loggerByLogType)=%v, want %v", opts, len(p.loggerByLogType), len(alpb.AuditLogRequest_LogType_name))
	}

}

func TestProcessor_Process(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		server        *fakeServer
		opts          []Option
		logReq        *alpb.AuditLogRequest
		wantLogReq    *alpb.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "happy_path_with_failclose_default_should_log_successfully",
			server: &fakeServer{
				resp: &logpb.WriteLogEntriesResponse{},
			},
			logReq: &alpb.AuditLogRequest{
				Type:      alpb.AuditLogRequest_DATA_ACCESS,
				Payload:   &audit.AuditLog{ServiceName: "test-service"},
				Labels:    map[string]string{"test-key": "test-value"},
				Operation: &logpb.LogEntryOperation{Id: "test-id", Producer: "test-producer"},
			},
			wantLogReq: &alpb.AuditLogRequest{
				Type:      alpb.AuditLogRequest_DATA_ACCESS,
				Payload:   &audit.AuditLog{ServiceName: "test-service"},
				Labels:    map[string]string{"test-key": "test-value"},
				Operation: &logpb.LogEntryOperation{Id: "test-id", Producer: "test-producer"},
			},
		},
		{
			name: "injected_error_should_block_with_failclose_default",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			logReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
			},
			wantLogReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
			},
			wantErrSubstr: "synchronous write to Cloud logging failed",
		},
		{
			name: "happy_path_with_besteffort_default_should_log_successfully",
			server: &fakeServer{
				resp: &logpb.WriteLogEntriesResponse{},
			},
			opts: []Option{WithDefaultBestEffort()},
			logReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Labels:  map[string]string{"test-key": "test-value"},
			},
			wantLogReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Labels:  map[string]string{"test-key": "test-value"},
			},
		},
		{
			name: "injected_error_with_besteffort_default_should_be_ignored",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			opts: []Option{WithDefaultBestEffort()},
			logReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
			},
			wantLogReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
			},
		},
		{
			name: "explicit_failclose_should_overwrite_default_besteffort",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			opts: []Option{WithDefaultBestEffort()},
			logReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Mode:    alpb.AuditLogRequest_FAIL_CLOSE,
			},
			wantLogReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Mode:    alpb.AuditLogRequest_FAIL_CLOSE,
			},
			wantErrSubstr: "synchronous write to Cloud logging failed",
		},
		{
			name: "explicit_besteffort_should_overwrite_default_failclose",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			logReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Mode:    alpb.AuditLogRequest_BEST_EFFORT,
			},
			wantLogReq: &alpb.AuditLogRequest{
				Type:    alpb.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Mode:    alpb.AuditLogRequest_BEST_EFFORT,
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup fake Cloud Logging server.
			s := grpc.NewServer()
			defer s.Stop()
			logpb.RegisterLoggingServiceV2Server(s, tc.server)

			lis, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatalf("net.Listen(tcp, localhost:0) failed: %v", err)
			}
			go func(t *testing.T, s *grpc.Server, lis net.Listener) {
				err := s.Serve(lis)
				if err != nil {
					t.Errorf("net.Listen(tcp, localhost:0) serve failed: %v", err)
				}
			}(t, s, lis)

			addr := lis.Addr().String()
			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("dialing %q: %v", addr, err)
			}

			// Setup fake Cloud Logging client.
			ctx := context.Background()
			loggingClient, err := logging.NewClient(ctx, "testProjectID", option.WithGRPCConn(conn))
			if err != nil {
				log.Fatalf("creating client for fake at %q: %v", addr, err)
			}

			// Setup Processor.
			opts := append(tc.opts, WithLoggingClient(loggingClient))
			p, err := NewProcessor(ctx, opts...)
			if err != nil {
				log.Fatalf("calling NewProcessor: %v", err)
			}

			// Run test.
			gotErr := p.Process(ctx, tc.logReq)
			if diff := errutil.DiffSubstring(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.logReq, diff)
			}
			// Verify that the log request is not modified.
			if diff := cmp.Diff(tc.wantLogReq, tc.logReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got diff (-want, +got): %v", tc.logReq, diff)
			}
		})
	}
}

func TestProcessor_Stop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		server        *fakeServer
		opts          []Option
		logReqs       []*alpb.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "should_return_injected_error_with_besteffort",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			opts: []Option{WithDefaultBestEffort()},
			logReqs: []*alpb.AuditLogRequest{
				{
					Type:    alpb.AuditLogRequest_DATA_ACCESS,
					Payload: &audit.AuditLog{ServiceName: "test-service"},
				},
			},
			wantErrSubstr: "error flushing logs with type DATA_ACCESS",
		},
		{
			name: "should_return_two_injected_errors_with_besteffort",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			opts: []Option{WithDefaultBestEffort()},
			logReqs: []*alpb.AuditLogRequest{
				{
					Type:    alpb.AuditLogRequest_DATA_ACCESS,
					Payload: &audit.AuditLog{ServiceName: "test-service"},
				},
				{
					Type:    alpb.AuditLogRequest_ADMIN_ACTIVITY,
					Payload: &audit.AuditLog{ServiceName: "test-service"},
				},
			},
			wantErrSubstr: "error flushing logs with type ADMIN_ACTIVITY",
		},
		{
			name: "should_ignore_injected_error_with_failclose",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			logReqs: []*alpb.AuditLogRequest{
				{
					Type:    alpb.AuditLogRequest_DATA_ACCESS,
					Payload: &audit.AuditLog{ServiceName: "test-service"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup fake Cloud Logging server.
			s := grpc.NewServer()
			defer s.Stop()
			logpb.RegisterLoggingServiceV2Server(s, tc.server)

			lis, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatalf("net.Listen(tcp, localhost:0) failed: %v", err)
			}
			go func(t *testing.T, s *grpc.Server, lis net.Listener) {
				err := s.Serve(lis)
				if err != nil {
					t.Errorf("net.Listen(tcp, localhost:0) serve failed: %v", err)
				}
			}(t, s, lis)

			addr := lis.Addr().String()
			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("dialing %q: %v", addr, err)
			}

			// Setup fake Cloud Logging client.
			ctx := context.Background()
			loggingClient, err := logging.NewClient(ctx, "testProjectID", option.WithGRPCConn(conn))
			if err != nil {
				log.Fatalf("creating client for fake at %q: %v", addr, err)
			}

			// Setup Processor.
			opts := append(tc.opts, WithLoggingClient(loggingClient))
			p, err := NewProcessor(ctx, opts...)
			if err != nil {
				log.Fatalf("calling NewProcessor: %v", err)
			}

			// Write the logs.
			for _, r := range tc.logReqs {
				err := p.Process(ctx, r)
				if err != nil {
					// TODO: it may be worth validating this scenario. #47
					fmt.Printf("failed to process: %v\n", err)
				}
			}

			// Run test.
			gotErr := p.Stop()
			if diff := errutil.DiffSubstring(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Stop() got unexpected error substring: %v", diff)
			}
		})
	}
}
