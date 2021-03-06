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

package audit

import (
	"fmt"
)

var (
	// ErrInvalidRequest is the (base) error to return when a log processor
	// considers a log request is invalid.
	ErrInvalidRequest = fmt.Errorf("invalid audit log request")

	// ErrFailedPrecondition is the (base) error to return when a log processor
	// considers a log request should not continue to be processed by any remaining
	// log processors. The audit client will not return this type of errors.
	ErrFailedPrecondition = fmt.Errorf("failed precondition")
)
