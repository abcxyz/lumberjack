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

func TestTailCommand(t *testing.T) {
	t.Parallel()

	ct := time.Now().UTC()

	cases := []struct {
		name         string
		args         []string
		puller       *fakePuller
		expFilter    string
		expMaxNum    int
		expOut       string
		expErrSubstr string
	}{
		{
			name: "success_tail",
			args: []string{"-scope", "projects/foo"},
			puller: &fakePuller{
				logEntries: []*loggingpb.LogEntry{{}},
			},
			expFilter: fmt.Sprintf(
				`LOG_ID("audit.abcxyz/unspecified") OR `+
					`LOG_ID("audit.abcxyz/activity") OR `+
					`LOG_ID("audit.abcxyz/data_access") OR `+
					`LOG_ID("audit.abcxyz/consent") OR `+
					`LOG_ID("audit.abcxyz/system_event") `+
					`AND timestamp >= %q`,
				ct.Add(-2*time.Hour).Format(time.RFC3339),
			),
			expMaxNum: 1,
			expOut:    `{}`,
		},
		{
			name: "success_validate",
			args: []string{
				"-scope", "projects/foo",
				"-duration", "4h",
				"-additional-filter", `resource.type = "gae_app" AND severity = ERROR`,
				"-validate",
				"-max-num", "2",
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
				ct.Add(-4*time.Hour).Format(time.RFC3339),
			),
			expMaxNum: 2,
			expOut: `{"jsonPayload":{"authentication_info":{"principal_email":"foo@bet.com"}, ` +
				`"method_name":"foo_method", ` +
				`"resource_name":"foo_resource", ` +
				`"service_name":"foo_service"}, ` +
				`"insertId":"test-log"}
Successfully validated log (InsertId: "test-log")

`,
		},
		{
			name: "success_validate_with_additional_check",
			args: []string{
				"-scope", "projects/foo",
				"-validate", "-additional-check",
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
					`AND timestamp >= %q`,
				ct.Add(-2*time.Hour).Format(time.RFC3339),
			),
			expMaxNum: 1,
			expOut: `{"jsonPayload":{"authentication_info":{"principal_email":"foo@bet.com"}, ` +
				`"method_name":"foo_method", ` +
				`"resource_name":"foo_resource", ` +
				`"service_name":"foo_service"}, ` +
				`"insertId":"test-log", ` +
				`"labels":{"accessing_process_name":"foo_apn", "environment":"dev"}}
Successfully validated log (InsertId: "test-log")

`,
		},
		{
			name: "validate_fail",
			args: []string{
				"-scope", "projects/foo",
				"-override-filter", "test-filter",
				"-v",
			},
			puller: &fakePuller{
				logEntries: []*loggingpb.LogEntry{{InsertId: "test"}},
			},
			expFilter:    "test-filter",
			expMaxNum:    1,
			expOut:       `{"insertId":"test"}`,
			expErrSubstr: "failed to validate log",
		},
		{
			name: "tail_fail",
			args: []string{
				"-scope", "projects/foo",
				"-override-filter", "test-filter",
			},
			puller: &fakePuller{
				logEntries: []*loggingpb.LogEntry{{InsertId: "test"}},
				injectErr:  fmt.Errorf("injected error"),
			},
			expFilter:    "test-filter",
			expMaxNum:    1,
			expErrSubstr: "injected error",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var cmd TailCommand
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
				t.Errorf("Process(%+v) got filter diff (-want, +got):\n%s", tc.name, diff)
			}
			if tc.expMaxNum != tc.puller.gotMaxNum {
				t.Errorf("Process(%+v) want max number %q but got %q", tc.name, tc.expMaxNum, tc.puller.gotMaxNum)
			}
		})
	}
}

type fakePuller struct {
	injectErr  error
	gotFilter  string
	gotMaxNum  int
	logEntries []*loggingpb.LogEntry
}

func (p *fakePuller) Pull(ctx context.Context, filter string, maxNum int) ([]*loggingpb.LogEntry, error) {
	p.gotFilter = filter
	p.gotMaxNum = maxNum
	return p.logEntries, p.injectErr
}
