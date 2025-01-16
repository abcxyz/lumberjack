// Copyright 2022 Lumberjack authors (see AUTHORS file)
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
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
)

func TestProcessLabels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name         string
		configLabels map[string]string
		logReq       *api.AuditLogRequest
		wantLogReq   *api.AuditLogRequest
		wantErr      error
	}{
		{
			name:         "adds_labels",
			configLabels: map[string]string{"label1": "value1"},
			logReq:       testutil.NewRequest(),
			wantLogReq:   testutil.NewRequest(testutil.WithLabels(map[string]string{"label1": "value1"})),
		},
		{
			name:         "adds_labels_without_overwriting",
			configLabels: map[string]string{"label1": "value1", "label2": "value2"},
			logReq:       testutil.NewRequest(testutil.WithLabels(map[string]string{"label1": "requestval"})),
			wantLogReq:   testutil.NewRequest(testutil.WithLabels(map[string]string{"label1": "requestval", "label2": "value2"})),
		},
		{
			name:       "adds_nothing_if_nil_labels",
			logReq:     testutil.NewRequest(testutil.WithLabels(map[string]string{"label1": "requestval"})),
			wantLogReq: testutil.NewRequest(testutil.WithLabels(map[string]string{"label1": "requestval"})),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := NewLabelProcessor(ctx, tc.configLabels)
			err := l.Process(ctx, tc.logReq)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Process(%+v) error got %v want %v", tc.logReq, err, tc.wantErr)
			}

			if diff := cmp.Diff(tc.wantLogReq, tc.logReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got diff (-want, +got): %v", tc.logReq, diff)
			}
		})
	}
}
