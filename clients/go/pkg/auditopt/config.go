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
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// The list of config variables that a user can set in a config
// file. The "." delimeter represents a nested field. For example,
// the config variable "filter.regex.principal_include" is represented
// in a YAML config file as:
// ```
// filter:
//  regex:
//    principal_exclude: test@google.com
// ```
const (
	versionKey = "version"
	// TODO(noamrabbani): rename `filter` to `condition`
	filterRegexPrincipalIncludeKey  = "filter.regex.principal_include"
	filterRegexPrincipalExcludeKey  = "filter.regex.principal_exclude"
	backendAddressKey               = "backend.address"
	backendInsecureEnabledKey       = "backend.insecure_enabled"
	backendImpersonateAccountKey    = "backend.impersonate_account"
	securityContext                 = "security_context"
	securityContextFromRawJWT       = "security_context.from_raw_jwt"
	securityContextFromRawJWTKey    = "security_context.from_raw_jwt.key"
	securityContextFromRawJWTPrefix = "security_context.from_raw_jwt.prefix"
)

// The version we expect in a config file.
const expectedVersion = "v1alpha1"

const defaultConfigFilePath = "/etc/auditlogging/config.yaml"

// MustFromConfigFile reads a config file to create:
//   - an Option to configure the audit client
//   - a SecurityContext to configure the interceptor
// The field `path` is required. If the config file is missing, we return an error.
func MustFromConfigFile(path string) (audit.Option, audit.SecurityContext, error) {
	v := prepareViper()
	if err := setupViperConfigFile(v, path); err != nil {
		return nil, nil, err
	}

	opt := func(c *audit.Client) error {
		return configureClientFromViper(c, v)
	}
	sc := securityContextFromViper(v)
	return opt, sc, nil
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
		v := prepareViper()
		// We don't return an error if the file is not found because we
		// still use env vars and defaults to setup the client.
		if err := setupViperConfigFile(v, path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return configureClientFromViper(c, v)
	}
}

func configureClientFromViper(c *audit.Client, v *viper.Viper) error {
	opts := []audit.Option{audit.WithRuntimeInfo()}

	withPrincipalFilter, err := principalFilterFromViper(v)
	if err != nil {
		return err
	}
	opts = append(opts, withPrincipalFilter)

	withBackend, err := backendFromViper(v)
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

func principalFilterFromViper(v *viper.Viper) (audit.Option, error) {
	var mopts []filtering.Option
	if v.IsSet(filterRegexPrincipalIncludeKey) {
		include := v.GetString(filterRegexPrincipalIncludeKey)
		mopts = append(mopts, filtering.WithIncludes(include))
	}
	if v.IsSet(filterRegexPrincipalExcludeKey) {
		exclude := v.GetString(filterRegexPrincipalExcludeKey)
		mopts = append(mopts, filtering.WithExcludes(exclude))
	}

	m, err := filtering.NewPrincipalEmailMatcher(mopts...)
	if err != nil {
		return nil, err
	}
	return audit.WithValidator(m), nil
}

func backendFromViper(v *viper.Viper) (audit.Option, error) {
	authopts := []remote.Option{}
	insecure := v.GetBool(backendInsecureEnabledKey)
	impersonate := v.GetString(backendImpersonateAccountKey)
	if !insecure {
		if impersonate == "" {
			authopts = append(authopts, remote.WithDefaultAuth())
		} else {
			authopts = append(authopts, remote.WithImpersonatedIDTokenAuth(context.Background(), impersonate))
		}
	}
	addr := v.GetString(backendAddressKey)
	if addr == "" {
		return nil, fmt.Errorf("config backend address is nil, set it as an env var or in a config file")
	}
	b, err := remote.NewProcessor(addr, authopts...)
	if err != nil {
		return nil, err
	}
	return audit.WithBackend(b), nil
}

func securityContextFromViper(v *viper.Viper) audit.SecurityContext {
	if !v.IsSet(securityContext) {
		return nil
	}
	if !v.IsSet(securityContextFromRawJWT) {
		return nil
	}
	v.SetDefault(securityContextFromRawJWTKey, "authorization")
	v.SetDefault(securityContextFromRawJWTPrefix, "bearer ")

	return &audit.FromRawJWT{
		Key:    v.GetString(securityContextFromRawJWTKey),
		Prefix: v.GetString(securityContextFromRawJWTPrefix),
	}
}

// prepareViper creates a Viper instance that:
//   1. declares config variables defaults
//   2. reads environment variables and maps them to config variables
// The Viper library does most of the heavy lifting. For details, see:
// https://github.com/spf13/viper#what-is-viper
func prepareViper() *viper.Viper {
	v := viper.New()

	// By default, we filter log requests that have an IAM
	// service account as the principal. All other config
	// variables have the default values associated with
	// their types.
	v.SetDefault(filterRegexPrincipalExcludeKey, ".iam.gserviceaccount.com$")

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

// setupViperConfigFile reads a config file and updates the given
// Viper to use the config file variables. Viper supports
// JSON, TOML, YAML, HCL, envfile and Java properties config files.
func setupViperConfigFile(v *viper.Viper, path string) error {
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed reading config file at %q: %w", path, err)
	}

	configFileVersion := v.GetString(versionKey)
	if configFileVersion != expectedVersion {
		return fmt.Errorf("config version %q unsupported, supported versions are [%q]", configFileVersion, expectedVersion)
	}
	return nil
}

// WithInterceptorFromConfig returns a gRPC server option that adds a unary interceptor
// to a gRPC server. This interceptor autofills and emits audit logs for gRPC unary
// calls. WithInterceptorFromConfig also returns the audit client that the interceptor
// uses. This allows the caller to close the client when shutting down the gRPC server.
// For example:
// ```
// opt, c, err := audit.WithInterceptorFromConfig("auditconfig.yaml")
// if err != nil {
//	log.Fatalf(err)
// }
// defer c.Stop()
// s := grpc.NewServer(opt)
// ```
// TODO(noamrabbani): add streaming interceptor.
func WithInterceptorFromConfig(path string) (grpc.ServerOption, *audit.Client, error) {
	opt, sc, err := MustFromConfigFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create audit option and security context from config file %q: %v", path, err)
	}
	if sc == nil {
		return nil, nil, fmt.Errorf("security_context is nil in config file %q", path)
	}

	auditClient, err := audit.NewClient(opt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create audit client from config file %q: %v", path, err)
	}
	interceptor := &audit.Interceptor{
		Client:          auditClient,
		SecurityContext: sc,
	}
	return grpc.UnaryInterceptor(interceptor.UnaryInterceptor), auditClient, nil
}
