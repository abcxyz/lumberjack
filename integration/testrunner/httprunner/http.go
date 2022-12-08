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
	"io"
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

	id := uuid.New().String()
	tb.Logf("using uuid %s", id)

	b := retry.NewExponential(cfg.AuditLogRequestWait)
	if err := retry.Do(ctx, retry.WithMaxRetries(cfg.MaxAuditLogRequestTries, b), func(ctx context.Context) error {
		resp, err := MakeAuditLogRequest(id, endpointURL, cfg.AuditLogRequestTimeout, idToken)
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

	bqClient, err := utils.MakeClient(ctx, projectID)
	if err != nil {
		tb.Fatal(err)
	}
	defer bqClient.Close()

	bqQuery := makeQueryForHTTP(*bqClient, id, projectID, datasetQuery)
	utils.QueryIfAuditLogExistsWithRetries(ctx, tb, bqQuery, cfg, "httpEndpointTest")
}

func makeQueryForHTTP(client bigquery.Client, id string, projectID string, datasetQuery string) *bigquery.Query {
	// Cast to int64 because the result checker expects a number.
	queryString := fmt.Sprintf("SELECT CAST(EXISTS (SELECT * FROM `%s.%s` WHERE labels.trace_id=?", projectID, datasetQuery)
	queryString += ` AND jsonPayload.service_name IS NOT NULL`
	queryString += ` AND jsonPayload.authentication_info.principal_email IS NOT NULL`
	queryString += ") AS INT64)"
	return utils.MakeQuery(client, id, queryString)
}
