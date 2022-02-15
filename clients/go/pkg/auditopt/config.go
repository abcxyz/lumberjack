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
	"gopkg.in/yaml.v2"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

const DefaultConfigFilePath = "/etc/auditlogging/config.yaml"

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
		if err := setAndValidate(cfg); err != nil {
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
			path = DefaultConfigFilePath
		}
		fc, err := ioutil.ReadFile(path)
		// We ignore ErrNotExist when reading the file because we
		// still use env vars and defaults to setup the client.
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		cfg := &alpb.Config{}
		if err := yaml.Unmarshal(fc, cfg); err != nil {
			return err
		}
		if err := setAndValidate(cfg); err != nil {
			return err
		}
		return clientFromConfig(c, cfg)
	}
}

// WithInterceptorFromConfigFile returns a gRPC server option that adds a unary interceptor
// to a gRPC server. This interceptor autofills and emits audit logs for gRPC unary
// calls. WithInterceptorFromConfigFile also returns the audit client that the interceptor
// uses. This allows the caller to close the client when shutting down the gRPC server.
// TODO(#152): Refactor this to separate option loading from construction of the interceptor.
func WithInterceptorFromConfigFile(path string) (*audit.Interceptor, error) {
	fc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	cfg := &alpb.Config{}
	if err := yaml.Unmarshal(fc, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file to yaml: %w", err)
	}
	if err := cfg.ValidateSecurityContext(); err != nil {
		return nil, err
	}
	if err := setAndValidate(cfg); err != nil {
		return nil, err
	}

	interceptor := &audit.Interceptor{}
	// Add security context to interceptor.
	switch {
	case cfg.SecurityContext.FromRawJWT != nil:
		fromRawJWT := &security.FromRawJWT{
			FromRawJWT: cfg.SecurityContext.FromRawJWT,
		}
		interceptor.SecurityContext = fromRawJWT
		interceptor.LogMode = cfg.GetLogMode()
	default:
		return nil, fmt.Errorf("no supported security context configured in config %+v", cfg)
	}

	// Add audit rules to interceptor.
	interceptor.Rules = cfg.Rules

	// Add audit client to interceptor.
	auditOpt := func(c *audit.Client) error {
		return clientFromConfig(c, cfg)
	}
	auditClient, err := audit.NewClient(auditOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit client from config %+v: %w", cfg, err)
	}
	interceptor.Client = auditClient

	return interceptor, nil
}

func clientFromConfig(c *audit.Client, cfg *alpb.Config) error {
	opts := []audit.Option{audit.WithRuntimeInfo()}

	withPrincipalFilter, err := principalFilterFromConfig(cfg)
	if err != nil {
		return err
	}
	if withPrincipalFilter != nil {
		opts = append(opts, withPrincipalFilter)
	}

	withBackend, err := backendFromConfig(cfg)
	if err != nil {
		return err
	}
	opts = append(opts, withBackend)

	withLabels := labelsFromConfig(cfg)
	opts = append(opts, withLabels)

	for _, o := range opts {
		if err := o(c); err != nil {
			return err
		}
	}
	return nil
}

func principalFilterFromConfig(cfg *alpb.Config) (audit.Option, error) {
	if cfg.Condition == nil || cfg.Condition.Regex == nil {
		return nil, nil
	}
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

func labelsFromConfig(cfg *alpb.Config) audit.Option {
	lp := audit.LabelProcessor{DefaultLabels: cfg.Labels}
	return audit.WithMutator(&lp)
}

// setAndValidate sets cfg values from env vars and defaults. Additionally,
// we validate the cfg values.
func setAndValidate(cfg *alpb.Config) error {
	// TODO(#123): envconfig.ProcessWith(...) traverses the cfg struct by creating
	// creating non-nil values for pointers to the nested fields. For example,
	// after a traversal, the value of `cfg.security_context.from_raw_jwt.key`
	// will be the empty string "" even if `cfg.security_context` was previously
	// unset and that `cfg.security_context.from_raw_jwt.key` cannot be set by env
	// vars. As a result, logic relying on nil values in `cfg` will be problematic.
	// For example, validating that `cfg.security_context` is set should be done
	// before calling `envconfig.ProcessWith(...)` because `envconfig.ProcessWith(...)`
	// sets the value of `cfg.security_context` to a non-nil value `from_raw_jwt`.
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
