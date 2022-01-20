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
)

func TestIsApplicable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rule       Rule
		methodName string
		want       bool
	}{
		{
			name: "wildcard_matches",
			rule: Rule{
				Selector: "*",
			},
			methodName: "foo",
			want:       true,
		},
		{
			name: "wildcard_suffix_matches",
			rule: Rule{
				Selector: "foo.*",
			},
			methodName: "foo.get",
			want:       true,
		},
		{
			name: "wildcard_suffix_barley_matches",
			rule: Rule{
				Selector: "foo*",
			},
			methodName: "foo",
			want:       true,
		},
		{
			name: "wildcard_suffix_mismatches",
			rule: Rule{
				Selector: "foo.*",
			},
			methodName: "bar.get",
			want:       false,
		},
		{
			name: "exact_match",
			rule: Rule{
				Selector: "foo.get",
			},
			methodName: "foo.get",
			want:       true,
		},
		{
			name: "exact_mismatch",
			rule: Rule{
				Selector: "foo.get",
			},
			methodName: "bar.get",
			want:       false,
		},
		{
			// todo: ask Ryan if this test case is consistent with Java
			name: "exact_mismatch_again",
			rule: Rule{
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

			got := tc.rule.isApplicable(tc.methodName)
			if got != tc.want {
				t.Errorf("isApplicable(%v) = %v, want %v", tc.methodName, got, tc.want)
			}
		})
	}
}

func TestMostRelevant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rules      []Rule
		methodName string
		wantRule   Rule
	}{
		{
			name: "exact_match_wins",
			rules: []Rule{
				{Selector: "a.b.c"},
				{Selector: "a.b.*"},
				{Selector: "*"},
			},
			methodName: "a.b.c",
			wantRule:   Rule{Selector: "a.b.c"},
		},
		{
			name: "partial_wildcard_match_wins",
			rules: []Rule{
				{Selector: "a.b.c"},
				{Selector: "a.b.*"},
				{Selector: "*"},
			},
			methodName: "a.b.d",
			wantRule:   Rule{Selector: "a.b.*"},
		},
		{
			name: "partial_wildcard_match_wins_again",
			rules: []Rule{
				{Selector: "*"},
				{Selector: "a.b.*"},
				{Selector: "a.b.c"},
			},
			methodName: "a.b.d",
			wantRule:   Rule{Selector: "a.b.*"},
		},
		{
			name: "wildcard_match_wins",
			rules: []Rule{
				{Selector: "a.b.c"},
				{Selector: "a.b.*"},
				{Selector: "*"},
			},
			methodName: "d.e.f",
			wantRule:   Rule{Selector: "*"},
		},
		{
			name: "no_match",
			rules: []Rule{
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

			gotRule := mostRelevant(tc.methodName, tc.rules)
			if gotRule != tc.wantRule {
				t.Errorf("mostRelevant(%v, %v) = %v, want %v", tc.methodName, tc.rules, gotRule, tc.wantRule)
			}
		})
	}
}
