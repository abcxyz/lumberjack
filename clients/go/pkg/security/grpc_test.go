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
	"google.golang.org/grpc/metadata"
)

func TestFromRawJWT_RequestPrincipal(t *testing.T) {
	t.Parallel()

	// Test JWT:
	// {
	// 	 "name": "user",
	// 	 "email": "user@example.com"
	// }
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6InVzZXIiLCJpYXQiOjE1MTYyMzkwMjIsImVtYWlsIjoidXNlckBleGFtcGxlLmNvbSJ9.PXl-SJniWHMVLNYb77HmVFFqWTlu28xf9fou2GaT0Jc"
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
				"authorization": "Bearer " + jwt,
			})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "Bearer ",
			},
			want: "user@example.com",
		},
		{
			name:          "missing_grpc_metadata",
			ctx:           context.Background(),
			wantErrSubstr: "gRPC metadata in incoming context is missing",
		},
		{
			name: "nil_jwt",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.New(nil)),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "Bearer ",
			},
			wantErrSubstr: `nil JWT under the key "authorization" in grpc metadata`,
		},
		{
			name: "unparsable_jwt",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": "Bearer " + "bananas",
			})),
			fromRawJWT: &v1alpha1.FromRawJWT{
				Key:    "authorization",
				Prefix: "Bearer ",
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
