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

// Package testutil provides utilities that are intended to enable easier
// and more concise writing of unit test code.
package testutil

import (
	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"

	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/types/known/structpb"
)

type RequestOptions func(r *api.AuditLogRequest) *api.AuditLogRequest

func NewRequest(opts ...RequestOptions) *api.AuditLogRequest {
	request := &api.AuditLogRequest{
		Type: api.AuditLogRequest_DATA_ACCESS,
		Payload: &audit.AuditLog{
			ServiceName:  "test-service",
			ResourceName: "test-resource",
			AuthenticationInfo: &audit.AuthenticationInfo{
				PrincipalEmail: "user@example.com",
			},
		},
	}
	for _, opt := range opts {
		request = opt(request)
	}
	return request
}

func WithLabels(labels map[string]string) RequestOptions {
	return func(r *api.AuditLogRequest) *api.AuditLogRequest {
		r.Labels = labels
		return r
	}
}

func WithPrincipal(principal string) RequestOptions {
	return func(r *api.AuditLogRequest) *api.AuditLogRequest {
		r.Payload.AuthenticationInfo.PrincipalEmail = principal
		return r
	}
}

func WithMethodName(method string) RequestOptions {
	return func(r *api.AuditLogRequest) *api.AuditLogRequest {
		r.Payload.MethodName = method
		return r
	}
}

func WithServiceName(service string) RequestOptions {
	return func(r *api.AuditLogRequest) *api.AuditLogRequest {
		r.Payload.ServiceName = service
		return r
	}
}

func WithMetadata(metadata *structpb.Struct) RequestOptions {
	return func(r *api.AuditLogRequest) *api.AuditLogRequest {
		r.Payload.Metadata = metadata
		return r
	}
}

func WithMode(mode api.AuditLogRequest_LogMode) RequestOptions {
	return func(r *api.AuditLogRequest) *api.AuditLogRequest {
		r.Mode = mode
		return r
	}
}
