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
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

type HTTPFields struct {
	PrincipalEmail string
	ServiceName    string
}

// testHTTPEndpoint runs the integration tests against a Lumberjack-integrated
// HTTP endpoint.
//
//nolint:thelper // Not really a helper.
func testHTTPEndpoint(ctx context.Context, tb testing.TB, endpointURL, idToken, projectID, datasetQuery string, cfg *Config) {
	// Don't mark t.Helper().
	// Here locates the actual test logic so we want to be able to locate the
	// actual line of error here instead of the main test.

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
	bqClient := makeBigQueryClient(ctx, tb, projectID)
	time.Sleep(10 * time.Second)
	bqQuery := makeQueryForHTTP(*bqClient, id, projectID, datasetQuery)
	tb.Log(bqQuery.Q)
	results := queryIfAuditLogsExists(ctx, tb, bqQuery, cfg, "httpEndpointTest")
	wantNum := 1
	if len(results) != wantNum {
		tb.Errorf("log number doesn't match (-want +got):\n - %d\n + %d\n", wantNum, len(results))
	}
	jsonPayloadInfo := parseJsonpayload(tb, results[0])
	diff := diffResults(jsonPayloadInfo, getMode("HTTP"))
	if diff != "" {
		tb.Errorf(diff)
	}
}

func makeQueryForHTTP(client bigquery.Client, id, projectID, datasetQuery string) *bigquery.Query {
	queryString := fmt.Sprintf(` WITH temptable AS (
		SELECT *
		FROM %s.%s
		WHERE labels.trace_id=?
 	)
 	SELECT TO_JSON(t) as result FROM temptable as t
	`, projectID, datasetQuery)
	return makeQuery(client, id, queryString)
}

func diffResults(jsonPayloadInfo *audit.AuditLog, isHTTPService bool) string {
	wantPrincipalEmail := "gh-access-sa@lumberjack-dev-infra.iam.gserviceaccount.com"
	wantHTTPServiceName := [2]string{"go-shell-app", "java-shell-app"}
	wantGRPCServiceName := "abcxyz.test.Talker"
	diffString := ""
	// if results[0].GetJsonPayload()
	if isHTTPService {
		if jsonPayloadInfo.ServiceName != wantHTTPServiceName[0] && jsonPayloadInfo.ServiceName != wantHTTPServiceName[1] {
			diffString += fmt.Sprintf("- %s or %s\n + %s\n", wantHTTPServiceName[0], wantHTTPServiceName[1], jsonPayloadInfo.ServiceName)
		}
	} else {
		if jsonPayloadInfo.ServiceName != wantGRPCServiceName {
			diffString += fmt.Sprintf("- %sn + %s\n", wantGRPCServiceName, jsonPayloadInfo.ServiceName)
		}
	}
	if jsonPayloadInfo.AuthenticationInfo.PrincipalEmail != wantPrincipalEmail {
		diffString += fmt.Sprintf("- %s\n + %s\n", wantPrincipalEmail, jsonPayloadInfo.AuthenticationInfo.PrincipalEmail)
	}
	if diffString != "" {
		diffString = "queryResult misMatch (-want +got):\n)" + diffString
	}
	return diffString
}

func getMode(s string) bool {
	return s == "HTTP"
}
