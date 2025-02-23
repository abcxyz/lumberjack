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

package filtering

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	cal "google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/testing/protocmp"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/auditerrors"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
	pkgtestutil "github.com/abcxyz/pkg/testutil"
)

func TestNewPrincipalEmailMatcher(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		opts          []Option
		wantErrSubstr string
	}{
		{
			name: "should_succeed_when_include_and_exclude_pass_regexp_compilation",
			opts: []Option{
				WithIncludes(`a`),
				WithExcludes(`b`),
			},
		},
		{
			name: "should_error_when_include_fails_regexp_compilation",
			opts: []Option{
				WithIncludes(`\`),
			},
			wantErrSubstr: "failed to apply NewPrincipalEmailMatcher options: failed to compile include string",
		},
		{
			name: "should_error_when_exclude_fails_regexp_compilation",
			opts: []Option{
				WithExcludes(`\`),
			},
			wantErrSubstr: "failed to apply NewPrincipalEmailMatcher options: failed to compile exclude string",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewPrincipalEmailMatcher(tc.opts...)
			if diff := pkgtestutil.DiffErrString(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("NewPrincipalEmailMatcher(%v) got unexpected error substring: %v", tc.opts, diff)
			}
		})
	}
}

func TestPrincipalEmailMatcher_Process(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		opts       []Option
		logReq     *api.AuditLogRequest
		wantLogReq *api.AuditLogRequest
		wantErr    error
	}{
		{
			name:       "should_succeed_when_include_and_exclude_are_nil",
			logReq:     testutil.NewRequest(),
			wantLogReq: testutil.NewRequest(),
		},
		{
			name:       "should_succeed_when_include_matches_and_exclude_is_nil",
			opts:       []Option{WithIncludes("foo@google.com")},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
		},
		{
			name:       "should_fail_precondition_when_include_mismatches_and_exclude_is_nil",
			opts:       []Option{WithIncludes("foo@google.com")},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("bar@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("bar@google.com")),
			wantErr:    auditerrors.ErrPreconditionFailed,
		},
		{
			name:       "should_fail_precondition_when_exclude_matches_and_include_is_nil",
			opts:       []Option{WithExcludes("foo@google.com")},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
			wantErr:    auditerrors.ErrPreconditionFailed,
		},
		{
			name:       "should_succeed_when_exclude_mismatches_and_include_is_nil",
			opts:       []Option{WithExcludes("foo@google.com")},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("bar@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("bar@google.com")),
		},
		{
			name: "should_fail_precondition_when_include_mismatches_and_exclude_matches",
			opts: []Option{
				WithIncludes("foo@google.com"),
				WithExcludes("bar@google.com"),
			},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("bar@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("bar@google.com")),
			wantErr:    auditerrors.ErrPreconditionFailed,
		},
		{
			name: "should_succeed_when_include_matches_and_exclude_mismatches",
			opts: []Option{
				WithIncludes("foo@google.com"),
				WithExcludes("bar@google.com"),
			},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
		},
		{
			name: "should_succeed_when_include_matches_and_exclude_matches",
			opts: []Option{
				WithIncludes("foo@google.com"),
				WithExcludes("@google.com"),
			},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
		},
		{
			name: "should_work_as_intended_with_multiple_includes_and_excludes",
			opts: []Option{
				WithIncludes("foo@google.com", "bar@google.com"),
				WithExcludes("baz@google.com", "qux@google.com"),
			},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("bar@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("bar@google.com")),
		},
		{
			name: "empty_string_as_exclude_is_a_noop",
			opts: []Option{
				WithExcludes(""),
			},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
		},
		{
			name: "empty_string_as_include_is_a_noop",
			opts: []Option{
				WithIncludes(""),
				WithExcludes("."),
			},
			logReq:     testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
			wantLogReq: testutil.NewRequest(testutil.WithPrincipal("foo@google.com")),
			wantErr:    auditerrors.ErrPreconditionFailed,
		},
		{
			name: "should_fail_if_authentication_info_is_missing",
			opts: []Option{
				WithIncludes("foo@google.com"),
			},
			logReq: &api.AuditLogRequest{
				Payload: &cal.AuditLog{},
			},
			wantLogReq: &api.AuditLogRequest{
				Payload: &cal.AuditLog{},
			},
			wantErr: auditerrors.ErrInvalidRequest,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m, err := NewPrincipalEmailMatcher(tc.opts...)
			if err != nil {
				t.Fatalf("NewPrincipalEmailMatcher(%v) unexpected err: %v", tc.opts, err)
			}

			err = m.Process(t.Context(), tc.logReq)
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
