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

// Package errutil provides util functions for errors.
package errutil

import (
	"fmt"
	"strings"
)

// DiffSubstring checks if given error's message contains the given string.
// If so, returns an empty string. Otherwise, returns a diff string.
func DiffSubstring(got error, want string) string {
	if want == "" {
		if got == nil {
			return ""
		}
		return fmt.Sprintf("got err=%v, want err=nil", got)
	}
	if got == nil {
		return fmt.Sprintf("got err=nil, want err containing %q", want)
	}
	if !strings.Contains(got.Error(), want) {
		return fmt.Sprintf("got err=%v, want err containing %q", got, want)
	}
	return ""
}
