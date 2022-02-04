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

package cloudlogging

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/reflect/protoreflect"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

// Processor is the remote Cloud Logging processor.
type Processor struct {
	client *logging.Client
	// bestEffort defaults to false and sets the default
	// logging mode to fail-close.
	bestEffort bool
	// loggerByLogType is not threadsafe.
	loggerByLogType map[alpb.AuditLogRequest_LogType]*logging.Logger
}

// Option is the option to set up a Cloud Logging processor.
type Option func(p *Processor) error

// WithLoggingClient provides a Cloud Logging client to the processor.
func WithLoggingClient(client *logging.Client) Option {
	return func(p *Processor) error {
		p.client = client
		return nil
	}
}

// WithDefaultBestEffort sets the default logging mode of operation to best effort.
// There are two possible logging modes of operation:
//   - fail-close. Without this option, this is default. The Cloud Logging
//   client blocks to emit a log, and immediately returns an error when
//   there's a failure.
//   - best effort. The Cloud Logging client emits logs asynchronously
//   and does not return an error on failure. Calling `Stop()` flushes
//   the logs and returns all previously encountered errors.
// TODO(b/203776475): individual log requests can specify failclose or besteffort logging
func WithDefaultBestEffort() Option {
	return func(p *Processor) error {
		p.bestEffort = true
		return nil
	}
}

// NewProcessor creates a new Cloud Logging log processor
// with the given options.
func NewProcessor(ctx context.Context, opts ...Option) (*Processor, error) {
	p := &Processor{}
	for _, o := range opts {
		if err := o(p); err != nil {
			return nil, fmt.Errorf("failed to apply client options: %w", err)
		}
	}

	// If the options didn't provide a Cloud Logging client, create one.
	if p.client == nil {
		projectID, err := metadata.ProjectID()
		if err != nil {
			return nil, fmt.Errorf("error getting project ID from metadata to initialize the default Cloud Logging client: %w", err)
		}
		client, err := logging.NewClient(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("error creating the default Cloud Logging client for project %v: %w", projectID, err)
		}
		p.client = client
	}

	loggerByLogType := map[alpb.AuditLogRequest_LogType]*logging.Logger{}
	for v := range alpb.AuditLogRequest_LogType_name {
		logType := alpb.AuditLogRequest_LogType(v)
		logName := logNameFromLogType(logType)
		if logName == "" {
			return nil, fmt.Errorf("the log type %v is not annotated with a log name", logType)
		}
		loggerByLogType[logType] = p.client.Logger(logName)
	}
	p.loggerByLogType = loggerByLogType
	return p, nil
}

// logNameFromLogType obtains the Cloud Logging LogName by reading the
// proto annotation of the AuditLogRequest.Type. If the proto annotation
// is missing, we use a default logName.
func logNameFromLogType(t alpb.AuditLogRequest_LogType) string {
	enumOpts := t.Descriptor().Values().ByNumber(t.Number()).Options().ProtoReflect()
	var logName string
	enumOpts.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.Name() == "log_name" {
			logName = v.String()
			return false
		}
		return true
	})
	return logName
}

// Process emits an audit logs to Cloud Logging synchronously.
func (p *Processor) Process(ctx context.Context, logReq *alpb.AuditLogRequest) error {
	logger, ok := p.loggerByLogType[logReq.Type]
	if !ok {
		// Hitting this code path would be unlikely because NewProcessor
		// creates loggers for every log type in the AuditLogRequest proto.
		logger = p.loggerByLogType[alpb.AuditLogRequest_UNSPECIFIED]
	}
	logEntry := logging.Entry{
		Payload:   logReq.Payload,
		Labels:    logReq.Labels,
		Operation: logReq.Operation,
	}

	bestEffort := p.bestEffort
	if logReq.Mode != alpb.AuditLogRequest_LOG_MODE_UNSPECIFIED {
		bestEffort = (logReq.Mode == alpb.AuditLogRequest_BEST_EFFORT)
	}

	if bestEffort {
		logger.Log(logEntry)
		return nil
	}
	if err := logger.LogSync(ctx, logEntry); err != nil {
		return fmt.Errorf("synchronous write to Cloud logging failed: %w", err)
	}
	return nil
}

// Stop stops the processor by flushing the logs from all loggers.
// Stop is only meaningful when the client emitted logs as best-effort.
func (p *Processor) Stop() error {
	var merr error
	for logtype, logger := range p.loggerByLogType {
		if err := logger.Flush(); err != nil {
			merr = multierr.Append(merr, fmt.Errorf("error flushing logs with type %v: %w", logtype, err))
		}
	}
	return merr
}
