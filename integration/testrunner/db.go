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
	"net/mail"
	"testing"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/sethvargo/go-retry"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/encoding/protojson"
)

// queryAuditLogs queries the DB and checks if audit log contained in the query exists or not and return the results.
// Since validateAuditLogsWithRetries is the only caller now, to make code clean, we won't return error here.
// For retryable error, log the error. For non-retryable errors, fail now.
func queryAuditLogs(ctx context.Context, tb testing.TB, query *bigquery.Query) []*loggingpb.LogEntry {
	tb.Helper()
	job, err := query.Run(ctx)
	if err != nil {
		tb.Logf("failed to run query: %v", err)
		return nil
	}

	if status, err := job.Wait(ctx); err != nil {
		tb.Logf("failed to wait for query: %v", err)
		return nil
	} else if err = status.Err(); err != nil {
		tb.Logf("query failed: %v", err)
		return nil
	}
	it, err := job.Read(ctx)
	if err != nil {
		tb.Logf("failed to read job: %v", err)
		return nil
	}
	var logEntries []*loggingpb.LogEntry
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			tb.Fatalf("failed to get next row: %v", err)
		}
		value, ok := row[0].(string)
		if !ok {
			tb.Fatalf("error converting query (%T) to string: %v", value[0], err)
		}
		tb.Logf("bq row is %s", value)
		logEntry := &loggingpb.LogEntry{}
		if err := protojson.Unmarshal([]byte(value), logEntry); err != nil {
			tb.Fatalf("error when unmarshal bq row to logEntry: %v", err)
		}
		logEntries = append(logEntries, logEntry)
	}
	return logEntries
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

// This calls the database to validate that an audit log exists with expected format
// and validate how many logs we expect to match, in order to handle streaming use cases.
// It uses the retries that are specified in the Config file.
func validateAuditLogsWithRetries(ctx context.Context, tb testing.TB, bqQuery *bigquery.Query, cfg *Config, wantNum int, requireJustification bool) {
	tb.Helper()
	tb.Logf("querying BigQuery:\n%s", bqQuery.Q)
	var logEntries []*loggingpb.LogEntry
	b := retry.NewConstant(cfg.LogRoutingWait)
	if err := retry.Do(ctx, retry.WithMaxRetries(cfg.MaxDBQueryTries, b), func(ctx context.Context) error {
		results := queryAuditLogs(ctx, tb, bqQuery)
		// Early exit retry if queried log already found.
		if len(results) == wantNum {
			logEntries = results
			return nil
		}
		if len(results) > wantNum {
			tb.Fatalf("log number doesn't match (-want +got):\n - %d\n + %d\n", wantNum, len(results))
		}
		tb.Log("Matching entry not found, retrying...")
		return retry.RetryableError(fmt.Errorf("no matching audit log found in bigquery after timeout"))
	}); err != nil {
		tb.Fatalf("retry failed: %v.", err)
	}
	for i, logEntry := range logEntries {
		tb.Logf("diffing LogEntry for index %v", i)
		diffLogEntry(tb, logEntry, requireJustification)
	}
}

func parseJSONPayload(tb testing.TB, logEntry *loggingpb.LogEntry) (*audit.AuditLog, error) {
	tb.Helper()
	jsonPayloadBytes, err := logEntry.GetJsonPayload().MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling *structpb.Struct to json: %w", err)
	}
	var auditLog audit.AuditLog
	if err := protojson.Unmarshal(jsonPayloadBytes, &auditLog); err != nil {
		return nil, fmt.Errorf("error parsing json %v to AuditLog: %w", string(jsonPayloadBytes), err)
	}
	return &auditLog, nil
}

// TODO: https://github.com/abcxyz/lumberjack/issues/381
// Make test logs consistent and
// see whether we can compare the whole log entry instead of validating specific fields.
func diffLogEntry(tb testing.TB, logEntry *loggingpb.LogEntry, requireJustification bool) {
	tb.Helper()

	tb.Logf("Got logEntry is %s", logEntry.String())

	jsonPayloadInfo, err := parseJSONPayload(tb, logEntry)
	if err != nil {
		tb.Fatalf("failed to get jsonPayload from logEntry: %v", err)
	}

	if logEntry.LogName == "" {
		tb.Errorf("queryResult field %v is blank", "logName")
	}

	if logEntry.Timestamp == nil {
		tb.Errorf("queryResult field %v is blank", "timestamp")
	}
	if !isValidEmail(jsonPayloadInfo.AuthenticationInfo.PrincipalEmail) {
		tb.Errorf("queryResult field %v is invalid, got %v", "jsonPayload.authentication_info.principal_email", jsonPayloadInfo.AuthenticationInfo.PrincipalEmail)
	}
	if jsonPayloadInfo.ServiceName == "" {
		tb.Errorf("queryResult field %v is blank", "jsonPayload.service_name")
	}
	if jsonPayloadInfo.MethodName == "" {
		tb.Errorf("queryResult field %v is blank", "jsonPayload.method_name")
	}

	if requireJustification {
		checkJustification(tb, jsonPayloadInfo)
	}
}

func checkJustification(tb testing.TB, jsonPayloadInfo *audit.AuditLog) {
	tb.Helper()
	justification, ok := jsonPayloadInfo.Metadata.AsMap()["justification"]
	if !ok {
		tb.Fatalf("queryResult field %v doesn't exist", "jsonPayload.metadata.justification")
	}
	b, err := json.Marshal(justification)
	if err != nil {
		tb.Fatalf("failed to marshal justification: %v", err)
	}
	if string(b) == "null" {
		tb.Errorf("queryResult field %v is blank", "jsonPayload.metadata.justification")
	}
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}
