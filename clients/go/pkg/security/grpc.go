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

// RequestPrincipal extracts the JWT principal from the grpcmetadata in the context.
// This method does not verify the JWT.
func (j *FromRawJWT) RequestPrincipal(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("gRPC metadata in incoming context is missing")
	}

	// Extract the JWT.
	if j == nil || j.FromRawJWT == nil {
		return "", fmt.Errorf(`expecting non-nil receiver field "j.FromRawJwt"`)
	}
	vals, ok := md[j.FromRawJWT.Key]
	if !ok {
		return "", fmt.Errorf("failed extracting the JWT due missing key %q in grpc metadata %+v", j.FromRawJWT.Key, md)
	}
	if len(vals) != 1 {
		return "", fmt.Errorf("expecting exaclty one value (a JWT) under key %q in grpc metadata %+v", j.FromRawJWT.Key, md)
	}
	jwtRaw := vals[0]
	if len(j.FromRawJWT.Prefix) > len(jwtRaw) {
		return "", fmt.Errorf("JWT prefix %q is longer than raw JWT %q", j.FromRawJWT.Prefix, jwtRaw)
	}
	idToken := jwtRaw[len(j.FromRawJWT.Prefix):] // trim prefix
	if idToken == "" {
		return "", fmt.Errorf("nil JWT ID token under the key %q in grpc metadata %+v", j.FromRawJWT.Key, md)
	}

	// Parse the JWT into claims.
	p := &jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err := p.ParseUnverified(idToken, claims)
	if err != nil {
		return "", fmt.Errorf("unable to parse JWT: %w", err)
	}

	// Retrieve the principal from claims.
	principalRaw, ok := claims[emailKey]
	if !ok {
		return "", fmt.Errorf("jwt claims are missing the email key %q", emailKey)
	}
	principal, ok := principalRaw.(string)
	if !ok {
		return "", fmt.Errorf("expecting string in jwt claims %q, got %T", emailKey, principalRaw)
	}
	if principal == "" {
		return "", fmt.Errorf("nil principal under claims %q", emailKey)
	}

	return principal, nil
}
