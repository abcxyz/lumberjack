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

package audit

import (
	"context"
	"testing"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
	"github.com/google/go-cmp/cmp"
	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	protostatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestUnaryLogger(t *testing.T) {
	t.Parallel()

	currentLogReq := &alpb.AuditLogRequest{}

	// Test JWT. See link below to view decoded version:
	// https://jwt.io/#debugger-io?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6InVzZXIiLCJpYXQiOjE1MTYyMzkwMjIsImVtYWlsIjoidXNlckBleGFtcGxlLmNvbSJ9.PXl-SJniWHMVLNYb77HmVFFqWTlu28xf9fou2GaT0Jc
	jwt := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6InVzZXIiLCJpYXQiOjE1MTYyMzkwMjIsImVtYWlsIjoidXNlckBleGFtcGxlLmNvbSJ9.PXl-SJniWHMVLNYb77HmVFFqWTlu28xf9fou2GaT0Jc"
	tests := []struct {
		name          string
		ctx           context.Context
		req           interface{}
		info          *grpc.UnaryServerInfo
		handler       grpc.UnaryHandler
		wantLogReq    *alpb.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "interceptor_autofills_everything",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": jwt,
			})),
			info: &grpc.UnaryServerInfo{
				FullMethod: "/ExampleService/ExampleMethod",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				LogReqInCtx(ctx).Payload.ResourceName = "ExampleResourceName"
				currentLogReq = LogReqInCtx(ctx)
				return nil, nil
			},
			wantLogReq: &alpb.AuditLogRequest{
				Payload: &calpb.AuditLog{
					ServiceName:  "ExampleService",
					MethodName:   "ExampleMethod",
					ResourceName: "ExampleResourceName",
					AuthenticationInfo: &calpb.AuthenticationInfo{
						PrincipalEmail: "user@example.com",
					},
					Request:  &structpb.Struct{},
					Response: &structpb.Struct{},
					Status:   &protostatus.Status{},
				},
			},
		},
	}
	for _, tc := range tests {
		// No parallel to avoid a race condition on variable currentLogReq.
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewClient()
			if err != nil {
				t.Fatal(err)
			}
			defer c.Stop()

			_, gotErr := c.UnaryLogger(tc.ctx, tc.req, tc.info, tc.handler)
			if diff := errutil.DiffSubstring(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Stop() got unexpected error substring: %v", diff)
			}

			if diff := cmp.Diff(tc.wantLogReq, currentLogReq, protocmp.Transform()); diff != "" {
				t.Errorf("UnaryLogger(...) got diff in LogReq (-want, +got): %v", diff)
			}

		})
	}
}
