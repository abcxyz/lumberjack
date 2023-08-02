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
	"fmt"
	"net/mail"
	"testing"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/abcxyz/pkg/bqutil"
	"github.com/abcxyz/pkg/logging"
	"github.com/sethvargo/go-retry"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/encoding/protojson"
)

type bqResult struct {
	// We expect a single JSON column named "result" from all queries.
	Result string
}

func makeQuery(bqClient *bigquery.Client, id, queryString string) *bigquery.Query {
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
func validateAuditLogsWithRetries(ctx context.Context, tb testing.TB, tcfg *TestCaseConfig, bqQuery *bigquery.Query, wantNum int) {
	tb.Helper()
	ctx = logging.WithLogger(ctx, logging.TestLogger(tb, zaptest.Level(zapcore.DebugLevel)))

	backoff := retry.WithMaxRetries(tcfg.MaxDBQueryTries, retry.NewConstant(tcfg.LogRoutingWait))
	q := bqutil.NewQuery[*bqResult](bqQuery)

	results, err := bqutil.RetryQueryEntries(ctx, q, wantNum, backoff)
	if err != nil {
		tb.Fatalf("query BigQuery failed: %v", err)
	}

	for i, r := range results {
		var logEntry loggingpb.LogEntry
		if err := protojson.Unmarshal([]byte(r.Result), &logEntry); err != nil {
			tb.Fatalf("unmarshal BigQuery row[%d] to LogEntry failed: %v", i, err)
		}

		tb.Logf("diffing LogEntry from row[%d]", i)
		diffLogEntry(tb, &logEntry)
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
func diffLogEntry(tb testing.TB, logEntry *loggingpb.LogEntry) {
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

	checkJustification(tb, jsonPayloadInfo)
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
	if string(b) == "null" || string(b) == "" {
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
