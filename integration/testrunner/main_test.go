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
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/abcxyz/pkg/testutil"
	"google.golang.org/api/idtoken"
)

// Global integration test config.
var cfg *Config

func TestMain(m *testing.M) {
	c, err := newTestConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to parse integration test config: %v", err)
	}
	cfg = c
	os.Exit(m.Run())
}

func TestHTTPEndpoints(t *testing.T) {
	t.Parallel()
	testutil.SkipIfNotIntegration(t)

	testsData := cfg.HTTPEndpoints
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

			ctx := context.Background()
			tc := &TestCase{
				Config:         cfg,
				Endpoint:       test,
				BigQueryClient: makeBigQueryClient(ctx, t, cfg.ProjectID),
			}
			if err := resolveIDToken(tc); err != nil {
				t.Fatalf("Resolving ID Token failed: %v.", err)
			}

			testHTTPEndpoint(ctx, t, tc)
		})
	}
}

func TestGRPCEndpoints(t *testing.T) {
	t.Parallel()
	testutil.SkipIfNotIntegration(t)

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

			ctx := context.Background()
			tc := &TestCase{
				Config:         cfg,
				Endpoint:       test,
				BigQueryClient: makeBigQueryClient(ctx, t, cfg.ProjectID),
			}
			if err := resolveIDToken(tc); err != nil {
				t.Fatalf("Resolving ID Token failed: %v.", err)
			}

			testGRPCEndpoint(ctx, t, tc)
		})
	}
}

func resolveIDToken(tc *TestCase) error {
	if tc.IDToken != "" {
		// ID token was provided via command line flag.
		return nil
	}

	// Attempt getting ID Token from service account if any.
	ts, err := idtoken.NewTokenSource(context.Background(), tc.Endpoint)
	if err != nil {
		return fmt.Errorf("unable to create token source: %w", err)
	}
	t, err := ts.Token()
	if err != nil {
		return fmt.Errorf("unable to get token: %w", err)
	}
	tc.IDToken = t.AccessToken
	return nil
}
