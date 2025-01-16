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
	"google.golang.org/protobuf/testing/protocmp"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/auditerrors"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
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
			wantErr:    auditerrors.ErrInvalidRequest,
		},
		{
			name:    "should_error_when_logReq_is_nil",
			wantErr: auditerrors.ErrInvalidRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := NewRequestValidator(ctx)
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
