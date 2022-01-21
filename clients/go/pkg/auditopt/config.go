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
	"strings"

	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/filtering"
	"github.com/abcxyz/lumberjack/clients/go/pkg/remote"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/abcxyz/lumberjack/clients/go/pkg/zlogger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

// The version we expect in a config file.
const expectedVersion = "v1alpha1"

const defaultConfigFilePath = "/etc/auditlogging/config.yaml"

// MustFromConfigFile specifies a config file to configure the
// audit client. `path` is required, and if the config file is
// missing, we return an error.
func MustFromConfigFile(path string) audit.Option {
	return func(c *audit.Client) error {
		v := viper.New()
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return err
		}
		cfg, err := configFromViper(v)
		if err != nil {
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
		v := viper.New()
		if path == "" {
			path = defaultConfigFilePath
		}
		// We don't return an error if the file is not found because we
		// still use env vars and defaults to setup the client.
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		cfg, err := configFromViper(v)
		if err != nil {
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
// TODO(noamrabbani): add streaming interceptor.
func WithInterceptorFromConfigFile(path string) (grpc.ServerOption, *audit.Client, error) {
	// Prepare Viper config.
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, nil, err
	}
	cfg, err := configFromViper(v)
	if err != nil {
		return nil, nil, err
	}

	// Create security context from config.
	if cfg.SecurityContext == nil {
		return nil, nil, fmt.Errorf("no supported security context configured in config %+v", cfg)
	}
	interceptor := &audit.Interceptor{}
	switch {
	case cfg.SecurityContext.FromRawJWT != nil:
		fromRawJWT, err := fromRawJWTFromConfig(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting `from_raw_jwt` in config %+v: %w", cfg, err)
		}
		interceptor.SecurityContext = fromRawJWT
	default:
		return nil, nil, fmt.Errorf("no supported security context configured in config %+v", cfg)
	}

	// Create audit rules from config.
	// todo: default+validate the whole config instead of individual fields
	for _, r := range cfg.Rules {
		r.SetDefault()
		if err := r.Validate(); err != nil {
			return nil, nil, fmt.Errorf("failed validating config rule %+v: %w", r, err)
		}
	}
	interceptor.Rules = cfg.Rules

	// Create audit client from config.
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
	if cfg.Condition != nil && cfg.Condition.Regex != nil {
		// Nil `PrincipalInclude` and `PrincipalExclude` is fine because
		// calling `filtering.WithIncludes("")` is a noop.
		withIncludes := filtering.WithIncludes(cfg.Condition.Regex.PrincipalInclude)
		withExcludes := filtering.WithExcludes(cfg.Condition.Regex.PrincipalExclude)
		opts = append(opts, withIncludes, withExcludes)
	}
	m, err := filtering.NewPrincipalEmailMatcher(opts...)
	if err != nil {
		return nil, err
	}
	return audit.WithValidator(m), nil
}

func backendFromConfig(cfg *alpb.Config) (audit.Option, error) {
	// TODO(#74): Fall back to stdout logging if address is missing.
	if cfg.Backend == nil || cfg.Backend.Address == "" {
		return nil, fmt.Errorf("backend address in the config is nil, set it as an env var or in a config file")
	}
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

// fromRawJWTFromConfig populates a `*security.FromRawJWT` from a config.
// We handle nil and unset values in the following way:
//
// security_context:
// # -> no defaulting because security_context is nil/unset
//
// security_context:
//   from_raw_jwt:
// # -> default values for `from_raw_jwt`
//
// security_context:
//   from_raw_jwt: {}
// # -> default values for `from_raw_jwt`
//
// security_context:
//   from_raw_jwt:
//     key: x-jwt-assertion
//     prefix:
// # -> no defaulting because the user specified values for `from_raw_jwt`
//
// security_context:
//   from_raw_jwt:
//     key: x-jwt-assertion
//     prefix: ""
// # -> no defaulting because the user specified one value for `from_raw_jwt`
//
// security_context:
//   from_raw_jwt:
//     key: x-jwt-assertion
// # -> no defaulting because the user specified one value for `from_raw_jwt`
//
// TODO(#73): add support for lists in `security_context`
func fromRawJWTFromConfig(cfg *alpb.Config) (*security.FromRawJWT, error) {
	//TODO: do validation+defaulting on the whole config instead of individual fields
	if cfg.SecurityContext == nil {
		return nil, fmt.Errorf("fromRawJWT in the config is nil, set it as an env var or in a config file")
	}
	if err := cfg.SecurityContext.Validate(); err != nil {
		return nil, err
	}
	cfg.SecurityContext.FromRawJWT.SetDefault()
	return &security.FromRawJWT{
		FromRawJWT: cfg.SecurityContext.FromRawJWT,
	}, nil
}

func configFromViper(v *viper.Viper) (*alpb.Config, error) {
	logger := zlogger.Default()
	v = setDefaultValues(v)
	v = bindEnvVars(v)
	config := &alpb.Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal viper into config struct: %w", err)
	}

	cfgVersion := config.Version
	if cfgVersion == "" {
		logger.Warnf("config version is unset, set your config to the supported version %q", expectedVersion)
	} else if cfgVersion != expectedVersion {
		return nil, fmt.Errorf("explicitly specified config version %q unsupported, supported version is %q", cfgVersion, expectedVersion)
	}
	return config, nil
}

func setDefaultValues(v *viper.Viper) *viper.Viper {
	// By default, we filter log requests that have an IAM
	// service account as the principal.
	// TODO: do validation+defaulting on the whole config instead of individual fields.
	v.SetDefault("condition.regex.principal_exclude", ".iam.gserviceaccount.com$")

	// Set default value for the security context. This
	// enables the following config file behaviours:
	//
	// security_context:
	// # -> no defaulting because security_context is nil/unset
	//
	// security_context:
	//   from_raw_jwt:
	// # -> default values for `from_raw_jwt`
	//
	// security_context:
	//   from_raw_jwt: {}
	// # -> default values for `from_raw_jwt`
	v.SetDefault("security_context", nil)
	sc := v.GetStringMap("security_context")
	if _, ok := sc["from_raw_jwt"]; ok {
		v.SetDefault("security_context.from_raw_jwt", map[string]string{})
	}

	return v
}

// bindEnvVars associates env vars with Viper config variables. This
// allows the user to overwrite a config file value or a default value
// by setting an env var. Note that:
//   - Env vars are prefixed with "AUDIT_CLIENT_".
//   - Only leaf config variables can be overwritten with env vars.
//   - Env vars cannot overwrite list config variables, such as `rules`.
//   - Nested config variables are set from env vars by replacing
//     the "." delimiter with "_". E.g. the config variable "backend.address"
//     is set with the env var "AUDIT_CLIENT_BACKEND_ADDRESS".
func bindEnvVars(v *viper.Viper) *viper.Viper {
	v.SetEnvPrefix("AUDIT_CLIENT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	for _, lk := range v1alpha1.LeafKeys() {
		// We don't use v.AutomaticEnv() because it fails to bind env vars
		// when they are inexistent in the config and when they don't have
		// an explicit default value.
		mustBindEnv(v, lk)
	}
	return v
}

func mustBindEnv(v *viper.Viper, env string) {
	// BindEnv only returns an error when the input _slice_ to the variadic
	// input is empty. Since this function signature requires a string value,
	// we can ignore the error.
	if err := v.BindEnv(env); err != nil {
		panic(fmt.Errorf("failed to bind env %q: %w", env, err))
	}
}
