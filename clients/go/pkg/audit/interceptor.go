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
	"context"
	"fmt"

	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"google.golang.org/grpc"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

// Interceptor contains the fields required for an interceptor
// to autofill and emit audit logs.
type Interceptor struct {
	*Client
	SecurityContext security.GRPCContext
	Rules           []Rule
}

type SecurityContext interface {
	extractPrincipal(context.Context) (string, error)
}

type FromRawJWT struct {
	Key    string
	Prefix string
	//TODO(noamrabbani): Add JWKS fields to validate JWT signature.
}

type Rule struct {
	Selector  string
	Directive Directive
	LogType   alpb.AuditLogRequest_LogType
}

type Directive string

const (
	AuditRequestAndResponse Directive = "AUDIT_REQUEST_AND_RESPONSE"
	AuditRequestOnly        Directive = "AUDIT_REQUEST_ONLY"
	AuditOnly               Directive = "AUDIT"
)

func (rawJWT *FromRawJWT) extractPrincipal(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not yet implemented")
}

// UnaryInterceptor is a unary interceptor that autofills and emits audit logs.
// TODO(noamrabbani): implement unary interceptor.
func (i *Interceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return nil, fmt.Errorf("not yet implemented")
}
