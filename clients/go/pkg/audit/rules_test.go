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
	"testing"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

func TestMostRelevantRule(t *testing.T) {
	t.Parallel()

	ruleBySelector := map[string]*api.AuditRule{
		"a.b.c":   {Selector: "a.b.c"},
		"a.b.*":   {Selector: "a.b.*"},
		"a.*":     {Selector: "a.*"},
		"*":       {Selector: "*"},
		"a.b.d":   {Selector: "a.b.d"},
		"foo*":    {Selector: "foo*"},
		"//a.b.c": {Selector: "//a.b.c"},
		"/a.b.*":  {Selector: "/a.b.*"},
	}

	tests := []struct {
		name       string
		rules      []*api.AuditRule
		methodName string
		wantRule   *api.AuditRule
	}{
		{
			name: "exact_match_wins",
			rules: []*api.AuditRule{
				ruleBySelector["a.b.c"],
				ruleBySelector["a.b.*"],
				ruleBySelector["*"],
			},
			methodName: "a.b.c",
			wantRule:   ruleBySelector["a.b.c"],
		},
		{
			name: "partial_wildcard_match_wins",
			rules: []*api.AuditRule{
				ruleBySelector["a.b.c"],
				ruleBySelector["a.b.*"],
				ruleBySelector["*"],
			},
			methodName: "a.b.d",
			wantRule:   ruleBySelector["a.b.*"],
		},
		{
			name: "partial_wildcard_match_wins_again",
			rules: []*api.AuditRule{
				ruleBySelector["*"],
				ruleBySelector["a.b.*"],
				ruleBySelector["a.b.c"],
			},
			methodName: "a.b.d",
			wantRule:   ruleBySelector["a.b.*"],
		},
		{
			name: "wildcard_match_wins",
			rules: []*api.AuditRule{
				ruleBySelector["a.b.c"],
				ruleBySelector["a.b.*"],
				ruleBySelector["*"],
			},
			methodName: "d.e.f",
			wantRule:   ruleBySelector["*"],
		},
		{
			name: "wildcard_suffix_barley_matches",
			rules: []*api.AuditRule{
				ruleBySelector["foo*"],
			},
			methodName: "foo",
			wantRule:   ruleBySelector["foo*"],
		},
		{
			name: "no_match",
			rules: []*api.AuditRule{
				ruleBySelector["a.b.c"],
				ruleBySelector["a.b.*"],
				ruleBySelector["a.*"],
			},
			methodName: "d.e.f",
		},
		{
			name: "match_ignore_leading_slashes",
			rules: []*api.AuditRule{
				ruleBySelector["//a.b.c"],
				ruleBySelector["/a.b.*"],
				ruleBySelector["a.*"],
			},
			methodName: "//a.b.z",
			wantRule:   ruleBySelector["/a.b.*"],
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotRule := mostRelevantRule(tc.methodName, tc.rules)
			if gotRule != tc.wantRule {
				t.Errorf("mostRelevantRule(%v, rules) = %v, want %v", tc.methodName, gotRule, tc.wantRule)
			}
		})
	}
}
