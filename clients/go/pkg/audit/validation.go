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
	"context"
	"fmt"
	"strings"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

type requestValidation struct{}

// Process with receiver auditLogRequestValidation verifies
// that the AuditLogRequest is properly formed.
func (p requestValidation) Process(_ context.Context, logReq *alpb.AuditLogRequest) error {
	if logReq == nil {
		return fmt.Errorf("request shouldn't be nil: %w", ErrInvalidRequest)
	}
	if logReq.Payload == nil {
		return fmt.Errorf("request.Payload shouldn't be nil: %w", ErrInvalidRequest)
	}

	if logReq.Payload.ServiceName == "" {
		return fmt.Errorf("ServiceName shouldn't be empty: %w", ErrInvalidRequest)
	}

	if logReq.Payload.AuthenticationInfo == nil {
		return fmt.Errorf("AuthenticationInfo shouldn't be nil: %w", ErrInvalidRequest)
	}

	email := logReq.Payload.AuthenticationInfo.PrincipalEmail
	if err := p.validateEmail(email); err != nil {
		return err
	}

	return nil
}

// This method is intended to validate that the email associated with
// the authentication request has the correct format and in a valid
// domain.
func (p requestValidation) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("PrincipalEmail shouldn't be nil: %w", ErrInvalidRequest)
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] == "" {
		return fmt.Errorf("email domain malformed: %w", ErrInvalidRequest)
	}
	return nil
}
