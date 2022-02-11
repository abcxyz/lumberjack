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
			name:   "nil_payload_should_error_from_default_validator",
			logReq: &alpb.AuditLogRequest{},
			opts: []Option{
				WithLogMode(alpb.AuditLogRequest_FAIL_CLOSE),
			},
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
			name:   "injected_error_in_mutator_should_return_error_on_fail_close",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithMutator(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error")}),
				WithLogMode(alpb.AuditLogRequest_FAIL_CLOSE),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
			wantErrSubstr: "failed to execute mutator",
		},
		{
			name:   "injected_error_in_mutator_should_return_nil_on_best_effort",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithMutator(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error")}),
				WithLogMode(alpb.AuditLogRequest_BEST_EFFORT),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
		},
		{
			name:   "failed_precondition_in_mutator_should_return_nil_on_best_effort",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithMutator(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error: %w", ErrFailedPrecondition)}),
				WithLogMode(alpb.AuditLogRequest_BEST_EFFORT),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
		},
		{
			name:   "injected_error_in_backend_should_return_error_on_fail_close",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithBackend(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error")}),
				WithLogMode(alpb.AuditLogRequest_FAIL_CLOSE),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
			wantErrSubstr: "failed to execute backend",
		},
		{
			name:   "failed_precondition_in_backend_should_return_error_on_fail_close",
			logReq: testutil.ReqBuilder().Build(),
			opts: []Option{
				WithBackend(testOrderProcessor{name: "fake", returnErr: fmt.Errorf("fake error: %w", ErrFailedPrecondition)}),
				WithLogMode(alpb.AuditLogRequest_FAIL_CLOSE),
			},
			wantLogReq: testutil.ReqBuilder().WithLabels(
				map[string]string{processorOrderKey: "fake, "},
			).Build(),
			wantErrSubstr: "failed to execute backend",
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
			t.Cleanup(func() {
				err := c.Stop()
				if err != nil {
					t.Errorf("failed to stop client: %v", err)
				}
			})
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

func TestHandleReturn_Client(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name           string
		clientLogMode  alpb.AuditLogRequest_LogMode
		requestLogMode alpb.AuditLogRequest_LogMode
		err            error
		wantErr        bool
	}{
		{
			name:          "returns_nil_with_err_best_effort",
			clientLogMode: alpb.AuditLogRequest_BEST_EFFORT,
			err:           errors.New("test error"),
			wantErr:       false,
		},
		{
			name:          "returns_err_with_err_fail_close",
			clientLogMode: alpb.AuditLogRequest_FAIL_CLOSE,
			err:           errors.New("test error"),
			wantErr:       true,
		},
		{
			name:           "returns_nil_with_err_request_is_best_effort",
			clientLogMode:  alpb.AuditLogRequest_FAIL_CLOSE,
			requestLogMode: alpb.AuditLogRequest_BEST_EFFORT,
			err:            errors.New("test error"),
			wantErr:        false,
		},
		{
			name:           "returns_err_with_err_request_is_fail_close",
			clientLogMode:  alpb.AuditLogRequest_BEST_EFFORT,
			requestLogMode: alpb.AuditLogRequest_FAIL_CLOSE,
			err:            errors.New("test error"),
			wantErr:        true,
		},
		{
			name:          "returns_nil_no_err",
			clientLogMode: alpb.AuditLogRequest_FAIL_CLOSE,
			wantErr:       false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts := []Option{
				WithLogMode(tc.clientLogMode),
			}
			c, err := NewClient(opts...)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				err := c.Stop()
				if err != nil {
					t.Errorf("failed to stop client: %v", err)
				}
			})

			gotErr := c.handleReturn(ctx, tc.err, tc.requestLogMode)

			if (gotErr != nil) != tc.wantErr {
				expected := "an error"
				if !tc.wantErr {
					expected = "nil"
				}
				t.Errorf("returned %v, but expected %v", gotErr, expected)
			}
		})
	}
}
