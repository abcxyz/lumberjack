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
//  1. env vars
//  2. a config file
//  3. defaults
//
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

	"cloud.google.com/go/logging"
	"github.com/abcxyz/jvs/client-lib/go/client"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/cloudlogging"
	"github.com/abcxyz/lumberjack/clients/go/pkg/filtering"
	"github.com/abcxyz/lumberjack/clients/go/pkg/remote"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v2"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

const DefaultConfigFilePath = "/etc/auditlogging/config.yaml"

// MustFromConfigFile specifies a config file to configure the
// audit client. `path` is required, and if the config file is
// missing, we return an error.
func MustFromConfigFile(path string) audit.Option {
	return mustFromConfigFile(path, envconfig.OsLookuper())
}

// mustFromConfigFile is like MustFromConfigFile, but exposes a custom lookuper
// for testing.
func mustFromConfigFile(path string, lookuper envconfig.Lookuper) audit.Option {
	return func(c *audit.Client) error {
		fc, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		cfg, err := loadConfig(fc, lookuper)
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
	return fromConfigFile(path, envconfig.OsLookuper())
}

// fromConfigFile is like FromConfigFile, but exposes a custom lookuper for
// testing.
func fromConfigFile(path string, lookuper envconfig.Lookuper) audit.Option {
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
		cfg, err := loadConfig(fc, lookuper)
		if err != nil {
			return err
		}
		return clientFromConfig(c, cfg)
	}
}

// InterceptorFromConfigFile returns an interceptor option from the given
// config file. The returned option can be used to create an interceptor
// to add capability to gRPC server.
func InterceptorFromConfigFile(ctx context.Context, path string) audit.InterceptorOption {
	return interceptorFromConfigFile(ctx, path, envconfig.OsLookuper())
}

// interceptorFromConfigFile is like InterceptorFromConfigFile, but exposes a
// custom lookuper for testing.
func interceptorFromConfigFile(ctx context.Context, path string, lookuper envconfig.Lookuper) audit.InterceptorOption {
	return func(i *audit.Interceptor) error {
		fc, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		cfg, err := loadConfig(fc, lookuper)
		if err != nil {
			return err
		}

		// Interceptor requires security context.
		if cfg.SecurityContext == nil {
			return fmt.Errorf("SecurityContext must be provided to use interceptor")
		}

		if cfg.JVSEndpoint != "" {
			jvsClient, err := client.NewJVSClient(ctx, &client.JVSConfig{JVSEndpoint: cfg.JVSEndpoint})
			if err != nil {
				return err
			}
			audit.WithJVS(jvsClient)
		}

		opts := []audit.InterceptorOption{
			audit.WithInterceptorLogMode(cfg.GetLogMode()),
			audit.WithAuditRules(cfg.Rules...),
		}

		// Add security context to interceptor.
		switch {
		case cfg.SecurityContext.FromRawJWT != nil:
			fromRawJWT := &security.FromRawJWT{
				FromRawJWT: cfg.SecurityContext.FromRawJWT,
			}
			opts = append(opts, audit.WithSecurityContext(fromRawJWT))
		default:
			// This should never happen because already validates the SecurityContext
			// when loading the config.
			return fmt.Errorf("no supported security context configured in config %+v", cfg)
		}

		// Add audit client to interceptor.
		auditOpt := func(c *audit.Client) error {
			return clientFromConfig(c, cfg)
		}
		auditClient, err := audit.NewClient(auditOpt)
		if err != nil {
			return fmt.Errorf("failed to create audit client from config %+v: %w", cfg, err)
		}
		opts = append(opts, audit.WithAuditClient(auditClient))

		// Apply all options.
		for _, o := range opts {
			if err := o(i); err != nil {
				return err
			}
		}
		return nil
	}
}

func clientFromConfig(c *audit.Client, cfg *api.Config) error {
	opts := []audit.Option{audit.WithRuntimeInfo()}

	withPrincipalFilter, err := principalFilterFromConfig(cfg)
	if err != nil {
		return err
	}
	if withPrincipalFilter != nil {
		opts = append(opts, withPrincipalFilter)
	}

	withBackends, err := backendsFromConfig(cfg)
	if err != nil {
		return err
	}
	opts = append(opts, withBackends...)

	withLabels := labelsFromConfig(cfg)
	opts = append(opts, withLabels)

	for _, o := range opts {
		if err := o(c); err != nil {
			return err
		}
	}
	return nil
}

func principalFilterFromConfig(cfg *api.Config) (audit.Option, error) {
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

func backendsFromConfig(cfg *api.Config) ([]audit.Option, error) {
	var backendOpts []audit.Option

	if cfg.Backend.Remote != nil {
		addr := cfg.Backend.Remote.Address
		authopts := []remote.Option{}
		if !cfg.Backend.Remote.InsecureEnabled {
			impersonate := cfg.Backend.Remote.ImpersonateAccount
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
		backendOpts = append(backendOpts, audit.WithBackend(b))
	}

	if cfg.Backend.CloudLogging != nil {
		var opts []cloudlogging.Option
		if cfg.GetLogMode() == api.AuditLogRequest_BEST_EFFORT {
			opts = append(opts, cloudlogging.WithDefaultBestEffort())
		}

		var p *cloudlogging.Processor
		var perr error

		if cfg.Backend.CloudLogging.DefaultProject {
			p, perr = cloudlogging.NewProcessor(context.TODO(), opts...)
		} else {
			clc, err := logging.NewClient(context.TODO(), cfg.Backend.CloudLogging.Project)
			if err != nil {
				return nil, fmt.Errorf("failed to create cloud logging client: %w", err)
			}
			opts = append(opts, cloudlogging.WithLoggingClient(clc))
			p, perr = cloudlogging.NewProcessor(context.TODO(), opts...)
		}

		if perr != nil {
			return nil, perr
		}
		backendOpts = append(backendOpts, audit.WithBackend(p))
	}

	return backendOpts, nil
}

func labelsFromConfig(cfg *api.Config) audit.Option {
	lp := audit.LabelProcessor{DefaultLabels: cfg.Labels}
	return audit.WithMutator(&lp)
}

func loadConfig(b []byte, lookuper envconfig.Lookuper) (*api.Config, error) {
	cfg := &api.Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	// Process overrides from env vars.
	l := envconfig.PrefixLookuper("AUDIT_CLIENT_", lookuper)
	if err := envconfig.ProcessWith(context.Background(), cfg, l); err != nil {
		return nil, err
	}

	cfg.SetDefault()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed validating config %+v: %w", cfg, err)
	}

	return cfg, nil
}
