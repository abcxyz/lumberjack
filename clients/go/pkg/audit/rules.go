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

type Directive string

const (
	Audit                   Directive = "AUDIT"
	AuditRequestAndResponse Directive = "AUDIT_REQUEST_AND_RESPONSE"
	AuditRequestOnly        Directive = "AUDIT_REQUEST_ONLY"
)

// A Rule tells the middleware which methods should be logged and how
// they should be logged.
type Rule struct {
	Selector  string
	Directive Directive
	LogType   alpb.AuditLogRequest_LogType
}

// isApplicable determines if a Rule applies to
// the given method by comparing the Rule's Selector
// to the methodName.
func (r Rule) isApplicable(methodName string) bool {
	sel := r.Selector
	if sel == wildcard {
		return true
	}
	if strings.HasSuffix(sel, wildcard) {
		return strings.HasPrefix(methodName, sel[:len(sel)-1])
	}
	return sel == methodName
}

// mostRelevant finds the most relevant Rule for a given method by
// comparing the Rules's Selector length. E.g. for the methodName
// "com.example.Hello", the selector relevance is:
// "com.example.Hello" > "com.example.*" > "*"
//
// If none of the Rules are relevant to the given method (i.e. the
// the selectors don't match), we return nil.
func mostRelevant(methodName string, rules []Rule) Rule {
	var longest int
	var mostRelevant Rule
	for _, r := range rules {
		if r.isApplicable(methodName) && len(r.Selector) > longest {
			longest = len(r.Selector)
			mostRelevant = r
			if longest == len(methodName) {
				// Return immediately if the methodName
				// is the same length as the selector.
				return mostRelevant
			}
		}
	}
	return mostRelevant
}
