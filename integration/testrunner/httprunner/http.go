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

func TestHTTPEndpoint(ctx context.Context, tb testing.TB, endpointURL string,
	idToken string, projectID string, datasetQuery string, cfg *utils.Config,
) {
	tb.Helper()

	u := uuid.New()
	tb.Logf("Generated UUID: %s", u.String())

	b := retry.NewExponential(cfg.AuditLogRequestWait)
	if err := retry.Do(ctx, retry.WithMaxRetries(cfg.MaxAuditLogRequestTries, b), func(ctx context.Context) error {
		resp, err := MakeAuditLogRequest(u, endpointURL, cfg.AuditLogRequestTimeout, idToken)
		if err != nil {
			tb.Logf("audit log request failed: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				tb.Log(err)
			}
		}()

		if resp.StatusCode == http.StatusOK {
			// Audit log request succeeded, exit the retry logic with success.
			return nil
		}

		tb.Logf("Audit log failed with status: %v.", resp.Status)
		return retry.RetryableError(fmt.Errorf("audit logging failed, retrying"))
	}); err != nil {
		tb.Fatalf("Retry failed: %v.", err)
	}

	bqClient, err := utils.MakeClient(ctx, projectID)
	if err != nil {
		tb.Fatalf("BigQuery request failed: %v.", err)
	}

	defer func() {
		if err := bqClient.Close(); err != nil {
			tb.Logf("Failed to close the BQ client: %v.", err)
		}
	}()

	bqQuery := makeQueryForHTTP(*bqClient, u, projectID, datasetQuery)
	utils.QueryIfAuditLogExistsWithRetries(ctx, tb, bqQuery, cfg, "httpEndpointTest")
}

func makeQueryForHTTP(client bigquery.Client, u uuid.UUID, projectID string, datasetQuery string) *bigquery.Query {
	// Cast to int64 because the result checker expects a number.
	queryString := fmt.Sprintf("SELECT CAST(EXISTS (SELECT * FROM %s.%s WHERE labels.trace_id=?) AS INT64)", projectID, datasetQuery)
	return utils.MakeQuery(client, u, queryString)
}
