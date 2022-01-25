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

// Package security describes the authentication technology that the
// middleware investigates to autofill the principal in a log request.
package security

import (
	"context"
	"testing"

	"github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/metadata"
)

func TestFromRawJWT_RequestPrincipal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ctx           context.Context
		fromRawJWT    *v1alpha1.FromRawJWT
		want          string
		wantErrSubstr string
	}{
		{
			name: "valid_jwt",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "Bearer " + jwtFromClaims(t, map[string]interface{}{
					"email": "user@example.com",
				}),
			})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "Bearer ",
			},
			want: "user@example.com",
		},
		{
			name: "error_from_missing_jwt_email_claim",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "Bearer " + jwtFromClaims(t, map[string]interface{}{}),
			})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "Bearer ",
			},
			wantErrSubstr: `jwt claims are missing the email key "email"`,
		},
		{
			name: "error_from_slice_as_jwt_email_claim",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "Bearer " + jwtFromClaims(t, map[string]interface{}{
					"email": []string{"foo", "bar"},
				}),
			})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "Bearer ",
			},
			wantErrSubstr: `expecting string in jwt claims "email", got []interface {}`,
		},
		{
			name:          "error_from_missing_grpc_metadata",
			ctx:           context.Background(),
			wantErrSubstr: "gRPC metadata in incoming context is missing",
		},
		{
			name: "error_from_nil_receiver_field",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "Bearer " + jwtFromClaims(t, map[string]interface{}{
					"email": "user@example.com",
				}),
			})),
			wantErrSubstr: `expecting non-nil receiver field "j.FromRawJwt"`,
		},
		{
			name: "error_from_inexistent_jwt_key",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "Bearer ",
			},
			wantErrSubstr: `failed extracting the JWT due missing key "authorization" in grpc metadata`,
		},

		{
			name: "error_from_slice_length_two_in_jwt_key",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{
				"authorization": []string{"foo", "bar"},
			}),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "Bearer ",
			},
			wantErrSubstr: `expecting exaclty one value (a JWT) under key "authorization" in grpc metadata`,
		},

		{
			name: "error_from_prefix_longer_than_jwt",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "short",
			})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "loooooong",
			},
			wantErrSubstr: `JWT prefix "loooooong" is longer than raw JWT "short"`,
		},
		{
			name: "error_from_empty_string_as_jwt",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "",
			})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "",
			},
			wantErrSubstr: `nil JWT ID token under the key "authorization" in grpc metadata`,
		},
		{
			name: "error_from_unparsable_jwt",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "bananas",
			})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "",
			},
			wantErrSubstr: "unable to parse JWT",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			j := &FromRawJWT{FromRawJWT: tc.fromRawJWT}
			got, err := j.RequestPrincipal(tc.ctx)
			if diff := errutil.DiffSubstring(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("j.RequestPrincipal()) got unexpected error substring: %v", diff)
			}

			if got != tc.want {
				t.Errorf("j.RequestPrincipal() = %v, want %v", got, tc.want)
			}
		})
	}
}

func jwtFromClaims(t *testing.T, claims map[string]interface{}) string {
	jwtMapClaims := jwt.MapClaims{}
	jwtMapClaims = claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtMapClaims)
	signedToken, err := token.SignedString([]byte("secureSecretText"))
	if err != nil {
		t.Fatal(err)
	}
	return signedToken
}
