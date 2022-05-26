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

package audit

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/testing/protocmp"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/internal/testutil"
)

func TestRequestValidation_Process(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := []struct {
		name       string
		logReq     *api.AuditLogRequest
		wantLogReq *api.AuditLogRequest
		wantErr    error
	}{
		{
			name:       "valid_AuditLogRequest",
			logReq:     testutil.NewRequest(),
			wantLogReq: testutil.NewRequest(),
		},
		{
			name:       "should_error_when_logReq_payload_is_nil",
			logReq:     &api.AuditLogRequest{},
			wantLogReq: &api.AuditLogRequest{},
			wantErr:    ErrInvalidRequest,
		},
		{
			name:    "should_error_when_logReq_is_nil",
			wantErr: ErrInvalidRequest,
		},
		{
			name: "should_error_when_authInfo_is_nil",
			logReq: &api.AuditLogRequest{
				Payload: &audit.AuditLog{
					ServiceName: "test-service",
				},
			},
			wantLogReq: &api.AuditLogRequest{
				Payload: &audit.AuditLog{
					ServiceName: "test-service",
				},
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "should_error_when_auth_email_is_nil",
			logReq: &api.AuditLogRequest{
				Payload: &audit.AuditLog{
					ServiceName:        "test-service",
					AuthenticationInfo: &audit.AuthenticationInfo{},
				},
			},
			wantLogReq: &api.AuditLogRequest{
				Payload: &audit.AuditLog{
					ServiceName:        "test-service",
					AuthenticationInfo: &audit.AuthenticationInfo{},
				},
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "should_error_when_auth_email_has_no_domain",
			logReq: &api.AuditLogRequest{
				Payload: &audit.AuditLog{
					ServiceName: "test-service",
					AuthenticationInfo: &audit.AuthenticationInfo{
						PrincipalEmail: "user",
					},
				},
			},
			wantLogReq: &api.AuditLogRequest{
				Payload: &audit.AuditLog{
					ServiceName: "test-service",
					AuthenticationInfo: &audit.AuthenticationInfo{
						PrincipalEmail: "user",
					},
				},
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "should_error_when_serviceName_is_empty",
			logReq: &api.AuditLogRequest{
				Payload: &audit.AuditLog{
					ServiceName: "",
					AuthenticationInfo: &audit.AuthenticationInfo{
						PrincipalEmail: "user",
					},
				},
			},
			wantLogReq: &api.AuditLogRequest{
				Payload: &audit.AuditLog{
					ServiceName: "",
					AuthenticationInfo: &audit.AuthenticationInfo{
						PrincipalEmail: "user",
					},
				},
			},
			wantErr: ErrInvalidRequest,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &requestValidation{}
			err := p.Process(ctx, tc.logReq)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Process(%+v) error got %T want %T", tc.logReq, err, tc.wantErr)
			}
			// Verify that the log request is not modified.
			if diff := cmp.Diff(tc.wantLogReq, tc.logReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got diff (-want, +got): %v", tc.logReq, diff)
			}
		})
	}
}
