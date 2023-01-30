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
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/sethvargo/go-retry"
)

// queryIfAuditLogExists queries the DB and checks if audit log contained in the query exists or not.
func queryIfAuditLogExists(ctx context.Context, tb testing.TB, query *bigquery.Query) ([]bigquery.Value, error) {
	tb.Helper()

	job, err := query.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %w", err)
	}

	if status, err := job.Wait(ctx); err != nil {
		return nil, fmt.Errorf("failed to wait for query: %w", err)
	} else if err = status.Err(); err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read job: %w", err)
	}
	var row []bigquery.Value
	if err = it.Next(&row); err != nil {
		return nil, fmt.Errorf("failed to get next row: %w", err)
	}
	return row, nil
}

func makeQuery(bqClient bigquery.Client, id, queryString string) *bigquery.Query {
	bqQuery := bqClient.Query(queryString)
	if id == "" {
		fmt.Print("hahaha")
	}
	return bqQuery
}

// makeBigQueryClient creates a new client and automatically closes the
// connection when the tests finish.
func makeBigQueryClient(ctx context.Context, tb testing.TB, projectID string) *bigquery.Client {
	tb.Helper()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		if err := client.Close(); err != nil {
			tb.Errorf("failed to close the biquery client: %v", err)
		}
	})

	return client
}

// This calls the database to check that an audit log exists. It uses the retries that are specified in the Config
// file. This method assumes that only a single audit log will match, which constitutes success.
func queryIfAuditLogExistsWithRetries(ctx context.Context, tb testing.TB, bqQuery *bigquery.Query, cfg *Config, testName string) []bigquery.Value {
	tb.Helper()

	return queryIfAuditLogsExistWithRetries(ctx, tb, bqQuery, cfg, testName)
}

// This calls the database to check that an audit log exists. It uses the retries that are specified in the Config
// file. This method allows for specifying how many logs we expect to match, in order to handle streaming use cases.
func queryIfAuditLogsExistWithRetries(ctx context.Context, tb testing.TB, bqQuery *bigquery.Query, cfg *Config, testName string) []bigquery.Value {
	tb.Helper()
	var result []bigquery.Value
	b := retry.NewExponential(cfg.LogRoutingWait)
	if err := retry.Do(ctx, retry.WithMaxRetries(cfg.MaxDBQueryTries, b), func(ctx context.Context) error {
		row, err := queryIfAuditLogExists(ctx, tb, bqQuery)
		if row != nil {
			// Early exit retry if queried log already found.
			result = row
			return nil
		}

		tb.Log("Matching entry not found, retrying...")

		if err != nil {
			tb.Logf("query error: %s", err.Error())
		}
		return retry.RetryableError(fmt.Errorf("no matching audit log found in bigquery after timeout for %q", testName))
	}); err != nil {
		tb.Errorf("retry failed: %v.", err)
	}
	return result
}
