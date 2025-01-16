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

// Package remote defines a remote audit log processor.
package remote

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

// Option is the option to set up a remote audit log processor.
type Option func(*Processor) error

// WithGRPCDialOptions allows provide raw grpc.DialOption for the underlying connection.
func WithGRPCDialOptions(opts ...grpc.DialOption) Option {
	return func(p *Processor) error {
		p.rawDialOpts = opts
		return nil
	}
}

// WithDefaultAuth sets up the processor to connect to remote with the default auth setting.
func WithDefaultAuth() Option {
	return WithIDTokenAuth(context.Background())
}

// grpcAuthOptions is the interface to get gRPC DialOptions and CallOptions
// specific to an auth setting.
type grpcAuthOptions interface {
	dialOpts() ([]grpc.DialOption, error)
	callOpts() ([]grpc.CallOption, error)
}

// Processor is the remote audit log processor.
type Processor struct {
	address     string
	conn        *grpc.ClientConn
	client      api.AuditLogAgentClient
	authOpts    grpcAuthOptions
	rawDialOpts []grpc.DialOption
}

// NewProcessor creates a new remote audit log processor.
//
// E.g.
//
//	p, err := NewProcessor("localhost:8080", WithDefaultAuth())
//	if err != nil { ... }
//	defer p.Close()
func NewProcessor(address string, opts ...Option) (*Processor, error) {
	p := &Processor{address: address}
	for _, o := range opts {
		if err := o(p); err != nil {
			return nil, fmt.Errorf("failed to set option: %w", err)
		}
	}

	dialOpts := append(p.rawDialOpts, grpc.WithAuthority(address))
	if p.authOpts != nil {
		authDialOpts, err := p.authOpts.dialOpts()
		if err != nil {
			return nil, fmt.Errorf("failed to generate gRPC auth dial options: %w", err)
		}
		dialOpts = append(dialOpts, authDialOpts...)
	} else {
		// If no auth option is provided, fall back to insecure.
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial remote log processor failed: %w", err)
	}

	p.conn = conn
	p.client = api.NewAuditLogAgentClient(conn)
	return p, nil
}

// Process processes the audit log request by calling a remote service.
func (p *Processor) Process(ctx context.Context, logReq *api.AuditLogRequest) error {
	var authCallOpts []grpc.CallOption
	if p.authOpts != nil {
		var err error
		authCallOpts, err = p.authOpts.callOpts()
		if err != nil {
			return fmt.Errorf("failed to generate gRPC auth call options: %w", err)
		}
	}
	resp, err := p.client.ProcessLog(ctx, logReq, authCallOpts...)
	if err != nil {
		return fmt.Errorf("remote log processing failed: %w", err)
	}

	if resp.GetResult() != nil {
		logReq.Labels = resp.GetResult().GetLabels()
		logReq.Payload = resp.GetResult().GetPayload()
		logReq.Type = resp.GetResult().GetType()
	}

	return nil
}

// Stop stops the processor.
func (p *Processor) Stop() error {
	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("failed to close grpc connection: %w", err)
	}
	return nil
}
