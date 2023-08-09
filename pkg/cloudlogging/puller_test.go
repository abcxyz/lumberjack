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

func TestPull(t *testing.T) {
	t.Parallel()

	filter := "test-filter"
	resource := "test-resource"

	cases := []struct {
		name          string
		server        *fakeServer
		wantReq       *loggingpb.ListLogEntriesRequest
		wantResult    *loggingpb.LogEntry
		wantErrSubstr string
	}{
		{
			name: "success",
			server: &fakeServer{
				resp: &loggingpb.ListLogEntriesResponse{
					Entries: []*loggingpb.LogEntry{{LogName: "test"}},
				},
			},
			wantReq:    &loggingpb.ListLogEntriesRequest{
				ResourceNames: []string{resource},
				Filter: filter,
				OrderBy: "timestamp desc",
				PageSize: 1000,
			},
			wantResult: &loggingpb.LogEntry{LogName: "test"},
		},
		{
			name: "failed_to_pull",
			server: &fakeServer{
				injectedErr: fmt.Errorf("injected error"),
			},
			wantReq:    &loggingpb.ListLogEntriesRequest{
				ResourceNames: []string{resource},
				Filter: filter,
				OrderBy: "timestamp desc",
				PageSize: 1000,
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
				resource,
				WithRetry(retry.WithMaxRetries(0, retry.NewFibonacci(500*time.Millisecond))),
			)
			gotResult := make(chan *loggingpb.LogEntry)
			gotErr := p.Pull(ctx, filter, gotResult)
			if diff := testutil.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.name, diff)
			}

			if diff := cmp.Diff(tc.wantResult, <-gotResult, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got result diff (-want, +got): %v", tc.name, diff)
			}

			if diff := cmp.Diff(tc.wantReq, tc.server.req, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got request diff (-want, +got): %v", tc.name, diff)
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
	// fakeClient := loggingpb.NewLoggingServiceV2Client(conn)
	fakeClient, err := logging.NewClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("creating client for fake at %q: %v", addr, err)
	}
	return fakeClient
}

type fakeServer struct {
	loggingpb.UnimplementedLoggingServiceV2Server

	req         *loggingpb.ListLogEntriesRequest
	resp        *loggingpb.ListLogEntriesResponse
	injectedErr error
}

func (s *fakeServer) ListLogEntries(ctx context.Context, req *loggingpb.ListLogEntriesRequest) (*loggingpb.ListLogEntriesResponse, error) {
	s.req = req
	return s.resp, s.injectedErr
}
