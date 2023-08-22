// Copyright 2023 The Authors (see AUTHORS file)
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
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/sethvargo/go-retry"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/testing/protocmp"

	logging "cloud.google.com/go/logging/apiv2"
)

const testResource = "test-resource"

func TestPull(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		server        *fakeServer
		filter        string
		wantReq       *loggingpb.ListLogEntriesRequest
		wantResult    []*loggingpb.LogEntry
		wantErrSubstr string
	}{
		{
			name:   "success",
			filter: "test-filter",
			server: &fakeServer{
				listResp: &loggingpb.ListLogEntriesResponse{
					Entries: []*loggingpb.LogEntry{{LogName: "test"}},
				},
			},
			wantReq: &loggingpb.ListLogEntriesRequest{
				ResourceNames: []string{testResource},
				Filter:        "test-filter",
				OrderBy:       "timestamp desc",
				PageSize:      1000,
			},
			wantResult: []*loggingpb.LogEntry{{LogName: "test"}},
		},
		{
			name:   "failed_to_pull",
			filter: "",
			server: &fakeServer{
				injectedErr: fmt.Errorf("injected error"),
			},
			wantReq: &loggingpb.ListLogEntriesRequest{
				ResourceNames: []string{testResource},
				OrderBy:       "timestamp desc",
				PageSize:      1000,
			},
			wantErrSubstr: "injected error",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fakeClient := setupFakeClient(t, ctx, tc.server)
			p := NewPuller(
				ctx,
				fakeClient,
				testResource,
				WithRetry(retry.WithMaxRetries(0, retry.NewFibonacci(500*time.Millisecond))),
			)
			gotResult, gotErr := p.Pull(ctx, tc.filter, 1)
			if diff := testutil.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.name, diff)
			}

			if diff := cmp.Diff(tc.wantResult, gotResult, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got result diff (-want, +got): %v", tc.name, diff)
			}

			if diff := cmp.Diff(tc.wantReq, tc.server.listReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got request diff (-want, +got): %v", tc.name, diff)
			}
		})
	}
}

func TestStreamPull(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		server        *fakeServer
		filter        string
		wantReq       *loggingpb.TailLogEntriesRequest
		wantResult    []*loggingpb.LogEntry
		wantNum       int
		wantErrSubstr string
	}{
		{
			name:   "success",
			filter: "test-filter",
			server: &fakeServer{
				tailResp: &loggingpb.TailLogEntriesResponse{
					Entries: []*loggingpb.LogEntry{{LogName: "test1"}, {LogName: "test2"}},
				},
			},
			wantResult: []*loggingpb.LogEntry{{LogName: "test1"}, {LogName: "test2"}},
			wantReq: &loggingpb.TailLogEntriesRequest{
				ResourceNames: []string{testResource},
				Filter:        "test-filter",
			},
			wantNum: 2,
		},
		{
			name:   "failed_to_pull",
			filter: "test-filter",
			server: &fakeServer{
				injectedErr: fmt.Errorf("injected error"),
			},
			wantReq: &loggingpb.TailLogEntriesRequest{
				ResourceNames: []string{testResource},
				Filter:        "test-filter",
			},
			wantErrSubstr: "injected error",
			wantNum:       0,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ch := make(chan *loggingpb.LogEntry)
			var gotLogs []*loggingpb.LogEntry
			done := make(chan struct{}, 1)
			t.Cleanup(func() {
				close(done)
			})

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				for l := range ch {
					gotLogs = append(gotLogs, l)
					if len(gotLogs) == tc.wantNum {
						done <- struct{}{}
						break
					}
				}
			}()

			fakeClient, streamClient := setupFakeStreamClient(t, ctx, tc.server)
			p := NewPuller(
				ctx,
				fakeClient,
				testResource,
				WithRetry(retry.WithMaxRetries(0, retry.NewFibonacci(500*time.Millisecond))),
			)

			var gotPullErr error
			go func() {
				gotPullErr = p.StreamPull(ctx, tc.filter, ch, streamClient)
			}()

			select {
			case <-done:
				t.Logf("Recevied enough logEntries")
				break
			case <-time.After(5 * time.Second):
				t.Logf("No enough logEntries recevied after time out")
				break
			}
			cancel()

			if diff := cmp.Diff(tc.wantResult, gotLogs, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got result diff (-want, +got): %v", tc.name, diff)
			}

			if diff := cmp.Diff(tc.wantReq, tc.server.tailReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got request diff (-want, +got): %v", tc.name, diff)
			}

			if diff := testutil.DiffErrString(gotPullErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.name, diff)
			}
			if tc.wantNum != int(tc.server.logCounter) {
				t.Errorf("Process(%v) got different logEntry number want: %d, got: %d", tc.name, tc.wantNum, tc.server.logCounter)
			}
		})
	}
}

func setupFakeClient(t *testing.T, ctx context.Context, s *fakeServer) *logging.Client {
	t.Helper()

	// Setup fake server.
	addr, conn := testutil.FakeGRPCServer(t, func(grpcS *grpc.Server) {
		loggingpb.RegisterLoggingServiceV2Server(grpcS, s)
	})
	t.Cleanup(func() {
		conn.Close()
	})

	fakeClient, err := logging.NewClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("creating client for fake at %q: %v", addr, err)
	}
	return fakeClient
}

func setupFakeStreamClient(t *testing.T, ctx context.Context, s *fakeServer) (*logging.Client, loggingpb.LoggingServiceV2_TailLogEntriesClient) {
	t.Helper()

	// Setup fake server.
	addr, conn := testutil.FakeGRPCServer(t, func(grpcS *grpc.Server) {
		loggingpb.RegisterLoggingServiceV2Server(grpcS, s)
	})
	t.Cleanup(func() {
		conn.Close()
	})

	fakeClient, err := logging.NewClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("creating client for fake at %q: %v", addr, err)
	}
	tailLogEntriesClient, _ := fakeClient.TailLogEntries(ctx)
	return fakeClient, tailLogEntriesClient
}

type fakeServer struct {
	loggingpb.UnimplementedLoggingServiceV2Server

	listReq     *loggingpb.ListLogEntriesRequest
	listResp    *loggingpb.ListLogEntriesResponse
	tailReq     *loggingpb.TailLogEntriesRequest
	tailResp    *loggingpb.TailLogEntriesResponse
	logCounter  int64
	injectedErr error
}

func (s *fakeServer) ListLogEntries(ctx context.Context, req *loggingpb.ListLogEntriesRequest) (*loggingpb.ListLogEntriesResponse, error) {
	s.listReq = req
	return s.listResp, s.injectedErr
}

func (s *fakeServer) TailLogEntries(server loggingpb.LoggingServiceV2_TailLogEntriesServer) error {
	var err error
	s.tailReq, err = server.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive tailLogEntry request")
	}
	if s.tailResp != nil {
		// send one log entry at a time to make sure tail StreamPull can receive
		// multiple response for server.Send()
		for _, v := range s.tailResp.Entries {
			s.logCounter += 1
			if err := server.Send(&loggingpb.TailLogEntriesResponse{Entries: []*loggingpb.LogEntry{v}}); err != nil {
				return fmt.Errorf("server failed to send: %w", err)
			}
		}
	}
	return s.injectedErr
}
