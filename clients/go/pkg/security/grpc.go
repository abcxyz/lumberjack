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
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwt"
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
	FromRawJWT []*v1alpha1.FromRawJWT
}

// RequestPrincipal extracts the JWT principal from the grpcmetadata in the
// context. This method does not verify the JWT.
func (j *FromRawJWT) RequestPrincipal(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("gRPC metadata in incoming context is missing")
	}

	idToken, err := j.findJWT(md)
	if err != nil {
		return "", err
	}

	token, err := jwt.ParseString(idToken, jwt.WithVerify(false))
	if err != nil {
		return "", fmt.Errorf("failed to parse jwt: %w", err)
	}

	// Extract the email claim.
	emailRaw, ok := token.Get(emailKey)
	if !ok {
		return "", fmt.Errorf("missing claim %q", emailKey)
	}
	email, ok := emailRaw.(string)
	if !ok {
		return "", fmt.Errorf("claim %q is not of type %T (got %T)", emailKey, "", emailRaw)
	}
	if email == "" {
		return "", fmt.Errorf("claim %q cannot be blank", emailKey)
	}

	return email, nil
}

// findJWT looks for a JWT from the gRPC metadata that matches the rules.
func (j *FromRawJWT) findJWT(md grpcmetadata.MD) (string, error) {
	for _, fj := range j.FromRawJWT {
		// Keys in grpc metadata are all lowercases.
		vals := md.Get(fj.Key)
		if len(vals) == 0 {
			continue
		}
		jwtRaw := vals[0]
		// We compare prefix case insensitively.
		if !strings.HasPrefix(strings.ToLower(jwtRaw), strings.ToLower(fj.Prefix)) {
			continue
		}
		idToken := jwtRaw[len(fj.Prefix):]
		return idToken, nil
	}

	return "", fmt.Errorf("no JWT found matching rules: %#v", j.FromRawJWT)
}
