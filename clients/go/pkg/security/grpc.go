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
)

// GRPCContext is an interface that retrieves the principal
// from a gRPC security context. A gRPC security context describes
// the technology used to authenticate a principal (e.g. JWT).
type GRPCContext interface {
	RequestPrincipal(context.Context) (string, error)
}

// FromRawJWT contains the information needed to retrieve
// the principal from a raw JWT. More specifically:
//   - `Key` is the grpcmetadata key that contains the JWT.
//   - `Prefix` is the prefix that should be be stripped before decoding the JWT.
type FromRawJWT struct {
	Key    string
	Prefix string
	//TODO: Add JWKS fields to validate JWT signature.
}

// principalFromContext extracts the principal from the context.
func (j *FromRawJWT) RequestPrincipal(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("gRPC incoming context is missing")
	}

	// Extract the JWT.
	var idToken string
	if auths := md[j.Key]; len(auths) > 0 {
		idToken = auths[0][len(j.Prefix):] // trim prefix
	}
	if idToken == "" {
		return "", fmt.Errorf("nil JWT under the key %q in metadata %+v", j.Key, md)
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
		return "", fmt.Errorf(`nil principal under claims "email"`)
	}

	return principal, nil
}
