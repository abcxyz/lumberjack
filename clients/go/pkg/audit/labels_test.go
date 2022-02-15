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

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestProcessLabels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name       string
		configLabels map[string]string
		logReq     *alpb.AuditLogRequest
		wantLogReq *alpb.AuditLogRequest
		wantErr    error
	}{
		{
			name:       "adds_labels",
			configLabels: map[string]string{"label1": "value1"},
			logReq:     testutil.ReqBuilder().Build(),
			wantLogReq: testutil.ReqBuilder().WithLabels(map[string]string{"label1": "value1"}).Build(),
		},
		{
			name:       "adds_labels_without_overwriting",
			configLabels: map[string]string{"label1": "value1", "label2": "value2"},
			logReq:     testutil.ReqBuilder().WithLabels(map[string]string{"label1": "requestval"}).Build(),
			wantLogReq: testutil.ReqBuilder().WithLabels(map[string]string{"label1": "requestval", "label2": "value2"}).Build(),
		},
		{
			name:       "adds_nothing_if_nil_labels",
			logReq:     testutil.ReqBuilder().WithLabels(map[string]string{"label1": "requestval"}).Build(),
			wantLogReq: testutil.ReqBuilder().WithLabels(map[string]string{"label1": "requestval"}).Build(),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := &LabelProcessor{DefaultLabels: tc.configLabels}
			err := l.Process(ctx, tc.logReq)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Process(%+v) error got %T want %T", tc.logReq, err, tc.wantErr)
			}

			if diff := cmp.Diff(tc.wantLogReq, tc.logReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got diff (-want, +got): %v", tc.logReq, diff)
			}
		})
	}
}