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

// Package audit provides functionality to validate and emit application audit logs.
package audit

import (
	"context"
	"errors"
	"fmt"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/auditerrors"
	"github.com/abcxyz/pkg/logging"
)

// Client is the Lumberjack audit logging Client.
type Client struct {
	validators []LogProcessor
	mutators   []LogProcessor
	backends   []LogProcessor
	logMode    api.AuditLogRequest_LogMode
}

// LogProcessor is the interface we use to process an AuditLogRequest.
// Examples include:
//   - validate that the AuditLogRequest is properly formed
//   - convert an AuditLogRequest to a Cloud LogEntry and write it to Cloud Logging
type LogProcessor interface {
	Process(context.Context, *api.AuditLogRequest) error
}

// StoppableProcessor is the interface to log processors that are stoppable.
type StoppableProcessor interface {
	Stop() error
}

// An Option is a configuration Option for NewClient.
type Option func(ctx context.Context, o *Client) error

// WithValidator adds the given log processor to validate audit log requests.
// The validators are executed in the order provided with this
// option and before any further audit log processing.
func WithValidator(p LogProcessor) Option {
	return func(ctx context.Context, o *Client) error {
		o.validators = append(o.validators, p)
		return nil
	}
}

// WithMutator adds the given log processor to mutate audit log requests.
// The mutators are executed in the order provided with this
// option. Mutators are executed after validators, but before backends.
func WithMutator(p LogProcessor) Option {
	return func(ctx context.Context, o *Client) error {
		o.mutators = append(o.mutators, p)
		return nil
	}
}

// WithRuntimeInfo adds the runtime info to all the audit log requests.
func WithRuntimeInfo() Option {
	return func(ctx context.Context, o *Client) error {
		r, err := newRuntimeInfo(ctx)
		if err != nil {
			return fmt.Errorf("error extracting runtime environment info: %w", err)
		}
		o.mutators = append(o.mutators, r)
		return nil
	}
}

// WithBackend adds the given log processor as a logging backend. Log
// backend processors are executed in the order provided with this
// option and after any other audit log processing.
// Examples of logging backends are:
//   - The Cloud Logging GCP service
//   - The custom Lumberjack gRPC service
func WithBackend(p LogProcessor) Option {
	return func(ctx context.Context, o *Client) error {
		o.backends = append(o.backends, p)
		return nil
	}
}

// Sets FailClose value. This specifies whether errors should be surfaced
// or swalled. Can be overridden on a per-request basis.
func WithLogMode(mode api.AuditLogRequest_LogMode) Option {
	return func(ctx context.Context, o *Client) error {
		o.logMode = mode
		return nil
	}
}

// NewClient initializes a logger with the given options.
func NewClient(ctx context.Context, opts ...Option) (*Client, error) {
	client := &Client{
		// Default processors.
		validators: []LogProcessor{
			NewRequestValidator(ctx),
		},
	}
	for _, f := range opts {
		if err := f(ctx, client); err != nil {
			return nil, fmt.Errorf("failed to apply client options: %w", err)
		}
	}
	return client, nil
}

// Stop stops the client.
func (c *Client) Stop() error {
	var merr error
	for _, ps := range [][]LogProcessor{c.validators, c.backends} {
		for _, p := range ps {
			if stoppable, ok := p.(StoppableProcessor); ok {
				if err := stoppable.Stop(); err != nil {
					merr = errors.Join(merr, err)
				}
			}
		}
	}

	return merr
}

// Log runs the client processors sequentially on the given AuditLogRequest.
func (c *Client) Log(ctx context.Context, logReq *api.AuditLogRequest) error {
	logger := logging.FromContext(ctx)

	if logReq.GetMode() == api.AuditLogRequest_LOG_MODE_UNSPECIFIED {
		logReq.Mode = c.logMode
	}

	for _, p := range c.validators {
		if err := p.Process(ctx, logReq); err != nil {
			if errors.Is(err, auditerrors.ErrPreconditionFailed) {
				logger.WarnContext(ctx, "stopped log request processing as validator precondition failed",
					"validator", p,
					"error", err)
				return nil
			}
			return c.handleReturn(ctx, fmt.Errorf("failed to execute validator %T: %w", p, err), logReq.GetMode())
		}
	}

	for _, p := range c.mutators {
		if err := p.Process(ctx, logReq); err != nil {
			if errors.Is(err, auditerrors.ErrPreconditionFailed) {
				logger.WarnContext(ctx, "stopped log request processing as mutator precondition failed",
					"validator", p,
					"error", err)
				return nil
			}
			return c.handleReturn(ctx, fmt.Errorf("failed to execute mutator %T: %w", p, err), logReq.GetMode())
		}
	}

	for _, p := range c.backends {
		if err := p.Process(ctx, logReq); err != nil {
			if errors.Is(err, auditerrors.ErrPreconditionFailed) {
				logger.WarnContext(ctx, "stopped log request processing as backend precondition failed",
					"validator", p,
					"error", err)
				return nil
			}
			return c.handleReturn(ctx, fmt.Errorf("failed to execute backend %T: %w", p, err), logReq.GetMode())
		}
	}

	return nil
}

// handleReturn is intended to be a wrapper that handles the LogMode correctly, and returns errors or
// nil depending on whether the config and request have specified that they want to fail close.
func (c *Client) handleReturn(ctx context.Context, err error, requestedLogMode api.AuditLogRequest_LogMode) error {
	// If there is no error, just return nil.
	if err == nil {
		return nil
	}

	// If there is an error, and we should fail close, return that error.
	if api.ShouldFailClose(requestedLogMode) {
		return err
	}

	// If there is an error, and we shouldn't fail close, log and return nil.
	logger := logging.FromContext(ctx)
	logger.ErrorContext(ctx, "failed to audit log; continuing without audit logging",
		"error", err)
	return nil
}
