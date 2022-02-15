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
	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/types/known/structpb"
)

// RequestBuilder is intended to be used as a helper for test classes.
// It builds basic AuditLogRequests with sane defaults for testing,
// and exposes a builder in order to create requests for tests that have
// specific requirements.
type RequestBuilder struct {
	auditLogRequest *alpb.AuditLogRequest
}

func ReqBuilder() *RequestBuilder {
	return &RequestBuilder{auditLogRequest: &alpb.AuditLogRequest{
		Type: alpb.AuditLogRequest_DATA_ACCESS,
		Payload: &audit.AuditLog{
			ServiceName:  "test-service",
			ResourceName: "test-resource",
			AuthenticationInfo: &audit.AuthenticationInfo{
				PrincipalEmail: "user@example.com",
			},
		},
	}}
}

func (b *RequestBuilder) WithLabels(labels map[string]string) *RequestBuilder {
	b.auditLogRequest.Labels = labels
	return b
}

func (b *RequestBuilder) WithPrincipal(principal string) *RequestBuilder {
	b.auditLogRequest.Payload.AuthenticationInfo.PrincipalEmail = principal
	return b
}

func (b *RequestBuilder) WithMethodName(methodName string) *RequestBuilder {
	b.auditLogRequest.Payload.MethodName = methodName
	return b
}

func (b *RequestBuilder) WithServiceName(serviceName string) *RequestBuilder {
	b.auditLogRequest.Payload.ServiceName = serviceName
	return b
}

func (b *RequestBuilder) WithMetadata(metadata *structpb.Struct) *RequestBuilder {
	b.auditLogRequest.Payload.Metadata = metadata
	return b
}

func (b *RequestBuilder) WithMode(logMode alpb.AuditLogRequest_LogMode) *RequestBuilder {
	b.auditLogRequest.Mode = logMode
	return b
}

func (b *RequestBuilder) Build() *alpb.AuditLogRequest {
	return b.auditLogRequest
}
