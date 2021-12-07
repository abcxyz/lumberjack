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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
)

const processorOrderKey = "processorOrder"

type testOrderProcessor struct {
	name      string
	returnErr error
}

func (p testOrderProcessor) Process(_ context.Context, logReq *alpb.AuditLogRequest) error {
	if logReq.Labels == nil {
		logReq.Labels = map[string]string{}
	}
	logReq.Labels[processorOrderKey] += p.name + ", "
	return p.returnErr
}

func TestLog(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := []struct {
		name          string
		logReq        *alpb.AuditLogRequest
		opts          []Option
		wantLogReq    *alpb.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name:       "well_formed_log_should_succeed_with_default_options",
			logReq:     testutil.ReqBuilder().Build(),
			wantLogReq: testutil.ReqBuilder().Build(),
		},
		{
			name:          "nil_payload_should_error_from_default_validator",
			logReq:        &alpb.AuditLogRequest{},
			wantLogReq:    &alpb.AuditLogRequest{},
			wantErrSubstr: "failed to execute validator",
		},
		{
			name:   "should_run_validator_then_mutator_then_backend",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithBackend(testOrderProcessor{name: "backend"}),
				WithMutator(testOrderProcessor{name: "mutator"}),
				WithValidator(testOrderProcessor{name: "validator"}),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "validator, mutator, backend, "},
			).Build(),
		},
		{
			name:   "multiple_ordered_validators_should_run_before_multiple_ordered_backends",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithBackend(testOrderProcessor{name: "backend0"}),
				WithValidator(testOrderProcessor{name: "validator0"}),
				WithValidator(testOrderProcessor{name: "validator1"}),
				WithBackend(testOrderProcessor{name: "backend1"}),
				WithValidator(testOrderProcessor{name: "validator2"}),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "validator0, validator1, validator2, backend0, backend1, "}).Build(),
		},
		{
			name:   "skip_subsequent_processors_when_precondition_failed",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithValidator(testOrderProcessor{name: "validator0"}),
				WithValidator(testOrderProcessor{name: "validator1", returnErr: fmt.Errorf("skip: %w", ErrFailedPrecondition)}),
				WithBackend(testOrderProcessor{name: "backend0"}),
				WithBackend(testOrderProcessor{name: "backend1"}),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "validator0, validator1, "},
			).Build(),
		},
		{
			name:   "injected_error_in_mutator_should_return_error",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithMutator(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error")}),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
			wantErrSubstr: "failed to execute mutator",
		},
		{
			name:   "failed_precondition_in_mutator_should_return_nil",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithMutator(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error: %w", ErrFailedPrecondition)}),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
		},
		{
			name:   "injected_error_in_backend_should_return_error",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithBackend(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error")}),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
			wantErrSubstr: "failed to execute backend",
		},
		{
			name:   "failed_precondition_in_backend_should_return_nil",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithBackend(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error: %w", ErrFailedPrecondition)}),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c, err := NewClient(test.opts...)
			if err != nil {
				t.Fatal(err)
			}
			defer c.Stop()
			err = c.Log(ctx, test.logReq)
			if diff := errutil.DiffSubstring(err, test.wantErrSubstr); diff != "" {
				t.Errorf("Log(%+v) got unexpected error substring: %v", test.logReq, diff)
			}
			if diff := cmp.Diff(test.wantLogReq, test.logReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got diff (-want, +got): %v", test.logReq, diff)
			}
		})
	}
}
