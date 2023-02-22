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

package remote

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/impersonate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// WithIDTokenAuth sets up the processor to do auth with ID token.
// TODO(b/201541513): It's not clear how to unit test this functionality.
// We can at least cover it in the integration test.
func WithIDTokenAuth(ctx context.Context) Option {
	return func(p *Processor) error {
		ts, err := idtoken.NewTokenSource(ctx, "https://"+strings.TrimSuffix(p.address, ":443"))
		if err != nil {
			return fmt.Errorf("failed idtoken.NewTokenSource: %w", err)
		}
		return setIDTokenAuth(p, ts)
	}
}

// WithImpersonatedIDTokenAuth sets up the processor to do auth with impersonated
// service account ID token.
func WithImpersonatedIDTokenAuth(ctx context.Context, targetPrincipal string) Option {
	return func(p *Processor) error {
		ts, err := impersonate.IDTokenSource(ctx, impersonate.IDTokenConfig{
			TargetPrincipal: targetPrincipal,
			Audience:        "https://" + strings.TrimSuffix(p.address, ":443"),
			IncludeEmail:    true,
		})
		if err != nil {
			return fmt.Errorf("failed impersonate.IDTokenSource: %w", err)
		}
		return setIDTokenAuth(p, ts)
	}
}

func setIDTokenAuth(p *Processor, ts oauth2.TokenSource) error {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("failed to load system cert pool: %w", err)
	}

	//nolint:gosec // We need to support TLS 1.2 for now (G402).
	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})

	p.authOpts = &idTokenAuth{
		dialOptions: []grpc.DialOption{grpc.WithTransportCredentials(cred)},

		// Persist a token source to reuse tokens.
		tokenSource: ts,
	}
	return nil
}

type idTokenAuth struct {
	dialOptions []grpc.DialOption
	tokenSource oauth2.TokenSource
}

func (a *idTokenAuth) dialOpts() ([]grpc.DialOption, error) {
	return a.dialOptions, nil
}

func (a *idTokenAuth) callOpts() ([]grpc.CallOption, error) {
	token, err := a.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id token: %w", err)
	}

	rpcCreds := oauth.TokenSource{
		TokenSource: oauth2.StaticTokenSource(token),
	}

	return []grpc.CallOption{grpc.PerRPCCredentials(rpcCreds)}, nil
}
