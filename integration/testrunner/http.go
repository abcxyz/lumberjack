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
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
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
	wantPrincipalEmail := "gh-access-sa@lumberjack-dev-infra.iam.gserviceaccount.com"
	wantServiceName := [2]string{"go-shell-app", "java-shell-app"}

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
	time.Sleep(10 * time.Second)
	bqClient := makeBigQueryClient(ctx, tb, projectID)

	bqQuery := makeQueryForHTTP(*bqClient, id, projectID, datasetQuery)
	tb.Log(bqQuery.Q)
	value := queryIfAuditLogExistsWithRetries(ctx, tb, bqQuery, cfg, "httpEndpointTest")
	result := parseQueryResultForHTTP(tb, value)
	tb.Log(result)
	tb.Log(result.PrincipalEmail)
	tb.Log(result.ServiceName)
	diffString := ""
	if result.ServiceName != wantServiceName[0] && result.ServiceName != wantServiceName[1] {
		diffString += fmt.Sprintf("- %s or %s\n + %s\n", wantServiceName[0], wantServiceName[1], result.ServiceName)
	}
	if result.PrincipalEmail != wantPrincipalEmail {
		diffString += fmt.Sprintf("- %s\n + %s\n", wantPrincipalEmail, result.PrincipalEmail)
	}
	if diffString != "" {
		diffString = "queryResult misMatch (-want +got):\n)" + diffString
		tb.Errorf("%s", diffString)
	}
}

func makeQueryForHTTP(client bigquery.Client, id, projectID, datasetQuery string) *bigquery.Query {
	queryString := "SELECT "
	queryString += fmt.Sprintf("%s as %s, ", "jsonPayload.authentication_info.principal_email", "PrincipalEmail")
	queryString += fmt.Sprintf("%s as %s,", "jsonPayload.service_name", "ServiceName")
	queryString += fmt.Sprintf("FROM `%s.%s` WHERE labels.trace_id='%s'", projectID, datasetQuery, id)
	return makeQuery(client, queryString)
}

// Parse bigquey.Value type into HttpFields, so we can use that to do diff.
func parseQueryResultForHTTP(tb testing.TB, value []bigquery.Value) HTTPFields {
	// The value paramerter is returned from a query to bigquery
	// and the format of that would be like
	// [SomePrincipalEmail SomeServiceName]
	// We want to parse that into a varible with type HTTPFields
	// with the format of
	// result := HTTPFields{
	//  PrincipalEmail: "SomePrincipalEmail"
	//  ServiceName: "SomeServiceName"
	// }.
	tb.Helper()
	result := HTTPFields{}
	elem := reflect.ValueOf(&result).Elem()
	for i, v := range value {
		result, ok := v.(string)
		if !ok {
			err := fmt.Errorf("error converting query results to string (got %T)", v)
			tb.Log(err)
		}
		elem.Field(i).SetString(result)
	}
	return result
}
