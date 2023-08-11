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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestTailCommand(t *testing.T) {
	t.Parallel()

	ct := time.Now().UTC()

	validLog := &loggingpb.LogEntry{
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
	}

	bs, err := protojson.Marshal(validLog)
	if err != nil {
		t.Fatalf("failed to mashal log to JSON: %v", err)
	}
	validLogJSON := stripSpaces(string(bs))

	cases := []struct {
		name            string
		args            []string
		puller          *fakePuller
		expFilter       string
		expMaxNum       int
		expOut          string
		expErrSubstr    string
		expStderrSubstr string
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
				logEntries: []*loggingpb.LogEntry{validLog},
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
			expOut: fmt.Sprintf(`%s
Successfully validated log (InsertId: "test-log")

Validation failed for 0 logs (out of 1)
`, validLogJSON),
		},
		{
			name: "success_validate_with_additional_check",
			args: []string{
				"-scope", "projects/foo",
				"-validate", "-additional-check",
			},
			puller: &fakePuller{
				logEntries: []*loggingpb.LogEntry{validLog},
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
			expOut: fmt.Sprintf(`%s
Successfully validated log (InsertId: "test-log")

Validation failed for 0 logs (out of 1)
`, validLogJSON),
		},
		{
			name: "validate_fail",
			args: []string{
				"-scope", "projects/foo",
				"-override-filter", "test-filter",
				"-v",
			},
			puller: &fakePuller{
				logEntries: []*loggingpb.LogEntry{
					{InsertId: "test"},
					validLog,
				},
			},
			expFilter: "test-filter",
			expMaxNum: 1,
			expOut: fmt.Sprintf(`
{"insertId":"test"}
%s
Successfully validated log (InsertId: "test-log")

Validation failed for 1 logs (out of 2)
`, validLogJSON),
			expStderrSubstr: `failed to validate log (InsertId: "test")`,
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
			_, stdout, stderr := cmd.Pipe()

			err := cmd.Run(ctx, tc.args)
			if diff := testutil.DiffErrString(err, tc.expErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got error diff (-want, +got):\n%s", tc.name, diff)
			}
			if !errContainSubstring(stderr.String(), tc.expStderrSubstr) {
				t.Errorf("Process(%+v) got stderr: %q, but want substring: %q", tc.name, stderr.String(), tc.expStderrSubstr)
			}
			if strings.TrimSpace(tc.expOut) != strings.TrimSpace(stdout.String()) {
				t.Errorf("Process(%+v) got output: %q, but want output: %q", tc.name, stdout.String(), tc.expOut)
			}
			if strings.TrimSpace(tc.expFilter) != strings.TrimSpace(tc.puller.gotFilter) {
				t.Errorf("Process(%+v) got filter: %q, but want output: %q", tc.name, tc.puller.gotFilter, tc.expFilter)
			}
			if tc.expMaxNum != tc.puller.gotMaxNum {
				t.Errorf("Process(%+v) want max number %q but got %q", tc.name, tc.expMaxNum, tc.puller.gotMaxNum)
			}
		})
	}
}

func errContainSubstring(gotErr, wantErr string) bool {
	if wantErr == "" {
		return gotErr == ""
	}
	return strings.Contains(gotErr, wantErr)
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
