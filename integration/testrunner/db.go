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
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sethvargo/go-retry"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/encoding/protojson"

	logpb "cloud.google.com/go/logging/apiv2/loggingpb"
)

// queryAuditLogs queries the DB and checks if audit log contained in the query exists or not and return the results.
func queryAuditLogs(ctx context.Context, tb testing.TB, query *bigquery.Query) ([]*logpb.LogEntry, error) {
	tb.Helper()
	time.Sleep(25 * time.Second)
	job, err := query.Run(ctx)
	if err != nil {
		tb.Logf("failed to run query: %s", err.Error())
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
	var logEntries []*logpb.LogEntry
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			tb.Logf("failed to get next row")
			return nil, fmt.Errorf("failed to get next row: %w", err)
		}
		pbJSON := &logpb.LogEntry{}
		value, ok := row[0].(string)
		if !ok {
			tb.Logf("error converting query results to string")
			return nil, fmt.Errorf("error converting query results to string (got %T)", value[0])
		}
		// When there are fields with null values in json string
		// protojson.Unmarshal will return an error: invalid value for string type: null
		// but the json string can still be unmarshal into pbJSON.
		if err := protojson.Unmarshal([]byte(value), pbJSON); err != nil {
			tb.Logf("ignoring error: %s as this behavior is expected", err.Error())
		}
		logEntries = append(logEntries, pbJSON)
	}
	return logEntries, nil
}

func makeQuery(bqClient bigquery.Client, id, queryString string) *bigquery.Query {
	bqQuery := bqClient.Query(queryString)
	bqQuery.Parameters = []bigquery.QueryParameter{{Value: id}}
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
// file. This method allows for specifying how many logs we expect to match, in order to handle streaming use cases.
func queryIfAuditLogsExistsWithRetries(ctx context.Context, tb testing.TB, bqQuery *bigquery.Query, cfg *Config, testName string) []*logpb.LogEntry {
	tb.Helper()
	var logEntries []*logpb.LogEntry
	b := retry.NewExponential(cfg.LogRoutingWait)
	if err := retry.Do(ctx, retry.WithMaxRetries(cfg.MaxDBQueryTries, b), func(ctx context.Context) error {
		rows, err := queryAuditLogs(ctx, tb, bqQuery)
		if rows != nil {
			// Early exit retry if queried log already found.
			logEntries = rows
			return nil
		}

		tb.Log("Matching entry not found, retrying...")

		if err != nil {
			return retry.RetryableError(fmt.Errorf("query error: %w", err))
		}
		return retry.RetryableError(fmt.Errorf("no matching audit log found in bigquery after timeout for %q", testName))
	}); err != nil {
		tb.Errorf("retry failed: %v.", err)
	}
	return logEntries
}

func parseJsonpayload(tb testing.TB, logEntry *logpb.LogEntry) *audit.AuditLog {
	tb.Helper()
	jsonPayloadContent := logEntry.GetJsonPayload()
	jsonPayloadString, err := jsonPayloadContent.MarshalJSON()
	if err != nil {
		err := fmt.Errorf("error parsing *pbstruct.Struct to json: %w)", err)
		tb.Log(err)
	}
	var auditLog audit.AuditLog
	if err := json.Unmarshal(jsonPayloadString, &auditLog); err != nil {
		err := fmt.Errorf("error parsing json to AuditLog: %w)", err)
		tb.Logf(err.Error())
	}
	return &auditLog
}
