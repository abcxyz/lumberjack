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

// Package server implements the gRPC server of the audit log agent.
package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
)

// AuditLogAgent is the implementation of the audit log agent server.
type AuditLogAgent struct {
	api.UnimplementedAuditLogAgentServer

	client *audit.Client
}

// NewAuditLogAgent creates a new AuditLogAgent.
func NewAuditLogAgent(client *audit.Client) (*AuditLogAgent, error) {
	return &AuditLogAgent{client: client}, nil
}

// ProcessLog processes the log requests by calling the internal client.
func (a *AuditLogAgent) ProcessLog(ctx context.Context, logReq *api.AuditLogRequest) (*api.AuditLogResponse, error) {
	if err := a.client.Log(ctx, logReq); err != nil {
		return nil, codifyErr(err)
	}

	return &api.AuditLogResponse{
		Result: logReq,
	}, nil
}

func codifyErr(err error) error {
	if errors.Is(err, audit.ErrInvalidRequest) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	// TODO: Handle other well-known errors if we have more.
	return status.Error(codes.Internal, err.Error())
}
