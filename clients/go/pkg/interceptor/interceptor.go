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

// Package interceptor provides gRPC interceptors to autofill audit logs.
package interceptor

import (
	"context"

	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/zlogger"
	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	protostatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

type logger struct {
	*audit.Client
}

// WithAuditInterceptor returns a gRPC server option that
// adds a unary interceptor to autofill audit logs.
// TODO(noamrabbani): add streaming interceptor.
func WithAuditInterceptors(c *audit.Client) grpc.ServerOption {
	l := &logger{c}
	return grpc.UnaryInterceptor(l.UnaryInterceptor)
}

// UnaryInterceptor is a unary interceptor that autofills and emits audit logs.
// Currently, this is a placeholder implementation.
func (l *logger) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logReq := &alpb.AuditLogRequest{
		Payload: &calpb.AuditLog{
			ServiceName: "testServerOption",
			AuthenticationInfo: &calpb.AuthenticationInfo{
				PrincipalEmail: "testServerOption@example.com",
			},
			Status: &protostatus.Status{},
		},
	}
	err := l.Log(ctx, logReq)
	if err != nil {
		zlogger.FromContext(ctx).Warnf("unary interceptor failed to emit log: %v", err)
	}

	resp, handlerErr := handler(ctx, req)
	return resp, handlerErr
}
