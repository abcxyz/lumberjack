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

// Package auditopt configures a client by looking three locations
// to determine the config variables:
//   1. env vars
//   2. a config file
//   3. defaults
// Each location takes precedence on the one below it. For example,
// if you set the config variable `FOO` in both env vars and a config
// file, the env var value overwrites the config file value.
package auditopt

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"

	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/filtering"
	"github.com/abcxyz/lumberjack/clients/go/pkg/remote"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/sethvargo/go-envconfig"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

const defaultConfigFilePath = "/etc/auditlogging/config.yaml"

// MustFromConfigFile specifies a config file to configure the
// audit client. `path` is required, and if the config file is
// missing, we return an error.
func MustFromConfigFile(path string) audit.Option {
	return func(c *audit.Client) error {
		fc, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		cfg := &alpb.Config{}
		if err := yaml.Unmarshal(fc, cfg); err != nil {
			return err
		}
		if err := loadEnvAndValidateCfg(cfg); err != nil {
			return err
		}
		return clientFromConfig(c, cfg)
	}
}

// FromConfigFile specifies a config file to configure the
// audit client. If `path` is nil, we use a default well known
// path. If the config file is not found, we keep going by
// using env vars and default values to configure the client.
func FromConfigFile(path string) audit.Option {
	return func(c *audit.Client) error {
		if path == "" {
			path = defaultConfigFilePath
		}
		fc, err := ioutil.ReadFile(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			// We don't return an error if the file is not found because we
			// still use env vars and defaults to setup the client.
			return err
		}
		cfg := &alpb.Config{}
		if err := yaml.Unmarshal(fc, cfg); err != nil {
			return err
		}
		if err := loadEnvAndValidateCfg(cfg); err != nil {
			return err
		}
		return clientFromConfig(c, cfg)
	}
}

// WithInterceptorFromConfigFile returns a gRPC server option that adds a unary interceptor
// to a gRPC server. This interceptor autofills and emits audit logs for gRPC unary
// calls. WithInterceptorFromConfigFile also returns the audit client that the interceptor
// uses. This allows the caller to close the client when shutting down the gRPC server.
// For example:
// ```
// opt, c, err := audit.WithInterceptorFromConfigFile("auditconfig.yaml")
// if err != nil {
//	log.Fatalf(err)
// }
// defer c.Stop()
// s := grpc.NewServer(opt)
// ```
// TODO(#109): add streaming interceptor.
func WithInterceptorFromConfigFile(path string) (grpc.ServerOption, *audit.Client, error) {
	fc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	cfg := &alpb.Config{}
	if err := yaml.Unmarshal(fc, cfg); err != nil {
		return nil, nil, err
	}
	if err := cfg.ValidateSecurityContext(); err != nil {
		return nil, nil, err
	}
	if err := loadEnvAndValidateCfg(cfg); err != nil {
		return nil, nil, err
	}

	interceptor := &audit.Interceptor{}
	// Add security context to interceptor.
	switch {
	case cfg.SecurityContext.FromRawJWT != nil:
		fromRawJWT := &security.FromRawJWT{
			FromRawJWT: cfg.SecurityContext.FromRawJWT,
		}
		interceptor.SecurityContext = fromRawJWT
	default:
		return nil, nil, fmt.Errorf("no supported security context configured in config %+v", cfg)
	}

	// Add audit rules to interceptor.
	interceptor.Rules = cfg.Rules

	// Add audit client to interceptor.
	auditOpt := func(c *audit.Client) error {
		return clientFromConfig(c, cfg)
	}
	auditClient, err := audit.NewClient(auditOpt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create audit client from config %+v: %w", cfg, err)
	}
	interceptor.Client = auditClient

	return grpc.UnaryInterceptor(interceptor.UnaryInterceptor), auditClient, nil
}

func clientFromConfig(c *audit.Client, cfg *alpb.Config) error {
	opts := []audit.Option{audit.WithRuntimeInfo()}

	withPrincipalFilter, err := principalFilterFromConfig(cfg)
	if err != nil {
		return err
	}
	opts = append(opts, withPrincipalFilter)

	withBackend, err := backendFromConfig(cfg)
	if err != nil {
		return err
	}
	opts = append(opts, withBackend)

	for _, o := range opts {
		if err := o(c); err != nil {
			return err
		}
	}
	return nil
}

func principalFilterFromConfig(cfg *alpb.Config) (audit.Option, error) {
	var opts []filtering.Option
	// Nil `PrincipalInclude` and `PrincipalExclude` is fine because
	// calling `filtering.WithIncludes("")` is a noop.
	withIncludes := filtering.WithIncludes(cfg.Condition.Regex.PrincipalInclude)
	withExcludes := filtering.WithExcludes(cfg.Condition.Regex.PrincipalExclude)
	opts = append(opts, withIncludes, withExcludes)
	m, err := filtering.NewPrincipalEmailMatcher(opts...)
	if err != nil {
		return nil, err
	}
	return audit.WithValidator(m), nil
}

func backendFromConfig(cfg *alpb.Config) (audit.Option, error) {
	// TODO(#74): Fall back to stdout logging if address is missing.
	addr := cfg.Backend.Address
	authopts := []remote.Option{}
	if !cfg.Backend.InsecureEnabled {
		impersonate := cfg.Backend.ImpersonateAccount
		if impersonate == "" {
			authopts = append(authopts, remote.WithDefaultAuth())
		} else {
			authopts = append(authopts, remote.WithImpersonatedIDTokenAuth(context.Background(), impersonate))
		}
	}
	b, err := remote.NewProcessor(addr, authopts...)
	if err != nil {
		return nil, err
	}
	return audit.WithBackend(b), nil
}

func loadEnvAndValidateCfg(cfg *alpb.Config) error {
	l := envconfig.PrefixLookuper("AUDIT_CLIENT_", envconfig.OsLookuper())
	if err := envconfig.ProcessWith(context.Background(), cfg, l); err != nil {
		return err
	}

	cfg.SetDefault()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("failed validating config %+v: %w", cfg, err)
	}

	return nil
}
