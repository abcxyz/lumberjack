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

package utils

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
)

// queryIfAuditLogExists queries the DB and checks if audit log contained in the query exists or not.
func QueryIfAuditLogExists(ctx context.Context, query *bigquery.Query) (bool, error) {
	job, err := query.Run(ctx)
	if err != nil {
		return false, err
	}

	if status, err := job.Wait(ctx); err != nil {
		return false, err
	} else if err = status.Err(); err != nil {
		return false, err
	}

	it, err := job.Read(ctx)
	if err != nil {
		return false, err
	}

	var row []bigquery.Value
	if err = it.Next(&row); err != nil {
		return false, err
	}

	// Check if the matching row count is equal to 1, if yes, then the audit log exists.
	return row[0] == int64(1), nil
}

func MakeQuery(bqClient bigquery.Client, u uuid.UUID, queryString string) *bigquery.Query {
	bqQuery := bqClient.Query(queryString)
	bqQuery.Parameters = []bigquery.QueryParameter{{Value: u.String()}}
	return bqQuery
}

func MakeClient(ctx context.Context, projectID string) (*bigquery.Client, error) {
	return bigquery.NewClient(ctx, projectID)
}

func QueryIfAuditLogExistsWithRetries(t testing.TB, ctx context.Context, bqQuery *bigquery.Query, cfg *Config) {
	b, err := retry.NewExponential(cfg.LogRoutingWait)
	if err != nil {
		t.Fatalf("Retry logic setup failed: %v.", err)
	}

	if err = retry.Do(ctx, retry.WithMaxRetries(cfg.MaxDBQueryTries, b), func(ctx context.Context) error {
		found, err := QueryIfAuditLogExists(ctx, bqQuery)
		if found {
			// Early exit retry if queried log already found.
			return nil
		}

		t.Log("Matching entry not found, retrying...")

		if err != nil {
			t.Logf("Query error: %v.", err)
		}
		return retry.RetryableError(fmt.Errorf("no matching audit log found in bigquery after timeout"))
	}); err != nil {
		t.Errorf("Retry failed: %v.", err)
	}
}
