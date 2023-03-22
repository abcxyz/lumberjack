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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

// StartLocalPublicKeyServer parse pre-made key and set up a server to host it in JWKS format.
// This is intended to stand in for the JVS in the integration tests.
// type publicKeyJSONData struct {
// 	Encoded string
// }

// func loadJSON() (*publicKeyJSONData, error) {
// 	var data publicKeyJSONData
// 	path := "/usr/local/google/home/qinhang/go/src/github.com/spc/lumberjack/integration/testrunner/public_key.json"
// 	jsonFile, err := os.Open(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open file: %w", err)
// 	}
// 	b, err := io.ReadAll(jsonFile)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read from file: %w", err)
// 	}
// 	if err = json.Unmarshal(b, &data); err != nil {
// 		return nil, fmt.Errorf("failed to parse files: %w", err)
// 	}
// 	return &data, nil
// }

func readBytes() (*[]byte, error) {
	path := "decoded_public_key.pub"
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read from file: %w", err)
	}
	return &b, nil
}

func StartLocalPublicKeyServer() (string, func(), error) {
	// publicKeyStr, err := loadJSON()
	// if err != nil {
	// 	return "", nil, fmt.Errorf("failed to load public key: %w", err)
	// }
	// // TODO: Enable this code to use decoded public key
	// // https://github.com/abcxyz/lumberjack/issues/406
	// block, _ := pem.Decode([]byte(strings.TrimSpace(publicKeyStr.Encoded)))
	// key, err := x509.ParsePKIXPublicKey(block.Bytes)
	// if err != nil {
	// 	return "", nil, fmt.Errorf("failed to parse public key: %w", err)
	// }
	// ecdsaKey, err := jwk.FromRaw(key)
	// if err != nil {
	// 	return "", nil, fmt.Errorf("failed to parse jwk: %w", err)
	// }
	// if err := ecdsaKey.Set(jwk.KeyIDKey, "integ-key"); err != nil {
	// 	return "", nil, fmt.Errorf("failed to set key id: %w", err)
	// }

	// jwks := make(map[string][]jwk.Key)
	// jwks["keys"] = []jwk.Key{ecdsaKey}
	// j, err := json.MarshalIndent(jwks, "", " ")
	j, err := readBytes()
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal jwks: %w", err)
	}

	path := "/.well-known/jwks"
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", *j)
	})
	svr := httptest.NewServer(mux)
	return svr.URL + path, func() { svr.Close() }, nil
}
