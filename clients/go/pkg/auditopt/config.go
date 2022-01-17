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

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/filtering"
	"github.com/abcxyz/lumberjack/clients/go/pkg/remote"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
)

// The list of config variables that a user can set in a config
// file. The "." delimeter represents a nested field. For example,
// the config variable "condition.regex.principal_include" is represented
// in a YAML config file as:
// ```
// condition:
//  regex:
//    principal_exclude: test@google.com
// ```
// If you add a new key, you need to explicitly give it a default value.
// Otherwise, Viper fails to map the new key to its env var representation.
const (
	versionKey                        = "version"
	conditionRegexPrincipalIncludeKey = "condition.regex.principal_include"
	// todo: add test to try overwriting unset key
	conditionRegexPrincipalExcludeKey = "condition.regex.principal_exclude"
	backendAddressKey                 = "backend.address"
	backendInsecureEnabledKey         = "backend.insecure_enabled"
	backendImpersonateAccountKey      = "backend.impersonate_account"
	securityContextKey                = "security_context"
	securityContextFromRawJWTKey      = "security_context.from_raw_jwt"
)

// The version we expect in a config file.
const expectedVersion = "v1alpha1"

const defaultConfigFilePath = "/etc/auditlogging/config.yaml"

// MustFromConfigFile specifies a config file to configure the
// audit client. `path` is required, and if the config file is
// missing, we return an error.
func MustFromConfigFile(path string) audit.Option {
	v := prepareViper()
	return func(c *audit.Client) error {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed reading config file at %q: %w", path, err)
		}
		cfg, err := configFromViper(v)
		if err != nil {
			return fmt.Errorf("failed to setup config from viper: %w", err)
		}
		return clientFromConfig(c, cfg)
	}
}

// FromConfigFile specifies a config file to configure the
// audit client. If `path` is nil, we use a default well known
// path. If the config file is not found, we keep going by
// using env vars and default values to configure the client.
func FromConfigFile(path string) audit.Option {
	v := prepareViper()
	return func(c *audit.Client) error {
		if path == "" {
			path = defaultConfigFilePath
		}
		// We don't return an error if the file is not found because we
		// still use env vars and defaults to setup the client.
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed reading config file at %q: %w", path, err)
		}
		cfg, err := configFromViper(v)
		if err != nil {
			return fmt.Errorf("failed to setup config from viper, %w", err)
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
	v := prepareViper()
	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed reading config file at %q: %w", path, err)
	}
	cfg, err := configFromViper(v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup config from viper: %w", err)
	}

	auditOpt := func(c *audit.Client) error {
		return clientFromConfig(c, cfg)
	}
	auditClient, err := audit.NewClient(auditOpt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create audit client from config file %q: %w", path, err)
	}

	interceptor := &audit.Interceptor{
		Client: auditClient,
	}

	fromRawJWT, err := fromRawJWTFromViper(v)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading %q in config file %q: %w", securityContextFromRawJWTKey, path, err)
	}
	switch {
	case fromRawJWT != nil:
		interceptor.SecurityContext = fromRawJWT
	default:
		return nil, nil, fmt.Errorf("no supported security context configured in config file %q", path)
	}

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
	if cfg == nil || cfg.Backend == nil {
		return nil, fmt.Errorf("config backend is nil, set it as an env var or in a config file")
	}
	authopts := []remote.Option{}
	if !cfg.Backend.InsecureEnabled {
		impersonate := cfg.Backend.ImpersonateAccount
		if impersonate == "" {
			authopts = append(authopts, remote.WithDefaultAuth())
		} else {
			authopts = append(authopts, remote.WithImpersonatedIDTokenAuth(context.Background(), impersonate))
		}
	}
	addr := cfg.Backend.Address
	if cfg.Backend.Address == "" {
		return nil, fmt.Errorf("config backend address is nil, set it as an env var or in a config file")
	}
	b, err := remote.NewProcessor(addr, authopts...)
	if err != nil {
		return nil, err
	}
	return audit.WithBackend(b), nil
}

// fromRawJWTFromViper populates a `fromRawJWT`` from a Viper instance.
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
func fromRawJWTFromViper(v *viper.Viper) (*security.FromRawJWT, error) {
	if !v.IsSet(securityContextKey) {
		return nil, nil
	}

	fromRawJWT := &security.FromRawJWT{}
	err := v.UnmarshalKey(securityContextFromRawJWTKey, fromRawJWT)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling key %q from config file: %w", securityContextFromRawJWTKey, err)
	}

	if fromRawJWT.Key == "" && fromRawJWT.Prefix == "" {
		fromRawJWT.Key = "authorization"
		fromRawJWT.Prefix = "bearer "
	}

	return fromRawJWT, nil
}

// prepareViper creates a Viper instance that:
//   1. declares config variables defaults
//   2. reads environment variables and maps them to config variables
// The Viper library does most of the heavy lifting. For details, see:
// https://github.com/spf13/viper#what-is-viper
func prepareViper() *viper.Viper {
	v := viper.New()

	// Set defaults
	// draft: if we don't set the default or the config file, automatic env doesn't pull them
	// todo: test this with a SA include because it has no default yet.
	// - I tested it and confirmed this is the actual behaviour
	// maybe add a test for prepareViper?
	v.SetDefault(versionKey, "v1alpha1")
	v.SetDefault(backendAddressKey, "")
	v.SetDefault(backendInsecureEnabledKey, false)
	// By default, we filter log requests that have an IAM
	// service account as the principal.
	v.SetDefault(conditionRegexPrincipalExcludeKey, ".iam.gserviceaccount.com$")

	// Bind env vars to our config variables.
	//   - Env vars should be prefixed with "AUDIT_CLIENT".
	//   - Nested config variables can be set from env vars by replacing
	//     the "." delimiter with "_". For example, the config variable "backend.address"
	//     can be set with the env var "AUDIT_CLIENT_BACKEND_ADDRESS".
	v.SetEnvPrefix("AUDIT_CLIENT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	return v
}

// configFromViper populates the config struct from a Viper.
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
