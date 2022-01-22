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

// Package security describes the authentication technology that the
// middleware investigates to autofill the principal in a log request.
package security

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt"
	grpcmetadata "google.golang.org/grpc/metadata"

	"github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

// Key in a JWT's `claims` where we expect the principal.
const emailKey = "email"

// GRPCContext is an interface that retrieves the principal
// from a gRPC security context. A gRPC security context describes
// the technology used to authenticate a principal (e.g. JWT).
type GRPCContext interface {
	RequestPrincipal(context.Context) (string, error)
}

// FromRawJWT contains the information needed to retrieve
// the principal from a raw JWT.
type FromRawJWT struct {
	FromRawJWT *v1alpha1.FromRawJWT
}

// principalFromContext extracts the principal from the context.
func (j *FromRawJWT) RequestPrincipal(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("gRPC incoming context is missing")
	}

	// Extract the JWT.
	var idToken string
	if auths := md[j.FromRawJWT.Key]; len(auths) > 0 {
		idToken = auths[0][len(j.FromRawJWT.Prefix):] // trim prefix
	}
	if idToken == "" {
		return "", fmt.Errorf("nil JWT under the key %q in metadata %+v", j.FromRawJWT.Key, md)
	}

	// Retrieve the principal from the JWT.
	p := &jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err := p.ParseUnverified(idToken, claims)
	if err != nil {
		return "", fmt.Errorf("unable to parse JWT: %w", err)
	}
	principal := claims["email"].(string)
	if principal == "" {
		return "", fmt.Errorf("nil principal under claims %q", emailKey)
	}

	return principal, nil
}
