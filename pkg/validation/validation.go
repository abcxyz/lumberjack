// Copyright 2023 The Authors (see AUTHORS file)
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

// Package validation provides utils for lumberjack/data access logs validation.
package validation

import (
	"errors"
	"fmt"
	"strings"

	lepb "cloud.google.com/go/logging/apiv2/loggingpb"
	cal "google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/encoding/protojson"
)

var requiredLabels = map[string]struct{}{
	"environment":            {},
	"accessing_process_name": {},
}

// Validator validates a lumberjack log entry.
type Validator func(le *lepb.LogEntry) error

// Validate validates a json string representation of a lumberjack log.
func Validate(log string, extra ...Validator) error {
	var logEntry lepb.LogEntry
	if err := protojson.Unmarshal([]byte(log), &logEntry); err != nil {
		return fmt.Errorf("failed to parse log entry as JSON: %w", err)
	}

	var retErr error
	for _, v := range append([]Validator{validatePayload}, extra...) {
		retErr = errors.Join(retErr, v(&logEntry))
	}
	return retErr
}

// ValidateLabels checks required lumberjack labels.
func ValidateLabels(le *lepb.LogEntry) error {
	if le.Labels == nil {
		return fmt.Errorf("missing labels")
	}

	var retErr error
	for k := range requiredLabels {
		if _, ok := le.GetLabels()[k]; !ok {
			retErr = errors.Join(retErr, fmt.Errorf("missing required label: %q", k))
		}
	}
	return retErr
}

// Required audit log payload check for lumberjack logs.
func validatePayload(logEntry *lepb.LogEntry) error {
	payload := logEntry.GetJsonPayload()
	if payload == nil {
		return fmt.Errorf("missing audit log payload")
	}

	var al cal.AuditLog
	val, err := payload.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to extract audit log from JSON payload: %w", err)
	}
	if err := protojson.Unmarshal(val, &al); err != nil {
		return fmt.Errorf("failed to parse JSON payload: %w", err)
	}
	if err := ValidateAuditLog(&al); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	return nil
}

// ValidateAuditLog validates the audit log payload for lumberjack.
func ValidateAuditLog(payload *cal.AuditLog) error {
	if payload == nil {
		return fmt.Errorf("audit log payload cannot be nil")
	}

	var retErr error
	if payload.GetMethodName() == "" {
		retErr = errors.Join(retErr, fmt.Errorf("MethodName cannot be empty"))
	}

	if payload.GetServiceName() == "" {
		retErr = errors.Join(retErr, fmt.Errorf("ServiceName cannot be empty"))
	}

	if payload.GetResourceName() == "" {
		retErr = errors.Join(retErr, fmt.Errorf("ResourceName cannot be empty"))
	}

	if payload.GetAuthenticationInfo() == nil {
		retErr = errors.Join(retErr, fmt.Errorf("AuthenticationInfo cannot be nil"))
	} else {
		email := payload.GetAuthenticationInfo().GetPrincipalEmail()
		if err := validateEmail(email); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}

	return retErr
}

// This method is intended to validate that the email associated with the
// authentication request has the correct format and in a valid domain.
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("PrincipalEmail cannot be empty")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] == "" {
		return fmt.Errorf("PrincipalEmail %q is malformed", email)
	}
	return nil
}
