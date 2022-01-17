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
	Key    string `yaml:"key"`
	Prefix string `yaml:"prefix"`
	//TODO: Add JWKS fields to validate JWT signature.
}

// RequestPrincipal returns the principal in a raw JWT.
func (rawJWT *FromRawJWT) RequestPrincipal(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not yet implemented")
}
