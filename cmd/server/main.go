// Copyright 2021 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cloudloggingsdk "cloud.google.com/go/logging"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/cloudlogging"
	"github.com/abcxyz/lumberjack/clients/go/pkg/trace"
	"github.com/abcxyz/lumberjack/internal/version"
	"github.com/abcxyz/lumberjack/pkg/server"
	"github.com/abcxyz/pkg/gcputil"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/serving"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer done()

	logger := logging.NewFromEnv("LUMBERJACK_")
	ctx = logging.WithLogger(ctx, logger)

	if err := realMain(ctx); err != nil {
		done()
		logger.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
	logger.InfoContext(ctx, "successful shutdown")
}

func realMain(ctx context.Context) (retErr error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "server starting",
		"commit", version.Commit,
		"version", version.Version)

	projectID := gcputil.ProjectID(ctx)

	cfg, err := server.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to process config: %w", err)
	}

	logger.DebugContext(ctx, "loaded configuration", "config", cfg)

	if err := trace.Init(cfg.TraceRatio); err != nil {
		return fmt.Errorf("failed to init tracing: %w", err)
	}
	defer func() {
		if err := trace.Shutdown(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to shutdown tracing: %w", err))
		}
	}()

	// Set up other log processors as we add more.
	// TODO(b/202328178): Allow setting other log processor(s) via config.
	// E.g. We can have a stdout log processor that write audit logs to stdout.
	cl, err := newCloudLoggingProcessor(ctx, projectID)
	if err != nil {
		return err
	}
	client, err := audit.NewClient(ctx, audit.WithBackend(cl))
	if err != nil {
		return fmt.Errorf("failed to create audit client: %w", err)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to stop audit client: %w", err))
		}
	}()

	logAgent, err := server.NewAuditLogAgent(client)
	if err != nil {
		return fmt.Errorf("failed to create audit log agent: %w", err)
	}

	// TODO(b/202320320): Build interceptors for observability, logger, etc.
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		otelgrpc.UnaryServerInterceptor(),
		logging.GRPCUnaryInterceptor(logger, projectID),
	))
	api.RegisterAuditLogAgentServer(grpcServer, logAgent)
	reflection.Register(grpcServer)

	server, err := serving.New(cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to create serving infrastructure: %w", err)
	}
	return server.StartGRPC(ctx, grpcServer)
}

func newCloudLoggingProcessor(ctx context.Context, projectID string) (*cloudlogging.Processor, error) {
	client, err := cloudloggingsdk.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud logging client: %w", err)
	}
	p, err := cloudlogging.NewProcessor(ctx, cloudlogging.WithLoggingClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud logging audit log processor: %w", err)
	}
	return p, nil
}
