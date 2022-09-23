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

package testutil

import (
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const signingString = "this-is-definitely-not-a-secret-string-for-signing-jwts-look-the-other-way-please-and-thank-you"

// JWTFromClaims is a testing helper that builds a JWT from the given claims.
func JWTFromClaims(tb testing.TB, claims map[string]any) string {
	tb.Helper()

	tokenBuilder := jwt.NewBuilder()
	for k, v := range claims {
		tokenBuilder = tokenBuilder.Claim(k, v)
	}

	token, err := tokenBuilder.Build()
	if err != nil {
		tb.Fatal(err)
	}

	b, err := jwt.Sign(token, jwt.WithKey(jwa.HS512, []byte(signingString)))
	if err != nil {
		tb.Fatal(err)
	}
	return string(b)
}
