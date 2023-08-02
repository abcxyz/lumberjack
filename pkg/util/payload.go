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

// Package util provides untils for audit log payload validation.
package util

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/cloud/audit"
)

// Validate validates the audit log payload for lumberjack.
func Validate(payload *audit.AuditLog) error {
	if payload == nil {
		return fmt.Errorf("audit log payload cannot be nil")
	}

	if payload.MethodName == "" {
		return fmt.Errorf("MethodName cannot be empty")
	}

	if payload.ServiceName == "" {
		return fmt.Errorf("ServiceName cannot be empty")
	}

	if payload.ResourceName == "" {
		return fmt.Errorf("ResourceName cannot be empty")
	}

	if payload.AuthenticationInfo == nil {
		return fmt.Errorf("AuthenticationInfo cannot be nil")
	}

	email := payload.AuthenticationInfo.PrincipalEmail
	if err := validateEmail(email); err != nil {
		return err
	}

	return nil
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
