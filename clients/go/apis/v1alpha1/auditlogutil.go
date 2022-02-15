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

// Package util provides utilities for the audit logger.
package v1alpha1

// ShouldFailClose returns true only if FAIL_CLOSE is explicitly configured. On BEST_EFFORT or LOG_MODE_UNSPECIFIED
// (the default) then return false.
func ShouldFailClose(logMode AuditLogRequest_LogMode) bool {
	return logMode == AuditLogRequest_FAIL_CLOSE
}
