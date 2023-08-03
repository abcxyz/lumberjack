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

	"google.golang.org/genproto/googleapis/cloud/audit"

	pkgtestutil "github.com/abcxyz/pkg/testutil"
)

func TestValidateAuditLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		payload       *audit.AuditLog
		wantErrSubstr string
	}{
		{
			name: "success",
			payload: &audit.AuditLog{
				MethodName:   "test-method",
				ServiceName:  "test-service",
				ResourceName: "test-resource",
				AuthenticationInfo: &audit.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
			},
		},
		{
			name:          "payload_is_nil",
			wantErrSubstr: "audit log payload cannot be nil",
		},
		{
			name: "authInfo_is_nil",
			payload: &audit.AuditLog{
				MethodName:   "test-method",
				ServiceName:  "test-service",
				ResourceName: "test-resource",
			},
			wantErrSubstr: "AuthenticationInfo cannot be nil",
		},
		{
			name: "auth_email_is_nil",
			payload: &audit.AuditLog{
				MethodName:         "test-method",
				ServiceName:        "test-service",
				ResourceName:       "test-resource",
				AuthenticationInfo: &audit.AuthenticationInfo{},
			},
			wantErrSubstr: "PrincipalEmail cannot be empty",
		},
		{
			name: "auth_email_has_no_domain",
			payload: &audit.AuditLog{
				MethodName:   "test-method",
				ServiceName:  "test-service",
				ResourceName: "test-resource",
				AuthenticationInfo: &audit.AuthenticationInfo{
					PrincipalEmail: "user",
				},
			},
			wantErrSubstr: `PrincipalEmail "user" is malformed`,
		},
		{
			name: "serviceName_is_empty",
			payload: &audit.AuditLog{
				MethodName:   "test-method",
				ResourceName: "test-resource",
				AuthenticationInfo: &audit.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
			},
			wantErrSubstr: "ServiceName cannot be empty",
		},
		{
			name: "resourceName_is_empty",
			payload: &audit.AuditLog{
				MethodName:  "test-method",
				ServiceName: "test-service",
				AuthenticationInfo: &audit.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
			},
			wantErrSubstr: "ResourceName cannot be empty",
		},
		{
			name: "methodName_is_empty",
			payload: &audit.AuditLog{
				ServiceName:  "test-service",
				ResourceName: "test-resource",
				AuthenticationInfo: &audit.AuthenticationInfo{
					PrincipalEmail: "user@example.com",
				},
			},
			wantErrSubstr: "MethodName cannot be empty",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateAuditLog(tc.payload)
			if diff := pkgtestutil.DiffErrString(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.name, diff)
			}
		})
	}
}
