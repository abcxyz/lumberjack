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

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/auditerrors"
)

// RequestValidator validates log request fields.
type RequestValidator struct{}

// NewRequestValidator returns a new validtor that processes log request fields.
func NewRequestValidator(ctx context.Context) *RequestValidator {
	return &RequestValidator{}
}

// Process with receiver auditLogRequestValidation verifies that the
// AuditLogRequest is properly formed.
func (p *RequestValidator) Process(ctx context.Context, logReq *api.AuditLogRequest) error {
	if err := p.process(ctx, logReq); err != nil {
		return fmt.Errorf("%w: %w", auditerrors.ErrInvalidRequest, err)
	}
	return nil
}

// process is like Process, but allows for easier error wrapping.
func (p *RequestValidator) process(ctx context.Context, logReq *api.AuditLogRequest) error {
	if logReq == nil {
		return fmt.Errorf("AuditLogRequest cannot be nil")
	}

	if logReq.Payload == nil {
		return fmt.Errorf("AuditLogRequest.Payload cannot be nil")
	}

	if logReq.Payload.ServiceName == "" {
		return fmt.Errorf("ServiceName cannot be empty")
	}

	if logReq.Payload.AuthenticationInfo == nil {
		return fmt.Errorf("AuthenticationInfo cannot be nil")
	}

	email := logReq.Payload.AuthenticationInfo.PrincipalEmail
	if err := p.validateEmail(email); err != nil {
		return err
	}

	return nil
}

// This method is intended to validate that the email associated with the
// authentication request has the correct format and in a valid domain.
func (p *RequestValidator) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("PrincipalEmail cannot be empty")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] == "" {
		return fmt.Errorf("PrincipalEmail %q is malformed", email)
	}
	return nil
}
