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
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// Matching private key here: https://github.com/abcxyz/lumberjack/blob/92782c326681157221df37e0897ba234c5a22240/integration/testrunner/grpcrunner/grpc.go#L60
// const publicKeyString = `
// -----BEGIN PUBLIC KEY-----
// MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEhBWj8vw5LkPRWbCr45k0cOarIcWg
// ApM03mSYF911de5q1wGOL7R9N8pC7jo2xbS+i1wGsMiz+AWnhhZIQcNTKg==
// -----END PUBLIC KEY-----
// `

// StartLocalPublicKeyServer parse pre-made key and set up a server to host it in JWKS format.
// This is intended to stand in for the JVS in the integration tests.
type jsonData struct {
	Encoded string
	Decoded string
}

func loadJSON() jsonData {
	jsonFile, err := os.Open("/etc/lumberjack/public_key.json")
	if err != nil {
		fmt.Println("Error opening files")
	}
	byteValue, _ := io.ReadAll(jsonFile)
	var tmp jsonData
	err = json.Unmarshal(byteValue, &tmp)
	if err != nil {
		fmt.Println("Error parsing files")
	}
	return tmp
}

func StartLocalPublicKeyServer() (string, func(), error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(loadJSON().Encoded)))
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Printf("Err when parsing key %v", err)
		return "", nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	ecdsaKey, err := jwk.FromRaw(key)
	if err != nil {
		log.Printf("Err when converting key to jwk %v", err)
		return "", nil, fmt.Errorf("failed to parse jwk: %w", err)
	}
	if err := ecdsaKey.Set(jwk.KeyIDKey, "integ-key"); err != nil {
		log.Printf("Err when setting key id %v", err)
		return "", nil, fmt.Errorf("failed to set key id: %w", err)
	}

	jwks := make(map[string][]jwk.Key)
	jwks["keys"] = []jwk.Key{ecdsaKey}
	j, err := json.MarshalIndent(jwks, "", " ")
	if err != nil {
		log.Printf("Err when creating jwks json %v", err)
		return "", nil, fmt.Errorf("failed to marshal jwks: %w", err)
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
