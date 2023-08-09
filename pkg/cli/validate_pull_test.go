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

package cli

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidatePullCommand(t *testing.T) {
	t.Parallel()

	ct := time.Now().UTC()

	cases := []struct {
		name         string
		args         []string
		puller       *fakePuller
		expFilter    string
		expMaxCount  int
		expOut       string
		expErrSubstr string
	}{
		{
			name: "success",
			args: []string{
				"-resource", "projects/foo",
				"-duration", "2h",
				"-query", `resource.type = "gae_app" AND severity = ERROR`,
				"-additional-check",
				"-max-count", "2",
			},
			puller: &fakePuller{
				logEntries: []*loggingpb.LogEntry{
					{
						InsertId: "test-log",
						Payload: &loggingpb.LogEntry_JsonPayload{
							JsonPayload: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"service_name":  structpb.NewStringValue("foo_service"),
									"method_name":   structpb.NewStringValue("foo_method"),
									"resource_name": structpb.NewStringValue("foo_resource"),
									"authentication_info": structpb.NewStructValue(&structpb.Struct{
										Fields: map[string]*structpb.Value{
											"principal_email": structpb.NewStringValue("foo@bet.com"),
										},
									}),
								},
							},
						},
						Labels: map[string]string{"environment": "dev", "accessing_process_name": "foo_apn"},
					},
				},
			},
			expFilter: fmt.Sprintf(
				`LOG_ID("audit.abcxyz/unspecified") OR `+
					`LOG_ID("audit.abcxyz/activity") OR `+
					`LOG_ID("audit.abcxyz/data_access") OR `+
					`LOG_ID("audit.abcxyz/consent") OR `+
					`LOG_ID("audit.abcxyz/system_event") `+
					`AND timestamp >= %q AND resource.type = "gae_app" `+
					`AND severity = ERROR`,
				ct.Add(-2*time.Hour).Format(time.RFC3339),
			),
			expMaxCount: 2,
			expOut:      `Successfully validated log (InsertId: "test-log")`,
		},
		{
			name: "validate_fail",
			args: []string{
				"-resource", "projects/foo",
				"-remove-lumberjack-log-type",
			},
			puller: &fakePuller{
				logEntries: []*loggingpb.LogEntry{{InsertId: "test"}},
			},
			expFilter: fmt.Sprintf(
				`timestamp >= %q`, ct.Add(-24*time.Hour).Format(time.RFC3339),
			),
			expMaxCount:  1,
			expErrSubstr: "failed to validate log",
		},
		{
			name: "pull_fail",
			args: []string{
				"-resource", "projects/foo",
				"-remove-lumberjack-log-type",
			},
			puller: &fakePuller{
				logEntries: []*loggingpb.LogEntry{{InsertId: "test"}},
				injectErr:  fmt.Errorf("injected error"),
			},
			expFilter: fmt.Sprintf(
				`timestamp >= %q`, ct.Add(-24*time.Hour).Format(time.RFC3339),
			),
			expMaxCount:  1,
			expErrSubstr: "injected error",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var cmd PullCommand
			cmd.testPuller = tc.puller
			_, stdout, _ := cmd.Pipe()

			err := cmd.Run(ctx, tc.args)
			if diff := testutil.DiffErrString(err, tc.expErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got error diff (-want, +got):\n%s", tc.name, diff)
			}
			if diff := cmp.Diff(strings.TrimSpace(tc.expOut), strings.TrimSpace(stdout.String())); diff != "" {
				t.Errorf("Process(%+v) got output diff (-want, +got):\n%s", tc.name, diff)
			}
			if diff := cmp.Diff(tc.expFilter, tc.puller.gotFilter); diff != "" {
				t.Errorf("Process(%+v) got request diff (-want, +got):\n%s", tc.name, diff)
			}
			if tc.expMaxCount != tc.puller.gotMaxCount {
				t.Errorf("Process(%+v) want max count %q but got %q", tc.name, tc.expMaxCount, tc.puller.gotMaxCount)
			}
		})
	}
}

type fakePuller struct {
	injectErr   error
	gotFilter   string
	gotMaxCount int
	logEntries  []*loggingpb.LogEntry
}

func (p *fakePuller) Pull(ctx context.Context, filter string, maxCount int) ([]*loggingpb.LogEntry, error) {
	p.gotFilter = filter
	p.gotMaxCount = maxCount
	return p.logEntries, p.injectErr
}
