// Copyright 2023 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
)

// StartLocalPublicKeyServer parse pre-made key and set up a server to host it in JWKS format.
func StartLocalPublicKeyServer() (string, func(), error) {
	j, err := os.ReadFile("test_public_key.key")
	if err != nil {
		return "", nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	path := "/.well-known/jwks"
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", j)
	})
	svr := httptest.NewServer(mux)
	return svr.URL + path, func() { svr.Close() }, nil
}
