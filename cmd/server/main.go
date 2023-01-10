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
	"fmt"
	"net"
	"os/signal"
	"syscall"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/cloudlogging"
	"github.com/abcxyz/lumberjack/clients/go/pkg/trace"
	"github.com/abcxyz/lumberjack/internal/version"
	"github.com/abcxyz/lumberjack/pkg/server"
	zlogger "github.com/abcxyz/pkg/logging"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	logger := zlogger.NewFromEnv("")
	ctx = zlogger.WithLogger(ctx, logger)

	if err := realMain(ctx); err != nil {
		done()
		logger.Fatal(err)
	}
	logger.Info("successful shutdown")
}

func realMain(ctx context.Context) error {
	logger := zlogger.FromContext(ctx)
	logger.Debugw("server starting",
		"commit", version.Commit,
		"version", version.Version)

	cfg, err := server.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to process config: %w", err)
	}

	if err := trace.Init(cfg.TraceRatio); err != nil {
		return fmt.Errorf("failed to init tracing: %w", err)
	}

	// Set up other log processors as we add more.
	// TODO(b/202328178): Allow setting other log processor(s) via config.
	// E.g. We can have a stdout log processor that write audit logs to stdout.
	cl, err := newCloudLoggingProcessor(ctx)
	if err != nil {
		return err
	}
	client, err := audit.NewClient(audit.WithBackend(cl))
	if err != nil {
		return fmt.Errorf("failed to create audit client: %w", err)
	}
	logAgent, err := server.NewAuditLogAgent(client)
	if err != nil {
		return fmt.Errorf("failed to create audit log agent: %w", err)
	}

	// TODO(b/202320320): Build interceptors for observability, logger, etc.
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		otelgrpc.UnaryServerInterceptor(),
		zlogger.GRPCInterceptor(logger),
	))
	api.RegisterAuditLogAgentServer(s, logAgent)
	reflection.Register(s)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %q: %w", cfg.Port, err)
	}

	// TODO: Do we need a gRPC health check server?
	// https://github.com/grpc/grpc/blob/master/doc/health-checking.md
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		logger.Infof("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			return fmt.Errorf("failed to start grpc server: %w", err)
		}
		return nil
	})

	// Either we have received a TERM signal or errgroup has encountered an err.
	<-ctx.Done()
	s.GracefulStop()
	if err := client.Stop(); err != nil {
		return fmt.Errorf("error stopping audit client: %w", err)
	}
	if err := trace.Shutdown(); err != nil {
		return fmt.Errorf("error shutdown tracing: %w", err)
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error running server: %w", err)
	}
	return nil
}

func newCloudLoggingProcessor(ctx context.Context) (*cloudlogging.Processor, error) {
	projectID, err := metadata.ProjectID()
	if err != nil {
		return nil, fmt.Errorf("failed to get project ID from metadata server: %w", err)
	}
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud logging client: %w", err)
	}
	p, err := cloudlogging.NewProcessor(ctx, cloudlogging.WithLoggingClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud logging audit log processor: %w", err)
	}
	return p, nil
}
