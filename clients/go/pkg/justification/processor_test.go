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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"

	jvsapi "github.com/abcxyz/jvs/apis/v0"
	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

func TestProcess(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		token      string
		validator  *fakeValidator
		wantLogReq *api.AuditLogRequest
		wantErr    bool
	}{{
		name:  "success",
		token: "token",
		validator: &fakeValidator{
			justifications: []*jvsapi.Justification{{Category: "explanation", Value: "need-access"}},
		},
		wantLogReq: &api.AuditLogRequest{
			Payload: &audit.AuditLog{
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						justificationLogMetadataKey: structpb.NewStructValue(&structpb.Struct{
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
			},
		},
	}, {
		name: "empty_token_no_action",
		validator: &fakeValidator{
			justifications: []*jvsapi.Justification{{Category: "explanation", Value: "need-access"}},
		},
		wantLogReq: &api.AuditLogRequest{Payload: &audit.AuditLog{}},
	}, {
		name:       "validation_err",
		token:      "token",
		validator:  &fakeValidator{returnErr: true},
		wantLogReq: &api.AuditLogRequest{Payload: &audit.AuditLog{}},
		wantErr:    true,
	}}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Processor{Validator: tc.validator}
			gotLogReq := &api.AuditLogRequest{Payload: &audit.AuditLog{}}
			gotErr := p.Process(tc.token, gotLogReq)
			if (gotErr == nil) == tc.wantErr {
				t.Errorf("Process got err=%v, want err %v", gotErr, tc.wantErr)
			}
			if diff := cmp.Diff(tc.wantLogReq, gotLogReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process got log request (-want,+got):\n%s", diff)
			}
		})
	}
}

type fakeValidator struct {
	returnErr      bool
	justifications []*jvsapi.Justification
}

func (v *fakeValidator) ValidateJWT(jvsToken string) (*jwt.Token, error) {
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
	return &tok, nil
}
