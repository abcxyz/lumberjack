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
	"strings"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

const wildcard = "*"

// mostRelevantRule finds the most relevant Rule for a given method by
// comparing the Rules's Selector length. E.g. given the methodName
// "com.example.Hello", the selector relevance is:
// "com.example.Hello" > "com.example.*" > "*"
//
// If none of the Rules are relevant to the given method (i.e. the
// the selectors don't match), we return nil.
func mostRelevantRule(methodName string, rules []*alpb.AuditRule) *alpb.AuditRule {
	var longest int
	var mostRelevant *alpb.AuditRule
	for _, r := range rules {
		if isRuleApplicable(r, methodName) && len(r.Selector) > longest {
			longest = len(r.Selector)
			mostRelevant = r
			if longest == len(methodName) {
				// Shortcircuit on exact match.
				return mostRelevant
			}
		}
	}
	return mostRelevant
}

// isRuleApplicable determines if a Rule applies to
// the given method by comparing the Rule's Selector
// to the methodName.
func isRuleApplicable(rule *alpb.AuditRule, methodName string) bool {
	sel := rule.Selector
	if sel == wildcard {
		return true
	}
	// Ignore any leading slashes when check rules.
	sel = strings.TrimLeft(sel, "/")
	methodName = strings.TrimLeft(methodName, "/")
	if strings.HasSuffix(sel, wildcard) {
		return strings.HasPrefix(methodName, sel[:len(sel)-1])
	}
	return sel == methodName
}
