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
	"github.com/spf13/viper"
	"google.golang.org/grpc"

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
			return fmt.Errorf("failed reading config file at %q: %w", path, err)
		}
		v, err := loadDefaultsAndEnvVars(v)
		if err != nil {
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
			return fmt.Errorf("failed reading config file at %q: %w", path, err)
		}
		v, err := loadDefaultsAndEnvVars(v)
		if err != nil {
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
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed reading config file at %q: %w", path, err)
	}
	v, err := loadDefaultsAndEnvVars(v)
	if err != nil {
		return nil, nil, err
	}
	cfg, err := configFromViper(v)
	if err != nil {
		return nil, nil, err
	}

	if cfg == nil || cfg.SecurityContext == nil {
		return nil, nil, fmt.Errorf("no supported security context configured in config file %q", path)
	}
	interceptor := &audit.Interceptor{}
	switch {
	case cfg.SecurityContext.FromRawJWT != nil:
		fromRawJWT, err := fromRawJWTFromConfig(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting `from_raw_jwt` in config file %q: %w", path, err)
		}
		interceptor.SecurityContext = fromRawJWT
	default:
		return nil, nil, fmt.Errorf("no supported security context configured in config file %q", path)
	}

	auditOpt := func(c *audit.Client) error {
		return clientFromConfig(c, cfg)
	}
	auditClient, err := audit.NewClient(auditOpt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create audit client from config file %q: %w", path, err)
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
	if cfg != nil && cfg.Condition != nil && cfg.Condition.Regex != nil {
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
	if cfg == nil || cfg.Backend == nil || cfg.Backend.Address == "" {
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

// fromRawJWTFromConfig populates a `fromRawJWT`` from a Viper instance.
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
// TODO(noamrabbani): add support for lists in `security_context`
func fromRawJWTFromConfig(cfg *alpb.Config) (*security.FromRawJWT, error) {
	if cfg == nil || cfg.SecurityContext == nil || cfg.SecurityContext.FromRawJWT == nil {
		return nil, fmt.Errorf("fromRawJWT in the config is nil, set it as an env var or in a config file")
	}
	fromRawJWT := &security.FromRawJWT{
		Key:    cfg.SecurityContext.FromRawJWT.Key,
		Prefix: cfg.SecurityContext.FromRawJWT.Prefix,
	}
	if fromRawJWT.Key == "" && fromRawJWT.Prefix == "" {
		fromRawJWT.Key = "authorization"
		fromRawJWT.Prefix = "bearer "
	}
	return fromRawJWT, nil
}

// loadDefaultsAndEnvVars modifies a Viper instance to:
//   - map env vars to config vars
//   - have config variable defaults
// The Viper library does most of the heavy lifting. For details, see:
// https://github.com/spf13/viper#what-is-viper
func loadDefaultsAndEnvVars(v *viper.Viper) (*viper.Viper, error) {
	// Explicitly bind env vars to leaf Viper keys. Note that:
	//   - Env vars are prefixed with "AUDIT_CLIENT_".
	//   - Nested config variables can be set from env vars by replacing
	//     the "." delimiter with "_". E.g. the config variable "backend.address"
	//     can be set with the env var "AUDIT_CLIENT_BACKEND_ADDRESS".
	v.SetEnvPrefix("AUDIT_CLIENT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	leafKeys := []string{
		// The "." delimeter represents a nested field. For example,
		// the config variable "condition.regex.principal_include"
		// is represented in a YAML config file as:
		// ```
		// condition:
		//  regex:
		//    principal_include: test@google.com
		// ```
		"backend.address",
		"backend.impersonate_account",
		"backend.insecure_enabled",
		"condition.regex.principal_exclude",
		"condition.regex.principal_include",
		"security_context.from_raw_jwt.prefix",
		"security_context.from_raw_jwt.key",
		"version",
	}
	for _, lk := range leafKeys {
		// We don't use v.AutomaticEnv() because it fails to bind env vars
		// when they are inexistent in the config and when they don't have
		// an explicit default value.
		if err := v.BindEnv(lk); err != nil {
			return nil, fmt.Errorf("error binding key %q to env var: %w", lk, err)
		}
	}

	// By default, we filter log requests that have an IAM
	// service account as the principal.
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

	return v, nil
}

// configFromViper populates the config struct from a Viper instance.
func configFromViper(v *viper.Viper) (*alpb.Config, error) {
	config := &alpb.Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshall viper into config struct: %w", err)
	}

	configFileVersion := config.Version
	if configFileVersion != expectedVersion {
		return nil, fmt.Errorf("config version %q unsupported, supported versions are [%q]", configFileVersion, expectedVersion)
	}
	return config, nil
}
