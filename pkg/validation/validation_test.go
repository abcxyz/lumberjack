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

package validation

import (
	"testing"

	pkgtestutil "github.com/abcxyz/pkg/testutil"
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

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		log           string
		wantErrSubstr string
	}{
		{
			name: "success",
			log:  validLog,
		},
		{
			name:          "invalid_log",
			log:           `invalid`,
			wantErrSubstr: "failed to parse log entry as JSON",
		},
		{
			name: "missing_log_payload",
			log: `
{}`,
			wantErrSubstr: "missing audit log payload",
		},
		{
			name: "invalid_log_payload_json",
			log: `
{
	"jsonPayload": {
		"invalidKey": "foo"
	}
}`,
			wantErrSubstr: "failed to parse JSON payload",
		},
		{
			name: "invalid_log_payload",
			log: `
{
	"jsonPayload": {
		"service_name": "foo"
	}
}`,
			wantErrSubstr: "invalid payload",
		},
		{
			name:          "labels_is_empty",
			log:           `{}`,
			wantErrSubstr: "labels is empty",
		},
		{
			name: "missing_labels",
			log: `{
				"labels": {}
			}`,
			wantErrSubstr: `missing label: "environment"`,
		},
	}
	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(tc.log)
			if diff := pkgtestutil.DiffErrString(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.name, diff)
			}
		})
	}
}
