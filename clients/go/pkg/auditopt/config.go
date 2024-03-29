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
//
// Usually, auditopt should be used once on program initialization (e.g. in main.go).
// For example, in main.go:
//
//	opts, err := auditopt.FromConfigFile("path/to/config.yaml")
//	if err != nil {
//		// Handle err
//	}
//	client, err := audit.NewClient(opts...)
//	if err != nil {
//		// Handle err
//	}
package auditopt

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"cloud.google.com/go/logging"
	"github.com/sethvargo/go-envconfig"

	jvspb "github.com/abcxyz/jvs/apis/v0"
	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/cloudlogging"
	"github.com/abcxyz/lumberjack/clients/go/pkg/filtering"
	"github.com/abcxyz/lumberjack/clients/go/pkg/justification"
	"github.com/abcxyz/lumberjack/clients/go/pkg/remote"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/abcxyz/pkg/cfgloader"
)

const DefaultConfigFilePath = "/etc/lumberjack/config.yaml"

// FromConfigFile specifies a config file to configure the
// audit client. If `path` is nil, we use a default well known
// path. If the config file is not found, we keep going by
// using env vars and default values to configure the client.
func FromConfigFile(path string) audit.Option {
	return fromConfigFile(path, envconfig.OsLookuper())
}

// FromConfig creates an audit client option from the given configuration.
func FromConfig(cfg *api.Config) audit.Option {
	return func(ctx context.Context, c *audit.Client) error {
		return clientFromConfig(ctx, c, cfg)
	}
}

// fromConfigFile is like FromConfigFile, but exposes a custom lookuper for
// testing.
func fromConfigFile(path string, lookuper envconfig.Lookuper) audit.Option {
	return func(ctx context.Context, c *audit.Client) error {
		if path == "" {
			path = DefaultConfigFilePath
		}
		fc, err := os.ReadFile(path)
		// We ignore ErrNotExist when reading the file because we
		// still use env vars and defaults to setup the client.
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		cfg, err := loadConfig(ctx, fc, lookuper)
		if err != nil {
			return err
		}
		return clientFromConfig(ctx, c, cfg)
	}
}

// InterceptorFromConfigFile returns an interceptor option from the given
// config file. The returned option can be used to create an interceptor
// to add capability to gRPC server.
func InterceptorFromConfigFile(path string) audit.InterceptorOption {
	return interceptorFromConfigFile(path, envconfig.OsLookuper())
}

// interceptorFromConfigFile is like InterceptorFromConfigFile, but exposes a
// custom lookuper for testing.
func interceptorFromConfigFile(path string, lookuper envconfig.Lookuper) audit.InterceptorOption {
	return func(ctx context.Context, i *audit.Interceptor) error {
		fc, err := os.ReadFile(path)
		// We ignore ErrNotExist when reading the file because we
		// still use env vars and defaults to setup the client.
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		cfg, err := loadConfig(ctx, fc, lookuper)
		if err != nil {
			return err
		}

		// Interceptor requires security context.
		if cfg.SecurityContext == nil {
			return fmt.Errorf("SecurityContext must be provided to use interceptor")
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
		auditOpt := func(ctx context.Context, c *audit.Client) error {
			return clientFromConfig(ctx, c, cfg)
		}
		auditClient, err := audit.NewClient(ctx, auditOpt)
		if err != nil {
			return fmt.Errorf("failed to create audit client from config %+v: %w", cfg, err)
		}
		opts = append(opts, audit.WithAuditClient(auditClient))

		// Apply all options.
		for _, o := range opts {
			if err := o(ctx, i); err != nil {
				return err
			}
		}
		return nil
	}
}

func clientFromConfig(ctx context.Context, c *audit.Client, cfg *api.Config) error {
	if cfg == nil {
		return fmt.Errorf("nil config")
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	opts := []audit.Option{audit.WithRuntimeInfo()}

	withPrincipalFilter, err := principalFilterFromConfig(cfg)
	if err != nil {
		return err
	}
	if withPrincipalFilter != nil {
		opts = append(opts, withPrincipalFilter)
	}

	withBackends, err := backendsFromConfig(ctx, cfg)
	if err != nil {
		return err
	}
	opts = append(opts, withBackends...)

	withLabels := labelsFromConfig(ctx, cfg)
	opts = append(opts, withLabels)

	withLogMode := audit.WithLogMode(cfg.GetLogMode())
	opts = append(opts, withLogMode)

	if cfg.Justification != nil && cfg.Justification.Enabled {
		withJustification, err := justificationFromConfig(ctx, cfg)
		if err != nil {
			return err
		}
		opts = append(opts, withJustification)
	}

	for _, o := range opts {
		if err := o(ctx, c); err != nil {
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
		return nil, fmt.Errorf("failed to create email matcher: %w", err)
	}
	return audit.WithValidator(m), nil
}

func backendsFromConfig(ctx context.Context, cfg *api.Config) ([]audit.Option, error) {
	var backendOpts []audit.Option

	if cfg.Backend.Remote != nil {
		addr := cfg.Backend.Remote.Address
		authopts := []remote.Option{}
		if !cfg.Backend.Remote.InsecureEnabled {
			impersonate := cfg.Backend.Remote.ImpersonateAccount
			if impersonate == "" {
				authopts = append(authopts, remote.WithDefaultAuth())
			} else {
				authopts = append(authopts, remote.WithImpersonatedIDTokenAuth(ctx, impersonate))
			}
		}
		b, err := remote.NewProcessor(addr, authopts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create remote processor: %w", err)
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
			p, perr = cloudlogging.NewProcessor(ctx, opts...)
		} else {
			clc, err := logging.NewClient(ctx, cfg.Backend.CloudLogging.Project)
			if err != nil {
				return nil, fmt.Errorf("failed to create cloud logging client: %w", err)
			}
			opts = append(opts, cloudlogging.WithLoggingClient(clc))
			p, perr = cloudlogging.NewProcessor(ctx, opts...)
		}

		if perr != nil {
			return nil, fmt.Errorf("failed to create processor: %w", perr)
		}
		backendOpts = append(backendOpts, audit.WithBackend(p))
	}

	return backendOpts, nil
}

func labelsFromConfig(ctx context.Context, cfg *api.Config) audit.Option {
	lp := audit.NewLabelProcessor(ctx, cfg.Labels)
	return audit.WithMutator(lp)
}

func justificationFromConfig(ctx context.Context, cfg *api.Config) (audit.Option, error) {
	jvsconfig := jvspb.Config{
		JWKSEndpoint:    cfg.Justification.PublicKeysEndpoint,
		AllowBreakglass: cfg.Justification.AllowBreakglass,
	}
	if err := cfgloader.Load(ctx, &jvsconfig, cfgloader.WithEnvPrefix("AUDIT_CLIENT_")); err != nil {
		return nil, fmt.Errorf("failed to load JVS config: %w", err)
	}
	jvsClient, err := jvspb.NewClient(ctx, &jvsconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create jvs client: %w", err)
	}
	p := justification.NewProcessor(jvsClient)
	return audit.WithMutator(p), nil
}

func loadConfig(ctx context.Context, b []byte, lookuper envconfig.Lookuper) (*api.Config, error) {
	var cfg api.Config
	loadOpts := []cfgloader.Option{
		cfgloader.WithEnvPrefix("AUDIT_CLIENT_"),
		cfgloader.WithYAML(b),
		cfgloader.WithLookuper(lookuper),
	}
	if err := cfgloader.Load(ctx, &cfg, loadOpts...); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}
