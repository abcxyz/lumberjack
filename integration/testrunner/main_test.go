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

// Tests HTTP endpoints provided and verifies logs make it to the BigQuery DB.
package testrunner

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/abcxyz/lumberjack/integration/testrunner/grpcrunner"
	"github.com/abcxyz/lumberjack/integration/testrunner/httprunner"
	"github.com/abcxyz/lumberjack/integration/testrunner/utils"
	"github.com/abcxyz/pkg/testutil"
	"google.golang.org/api/idtoken"
)

var (
	idTokenPtr      = flag.String("id-token", "", `Identity token, can be obtained with "gcloud auth print-identity-token", can be omitted if service account key is provided.`)
	projectIDPtr    = flag.String("project-id", "", "Cloud project ID of which the Database will be queried.")
	datasetQueryPtr = flag.String("dataset-query", "", "BigQuery dataset query string to get the audit logs.")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func validateCfg(t *testing.T) *utils.Config {
	t.Helper()

	if *projectIDPtr == "" {
		t.Fatal("Cloud Project ID of the Database to query must be provided with the -project-id flag.")
	}
	if *datasetQueryPtr == "" {
		t.Fatal("BigQuery dataset query string must be provided with the -dataset-query flag.")
	}

	cfg, err := utils.NewConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	return cfg
}

func TestHTTPEndpoints(t *testing.T) {
	t.Parallel()
	testutil.SkipIfNotIntegration(t)

	cfg := validateCfg(t)
	testsData := cfg.HTTPEndpoints
	var tests []string
	if err := json.Unmarshal([]byte(testsData), &tests); err != nil {
		t.Fatalf("Unable to parse HTTP endpoints: %v.", err)
	}

	for i, tc := range tests {
		tc := tc

		t.Run(tc, func(t *testing.T) {
			t.Parallel()

			if tc == "" {
				t.Fatalf("URL for test with index %v not found.", i)
			}

			idToken, err := resolveIDToken(tc)
			if err != nil {
				t.Fatalf("Resolving ID Token failed: %v.", err)
			}

			ctx := context.Background()
			httprunner.CheckHTTPEndpoint(ctx, t, tc, idToken, *projectIDPtr, *datasetQueryPtr, cfg)
		})
	}
}

func TestGRPCEndpoints(t *testing.T) {
	t.Parallel()
	testutil.SkipIfNotIntegration(t)

	cfg := validateCfg(t)
	testsData := cfg.GRPCEndpoints
	var tests []string
	if err := json.Unmarshal([]byte(testsData), &tests); err != nil {
		t.Fatalf("Unable to parse HTTP endpoints: %v.", err)
	}

	for i, test := range tests {
		i, test := i, test

		t.Run(test, func(t *testing.T) {
			t.Parallel()

			if test == "" {
				t.Fatalf("URL for test with index %v not found.", i)
			}

			idToken, err := resolveIDToken(test)
			if err != nil {
				t.Fatalf("Resolving ID Token failed: %v.", err)
			}

			ctx := context.Background()
			grpcrunner.CheckGRPCEndpoint(ctx, t, &grpcrunner.GRPC{
				ProjectID:    *projectIDPtr,
				DatasetQuery: *datasetQueryPtr,

				IDToken:     idToken,
				EndpointURL: test,

				Config:               cfg,
				RequireJustification: true,
			})
		})
	}
}

// resolveIDToken Resolves the ID token passed via the "id-token" flag if provided,
// otherwise looks for the ID token from the provided service account, if any.
func resolveIDToken(endpointURL string) (string, error) {
	if *idTokenPtr != "" {
		// ID token was provided via command line flag.
		return *idTokenPtr, nil
	}

	// Attempt getting ID Token from service account if any.
	ts, err := idtoken.NewTokenSource(context.Background(), endpointURL)
	if err != nil {
		return "", fmt.Errorf("unable to create token source: %w", err)
	}
	t, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("unable to get token: %w", err)
	}
	return t.AccessToken, nil
}
