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

package testrunner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
)

// testHTTPEndpoint runs the integration tests against a Lumberjack-integrated
// HTTP endpoint.
//
//nolint:thelper // Not really a helper.
func testHTTPEndpoint(ctx context.Context, tb testing.TB, tc *TestCase) {
	// Don't mark t.Helper().
	// Here locates the actual test logic so we want to be able to locate the
	// actual line of error here instead of the main test.

	tc.TraceID = uuid.New().String()
	tb.Logf("Using trace ID: %s", tc.TraceID)

	b := retry.NewConstant(tc.AuditLogRequestWait)
	if err := retry.Do(ctx, retry.WithMaxRetries(tc.MaxAuditLogRequestTries, b), func(ctx context.Context) error {
		resp, err := makeHTTPAuditLogRequest(tc)
		if err != nil {
			return fmt.Errorf("failed to make audit log request: %w", err)
		}
		defer resp.Body.Close()

		// Audit log request succeeded, exit the retry logic with success.
		if resp.StatusCode == http.StatusOK {
			return nil
		}

		b, err := io.ReadAll(io.LimitReader(resp.Body, 64*1_000))
		if err != nil {
			err = fmt.Errorf("failed to read response body: %w", err)
			return retry.RetryableError(err)
		}

		tb.Logf("bad response (%d):\n\n%s\n\n",
			resp.StatusCode, string(b))

		return retry.RetryableError(fmt.Errorf("audit logging failed, retrying"))
	}); err != nil {
		tb.Fatal(err)
	}
	bqQuery := makeQueryForHTTP(tc)
	validateAuditLogsWithRetries(ctx, tb, tc, bqQuery, 1)
}

func makeQueryForHTTP(tc *TestCase) *bigquery.Query {
	queryString := fmt.Sprintf(` WITH temptable AS (
		SELECT *
		FROM %s.%s
		WHERE labels.trace_id=?
 	)
 	SELECT TO_JSON(t) as result FROM temptable as t
	`, tc.ProjectID, tc.BigQueryDataset)
	return makeQuery(tc.BigQueryClient, tc.TraceID, queryString)
}

func makeHTTPAuditLogRequest(tc *TestCase) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit log http request: %w", err)
	}

	signedToken, err := justificationToken("logging-shell", tc.JustificationSubject)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate justification token: %w", err)
	}
	req.Header.Set("justification-token", signedToken)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.IDToken))

	// Insert the UUID used in tracing the log as a query parameter.
	q := req.URL.Query()
	q.Add("trace_id", tc.TraceID)
	req.URL.RawQuery = q.Encode()

	httpClient := &http.Client{Timeout: tc.AuditLogRequestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute audit log request: %w", err)
	}
	return resp, nil
}
