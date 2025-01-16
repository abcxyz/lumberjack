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
	"testing"

	"cloud.google.com/go/logging"
	logpb "cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
	pkgtestutil "github.com/abcxyz/pkg/testutil"
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
	if len(p.loggerByLogType) != len(api.AuditLogRequest_LogType_name) {
		t.Errorf("NewProcessor(%v) got len(p.loggerByLogType)=%v, want %v", opts, len(p.loggerByLogType), len(api.AuditLogRequest_LogType_name))
	}
}

func TestProcessor_Process(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		server        *fakeServer
		opts          []Option
		logReq        *api.AuditLogRequest
		wantLogReq    *api.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "happy_path_with_failclose_default_should_log_successfully",
			server: &fakeServer{
				resp: &logpb.WriteLogEntriesResponse{},
			},
			logReq: &api.AuditLogRequest{
				Type:      api.AuditLogRequest_DATA_ACCESS,
				Payload:   &audit.AuditLog{ServiceName: "test-service"},
				Labels:    map[string]string{"test-key": "test-value"},
				Operation: &logpb.LogEntryOperation{Id: "test-id", Producer: "test-producer"},
			},
			wantLogReq: &api.AuditLogRequest{
				Type:      api.AuditLogRequest_DATA_ACCESS,
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
			logReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
			},
			wantLogReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
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
			logReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Labels:  map[string]string{"test-key": "test-value"},
			},
			wantLogReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
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
			logReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
			},
			wantLogReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
			},
		},
		{
			name: "explicit_failclose_should_overwrite_default_besteffort",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			opts: []Option{WithDefaultBestEffort()},
			logReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Mode:    api.AuditLogRequest_FAIL_CLOSE,
			},
			wantLogReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Mode:    api.AuditLogRequest_FAIL_CLOSE,
			},
			wantErrSubstr: "synchronous write to Cloud logging failed",
		},
		{
			name: "explicit_besteffort_should_overwrite_default_failclose",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			logReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Mode:    api.AuditLogRequest_BEST_EFFORT,
			},
			wantLogReq: &api.AuditLogRequest{
				Type:    api.AuditLogRequest_DATA_ACCESS,
				Payload: &audit.AuditLog{ServiceName: "test-service"},
				Mode:    api.AuditLogRequest_BEST_EFFORT,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup fake Cloud Logging server.
			addr, conn := testutil.TestFakeGRPCServer(t, func(s *grpc.Server) {
				logpb.RegisterLoggingServiceV2Server(s, tc.server)
			})

			// Setup fake Cloud Logging client.
			ctx := context.Background()
			loggingClient, err := logging.NewClient(ctx, "testProjectID", option.WithGRPCConn(conn))
			if err != nil {
				t.Fatalf("creating client for fake at %q: %v", addr, err)
			}

			// Setup Processor.
			opts := append(tc.opts, WithLoggingClient(loggingClient))
			p, err := NewProcessor(ctx, opts...)
			if err != nil {
				t.Fatalf("calling NewProcessor: %v", err)
			}

			// Run test.
			gotErr := p.Process(ctx, tc.logReq)
			if diff := pkgtestutil.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
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
		logReqs       []*api.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "should_return_injected_error_with_besteffort",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			opts: []Option{WithDefaultBestEffort()},
			logReqs: []*api.AuditLogRequest{
				{
					Type:    api.AuditLogRequest_DATA_ACCESS,
					Payload: &audit.AuditLog{ServiceName: "test-service"},
				},
			},
			wantErrSubstr: "failed to flush DATA_ACCESS logs",
		},
		{
			name: "should_return_two_injected_errors_with_besteffort",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			opts: []Option{WithDefaultBestEffort()},
			logReqs: []*api.AuditLogRequest{
				{
					Type:    api.AuditLogRequest_DATA_ACCESS,
					Payload: &audit.AuditLog{ServiceName: "test-service"},
				},
				{
					Type:    api.AuditLogRequest_ADMIN_ACTIVITY,
					Payload: &audit.AuditLog{ServiceName: "test-service"},
				},
			},
			wantErrSubstr: "failed to flush ADMIN_ACTIVITY logs",
		},
		{
			name: "should_ignore_injected_error_with_failclose",
			server: &fakeServer{
				returnErr: status.Error(codes.FailedPrecondition, "injected err"),
			},
			logReqs: []*api.AuditLogRequest{
				{
					Type:    api.AuditLogRequest_DATA_ACCESS,
					Payload: &audit.AuditLog{ServiceName: "test-service"},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup fake Cloud Logging server.
			addr, conn := testutil.TestFakeGRPCServer(t, func(s *grpc.Server) {
				logpb.RegisterLoggingServiceV2Server(s, tc.server)
			})

			// Setup fake Cloud Logging client.
			ctx := context.Background()
			loggingClient, err := logging.NewClient(ctx, "testProjectID", option.WithGRPCConn(conn))
			if err != nil {
				t.Fatalf("creating client for fake at %q: %v", addr, err)
			}

			// Setup Processor.
			opts := append(tc.opts, WithLoggingClient(loggingClient))
			p, err := NewProcessor(ctx, opts...)
			if err != nil {
				t.Fatalf("calling NewProcessor: %v", err)
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
			if diff := pkgtestutil.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Stop() got unexpected error substring: %v", diff)
			}
		})
	}
}
