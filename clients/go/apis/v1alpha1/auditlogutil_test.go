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

package v1alpha1

import (
	"testing"
)

func TestShouldFailClose(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		logMode AuditLogRequest_LogMode
		wanted  bool
	}{
		{
			name:    "should_return_true_on_fail_close",
			logMode: AuditLogRequest_FAIL_CLOSE,
			wanted:  true,
		},
		{
			name:    "should_return_false_on_best_effort",
			logMode: AuditLogRequest_BEST_EFFORT,
			wanted:  false,
		},
		{
			name:    "should_return_false_on_unspecified",
			logMode: AuditLogRequest_LOG_MODE_UNSPECIFIED,
			wanted:  false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ShouldFailClose(tc.logMode)
			if got != tc.wanted {
				t.Errorf("got=%t, wanted=%t", got, tc.wanted)
			}
		})
	}
}
