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

// Package filtering provides a processor to filter audit log requests.
package securitycontext

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt"
	grpcmetadata "google.golang.org/grpc/metadata"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

type securityContext struct {
	authnMethod authnMethod
}

type authnMethod interface {
	extractPrincipal(context.Context)
}

type fromRawJWT struct {
	key    string
	prefix string
}

func (rawJWT *fromRawJWT) extractPrincipal(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("cannot extract principal because of missing gRPC incoming context")
	}

	// Extract the JWT from the given key and prefix.
	var idToken string
	if auths := md[rawJWT.key]; len(auths) > 0 {
		idToken = auths[0][len(rawJWT.prefix):] // trim prefix
	}
	if idToken == "" {
		return "", fmt.Errorf("cannot extract principal because JWT id token under the key `authorization` is nil: +%v", md)
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
		return "", fmt.Errorf("cannot extract principal because it's nil")
	}

	return principal, nil
}

func NewSecurityContext() (*securityContext, error) {
	return &securityContext{}, nil
}

// Process is a mutator that fills out the principal email.
func (sc *securityContext) Process(ctx context.Context, logReq *alpb.AuditLogRequest) error {
	principal, err := principalFromContext(ctx)
	if err != nil {
		return err
	}

	if principal == "" {
		return fmt.Errorf("unable to autofill principal")
	}
	return nil
}

// principalFromContext extracts the principal from the context.
// This method assumes that a JWT exists the grpcmetadata under
// the key `authorization` and with prefix `Bearer `. If that's
// not the case, we return an error.
func principalFromContext(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("cannot extract principal because of missing gRPC incoming context")
	}

	// Extract the JWT.
	var idToken string
	if auths := md["authorization"]; len(auths) > 0 {
		idToken = auths[0][7:] // trim "Bearer: " prefix
	}
	if idToken == "" {
		return "", fmt.Errorf("cannot extract principal because JWT under the key `authorization` is nil: +%v", md)
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
		return "", fmt.Errorf("cannot extract principal because it's nil")
	}

	return principal, nil
}
