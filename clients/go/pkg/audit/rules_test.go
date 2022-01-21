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

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

func TestIsRuleApplicable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rule       *alpb.AuditRule
		methodName string
		want       bool
	}{
		{
			name: "wildcard_matches",
			rule: &alpb.AuditRule{
				Selector: "*",
			},
			methodName: "foo",
			want:       true,
		},
		{
			name: "wildcard_suffix_matches",
			rule: &alpb.AuditRule{
				Selector: "foo.*",
			},
			methodName: "foo.get",
			want:       true,
		},
		{
			name: "wildcard_suffix_barley_matches",
			rule: &alpb.AuditRule{
				Selector: "foo*",
			},
			methodName: "foo",
			want:       true,
		},
		{
			name: "wildcard_suffix_mismatches",
			rule: &alpb.AuditRule{
				Selector: "foo.*",
			},
			methodName: "bar.get",
			want:       false,
		},
		{
			name: "exact_match",
			rule: &alpb.AuditRule{
				Selector: "foo.get",
			},
			methodName: "foo.get",
			want:       true,
		},
		{
			name: "exact_mismatch",
			rule: &alpb.AuditRule{
				Selector: "foo.get",
			},
			methodName: "bar.get",
			want:       false,
		},
		{
			name: "exact_mismatch_again",
			rule: &alpb.AuditRule{
				Selector: "foo.get",
			},
			methodName: "bar.getgud",
			want:       false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isRuleApplicable(tc.rule, tc.methodName)
			if got != tc.want {
				t.Errorf("isApplicable(%v) = %v, want %v", tc.methodName, got, tc.want)
			}
		})
	}
}

func TestMostRelevantRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rules      []*alpb.AuditRule
		methodName string
		wantRule   alpb.AuditRule
	}{
		{
			name: "exact_match_wins",
			rules: []*alpb.AuditRule{
				{Selector: "a.b.c"},
				{Selector: "a.b.*"},
				{Selector: "*"},
			},
			methodName: "a.b.c",
			wantRule:   alpb.AuditRule{Selector: "a.b.c"},
		},
		{
			name: "partial_wildcard_match_wins",
			rules: []*alpb.AuditRule{
				{Selector: "a.b.c"},
				{Selector: "a.b.*"},
				{Selector: "*"},
			},
			methodName: "a.b.d",
			wantRule:   alpb.AuditRule{Selector: "a.b.*"},
		},
		{
			name: "partial_wildcard_match_wins_again",
			rules: []*alpb.AuditRule{
				{Selector: "*"},
				{Selector: "a.b.*"},
				{Selector: "a.b.c"},
			},
			methodName: "a.b.d",
			wantRule:   alpb.AuditRule{Selector: "a.b.*"},
		},
		{
			name: "wildcard_match_wins",
			rules: []*alpb.AuditRule{
				{Selector: "a.b.c"},
				{Selector: "a.b.*"},
				{Selector: "*"},
			},
			methodName: "d.e.f",
			wantRule:   alpb.AuditRule{Selector: "*"},
		},
		{
			name: "no_match",
			rules: []*alpb.AuditRule{
				{Selector: "a.b.c"},
				{Selector: "a.b.*"},
				{Selector: "a.*"},
			},
			methodName: "d.e.f",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotRule := mostRelevantRule(tc.methodName, tc.rules)
			if gotRule != tc.wantRule {
				t.Errorf("mostRelevantRule(%v, %v) = %v, want %v", tc.methodName, tc.rules, gotRule, tc.wantRule)
			}
		})
	}
}
