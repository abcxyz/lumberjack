// Copyright 2023 Lumberjack authors (see AUTHORS file)
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

// Package auditerrors defines the sentinel errors for the project.
package auditerrors

// Error is a concrete error implementation.
type Error string

// Error satisfies the error interface.
func (e Error) Error() string {
	return string(e)
}

const (
	// ErrInvalidRequest is the (base) error to return when a log processor
	// considers a log request is invalid.
	ErrInvalidRequest = Error("invalid audit log request")

	// ErrPreconditionFailed is the (base) error to return when a log processor
	// considers a log request should not continue to be processed by any
	// remaining log processors. The audit client will not return this type of
	// errors.
	ErrPreconditionFailed = Error("precondition failed")

	// ErrLogRequestMissingPayload is the error returned when the log request does
	// not have a payload set.
	ErrLogRequestMissingPayload = Error("missing payload in log request")

	// ErrJustificationMissing is the error returned with a justification token
	// was
	// not provided with the request.
	ErrJustificationMissing = Error("missing justification token")

	// ErrJustificationInvalid is the error returned when a justification token
	// was present, but failed validation.
	ErrJustificationInvalid = Error("invalid justification token")
)
