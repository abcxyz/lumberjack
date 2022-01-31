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

package httprunner

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/abcxyz/lumberjack/integration/testrunner/utils"
	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
)

func TestHttpEndpoint(t *testing.T, ctx context.Context, endpointURL string,
	idToken string, projectID string, datasetQuery string, cfg *utils.Config) {
	u := uuid.New()
	t.Logf("Generated UUID: %s.", u.String())

	b, err := retry.NewExponential(cfg.AuditLogRequestWait)
	if err != nil {
		t.Fatalf("Retry logic setup failed: %v.", err)
	}

	if err = retry.Do(ctx, retry.WithMaxRetries(cfg.MaxAuditLogRequestTries, b), func(ctx context.Context) error {
		resp, err := MakeAuditLogRequest(u, endpointURL, cfg.AuditLogRequestTimeout, idToken)
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

	bqClient, err := utils.MakeClient(ctx, projectID)
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

	bqQuery := makeQueryForHttp(*bqClient, u, projectID, datasetQuery)
	if err = retry.Do(ctx, retry.WithMaxRetries(cfg.MaxDBQueryTries, b), func(ctx context.Context) error {
		found, err := utils.QueryIfAuditLogExists(ctx, bqQuery)
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

func makeQueryForHttp(client bigquery.Client, u uuid.UUID, projectID string, datasetQuery string) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM (SELECT * FROM %s.%s WHERE labels.trace_id=? LIMIT 1)", projectID, datasetQuery)
	return utils.MakeQuery(client, u, queryString)
}
