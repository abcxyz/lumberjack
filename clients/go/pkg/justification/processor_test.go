// Copyright 2022 Lumberjack authors (see AUTHORS file)
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

package justification

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/genproto/googleapis/rpc/context/attribute_context"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"

	jvspb "github.com/abcxyz/jvs/apis/v0"
	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/pkg/logging"
	pkgtestutil "github.com/abcxyz/pkg/testutil"
)

func TestProcess(t *testing.T) {
	t.Parallel()

	ctx := logging.WithLogger(context.Background(), logging.TestLogger(t))

	cases := []struct {
		name       string
		token      string
		validator  *fakeValidator
		logReq     *api.AuditLogRequest
		wantLogReq *api.AuditLogRequest
		wantErr    string
	}{
		{
			name: "success",
			validator: &fakeValidator{
				justifications: []*jvspb.Justification{
					{Category: "explanation", Value: "need-access"},
				},
			},
			logReq: &api.AuditLogRequest{
				JustificationToken: "token",
				Payload:            &audit.AuditLog{},
			},
			wantLogReq: &api.AuditLogRequest{
				JustificationToken: "token",
				Payload: &audit.AuditLog{
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							LogMetadataKey: structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"iss":   structpb.NewStringValue("test_iss"),
									"sub":   structpb.NewStringValue("test_sub"),
									"email": structpb.NewStringValue("user@example.com"),
									"justs": structpb.NewListValue(&structpb.ListValue{
										Values: []*structpb.Value{structpb.NewStructValue(
											&structpb.Struct{Fields: map[string]*structpb.Value{
												"category": structpb.NewStringValue("explanation"),
												"value":    structpb.NewStringValue("need-access"),
											}},
										)},
									}),
								},
							}),
						},
					},
					RequestMetadata: &audit.RequestMetadata{
						RequestAttributes: &attribute_context.AttributeContext_Request{
							Reason: `[{"category":"explanation","value":"need-access"}]`,
						},
					},
				},
			},
		},
		{
			name: "empty_token_error",
			validator: &fakeValidator{
				justifications: []*jvspb.Justification{
					{Category: "explanation", Value: "need-access"},
				},
			},
			logReq:     &api.AuditLogRequest{Payload: &audit.AuditLog{}},
			wantLogReq: &api.AuditLogRequest{Payload: &audit.AuditLog{}},
			wantErr:    "missing justification token",
		},
		{
			name:      "validation_err",
			validator: &fakeValidator{returnErr: true},
			logReq: &api.AuditLogRequest{
				JustificationToken: "token",
				Payload:            &audit.AuditLog{},
			},
			wantLogReq: &api.AuditLogRequest{
				JustificationToken: "token",
				Payload:            &audit.AuditLog{},
			},
			wantErr: "invalid justification token",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Processor{validator: tc.validator}
			err := p.Process(ctx, tc.logReq)

			if diff := pkgtestutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Error(diff)
			}

			if diff := cmp.Diff(tc.wantLogReq, tc.logReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process got log request (-want,+got):\n%s", diff)
			}
		})
	}
}

type fakeValidator struct {
	returnErr      bool
	justifications []*jvspb.Justification
}

func (v *fakeValidator) ValidateJWT(_ context.Context, jvsToken string) (jwt.Token, error) {
	if v.returnErr {
		return nil, fmt.Errorf("failed to validate JWT")
	}
	if jvsToken == "" {
		return nil, fmt.Errorf("token empty")
	}

	tok, err := jwt.NewBuilder().
		Issuer(`test_iss`).
		Subject("test_sub").
		Build()
	if err != nil {
		return nil, err
	}
	if err := tok.Set("email", "user@example.com"); err != nil {
		return nil, err
	}
	if err := tok.Set("justs", v.justifications); err != nil {
		return nil, err
	}
	return tok, nil
}
