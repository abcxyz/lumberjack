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

// A Rule tells the middleware which methods should be logged and how
// they should be logged.
type Rule struct {
	Selector  string
	Directive string
	LogType   alpb.AuditLogRequest_LogType
}

// isApplicable determines if the a Rule's selector applies to
// the given method.
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
