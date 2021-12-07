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
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
	"google.golang.org/api/idtoken"
)

var (
	idTokenPtr      = flag.String("id-token", "", `Identity token, can be obtained with "gcloud auth print-identity-token", can be omitted if service account key is provided.`)
	projectIDPtr    = flag.String("project-id", "", "Cloud project ID of which the Database will be queried.")
	datasetQueryPtr = flag.String("dataset-query", "", "BigQuery dataset query string to get the audit logs.")
	cfg             *Config
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *projectIDPtr == "" {
		log.Fatal("Cloud Project ID of the Database to query must be provided with the -project-id flag.")
	}
	if *datasetQueryPtr == "" {
		log.Fatal("BigQuery dataset query string must be provided with the -dataset-query flag.")
	}

	var err error
	if cfg, err = NewConfig(context.Background()); err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestEndpoints(t *testing.T) {
	t.Parallel()
	testsData := cfg.HttpEndpoints
	var tests []string
	if err := json.Unmarshal([]byte(testsData), &tests); err != nil {
		t.Fatalf("Unable to parse HTTP endpoints: %v.", err)
	}

	for i, test := range tests {
		test := test
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			if test == "" {
				t.Fatalf("URL for test with index %v not found.", i)
			}
			ctx := context.Background()
			testEndpoint(t, ctx, test)
		})
	}
}

func testEndpoint(t *testing.T, ctx context.Context, endpointURL string) {
	u := uuid.New()
	t.Logf("Generated UUID: %s.", u.String())

	idToken, err := resolveIDToken(endpointURL)
	if err != nil {
		t.Fatalf("Resolving ID Token failed: %v.", err)
	}

	b, err := retry.NewExponential(cfg.AuditLogRequestWait)
	if err != nil {
		t.Fatalf("Retry logic setup failed: %v.", err)
	}

	if err = retry.Do(ctx, retry.WithMaxRetries(cfg.MaxAuditLogRequestTries, b), func(ctx context.Context) error {
		resp, err := makeAuditLogRequest(u, endpointURL, cfg.AuditLogRequestTimeout, idToken)
		if err != nil {
			t.Logf("Audit log request failed: %v.", err)
		}

		if resp.StatusCode == http.StatusOK {
			// Audit log request succeeded, exit the retry logic with success.
			return nil
		}

		t.Logf("Audit log failed with status: %v.", resp.Status)
		return retry.RetryableError(fmt.Errorf("Audit logging failed, retrying."))
	}); err != nil {
		t.Fatalf("Retry failed: %v.", err)
	}

	bqClient, bqQuery, err := makeClientAndQuery(ctx, u, *projectIDPtr, *datasetQueryPtr)
	if err != nil {
		t.Fatalf("BigQuery request failed: %v.", err)
	}

	defer func() {
		if err := bqClient.Close(); err != nil {
			t.Logf("Failed to close the BQ client: %v.", err)
		}
	}()

	b, err = retry.NewExponential(cfg.LogRoutingWait)
	if err != nil {
		t.Fatalf("Retry logic setup failed: %v.", err)
	}

	if err = retry.Do(ctx, retry.WithMaxRetries(cfg.MaxDBQueryTries, b), func(ctx context.Context) error {
		found, err := queryIfAuditLogExists(ctx, bqQuery)
		if found {
			// Early exit retry if queried log already found.
			return nil
		}

		t.Log("Matching entry not found, retrying...")

		if err != nil {
			t.Logf("Query error: %v.", err)
		}
		return retry.RetryableError(fmt.Errorf("audit log not found"))
	}); err != nil {
		t.Errorf("Retry failed: %v.", err)
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
		return "", fmt.Errorf("unable to create token source: %v", err)
	}
	t, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("unable to get token: %v", err)
	}
	return t.AccessToken, nil
}
