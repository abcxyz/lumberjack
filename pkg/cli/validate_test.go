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
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
)

const validLog = `
{
	"insertId": "foo",
	"jsonPayload": {
		"metadata": {
			"originating_resource": {
				"labels": {
					"service_name": "foo_service",
					"project_id": "foo_project",
					"revision_name": "foo_revision",
					"configuration_name": "foo_configuration",
					"location": "us-central1"
				},
				"type": "foo_type"
			}
		},
		"request": {
			"foo": "bar",
			"trace_id": "foo_trace_id"
		},
		"service_name": "foo_service",
		"method_name": "foo_method",
		"resource_name": "foo_resource",
		"authentication_info": {
			"principal_email": "foo@bet.com"
		}
	},
	"resource": {
		"type": "foo_type",
		"labels": {
			"project_id": "foo_project",
			"configuration_name": "foo_configuration_name",
			"service_name": "foo_service",
			"location": "us-central1",
			"revision_name": "foo_revision"
		}
	},
	"timestamp": "2022-01-19T22:05:15.687616341Z",
	"labels": {
		"trace_id": "foo_trace_id",
		"accessing_process_family": "foo-process-family",
		"accessing_process_name": "foo-process",
		"accessed_data_type": "foo-customer-info",
		"environment": "dev"
	},
	"logName": "projects/foo_project/logs/auditlog.gcloudsolutions.dev%2Fdata_access",
	"receiveTimestamp": "2022-01-19T22:05:15.706507037Z"
}`

const missingLabel = `
{
	"jsonPayload": {
		"service_name": "foo_service",
		"method_name": "foo_method",
		"resource_name": "foo_resource",
		"authentication_info": {
			"principal_email": "foo@bet.com"
		}
	},
	"labels": {}
}`

func TestValidateCommand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		args   []string
		stdin  io.Reader
		expOut string
		expErr string
	}{
		{
			name:   "success",
			args:   []string{"-log-entry", validLog},
			expOut: `Successfully validated log`,
		},
		{
			name:   "from_stdin",
			args:   []string{"-log-entry", "-"},
			stdin:  strings.NewReader(validLog),
			expOut: `Successfully validated log`,
		},
		{
			name:   "additional_check",
			args:   []string{"-log-entry", validLog, "-additional-check"},
			expOut: `Successfully validated log`,
		},
		{
			name:   "unexpected_args",
			args:   []string{"foo"},
			expErr: `unexpected arguments: ["foo"]`,
		},
		{
			name:   "missing_log",
			args:   []string{},
			expErr: `log is required`,
		},
		{
			name:   "invalid_json",
			args:   []string{"-log-entry", "invalid"},
			expErr: "failed to validate log",
		},
		{
			name:   "additional_check_fail",
			args:   []string{"-log-entry", missingLabel, "-additional-check"},
			expErr: `missing required label`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := logging.WithLogger(context.Background(), logging.TestLogger(t))

			var cmd ValidateCommand
			stdin, stdout, _ := cmd.Pipe()

			// Write stdin if given
			if tc.stdin != nil {
				if _, err := io.Copy(stdin, tc.stdin); err != nil {
					t.Fatal(err)
				}
			}

			args := append([]string{}, tc.args...)

			err := cmd.Run(ctx, args)
			if diff := testutil.DiffErrString(err, tc.expErr); diff != "" {
				t.Errorf("Process(%+v) got error diff (-want, +got):\n%s", tc.name, diff)
			}
			if diff := cmp.Diff(strings.TrimSpace(tc.expOut), strings.TrimSpace(stdout.String())); diff != "" {
				t.Errorf("Process(%+v) got output diff (-want, +got):\n%s", tc.name, diff)
			}
		})
	}
}
