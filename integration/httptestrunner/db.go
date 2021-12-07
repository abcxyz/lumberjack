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

package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
)

// queryIfAuditLogExists queries the DB and checks if audit log contained in the query exists or not.
func queryIfAuditLogExists(ctx context.Context, query *bigquery.Query) (bool, error) {
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

// makeClientAndQuery prepares the BQ client and query for verifying that the audit log made it to the BQ DB.
func makeClientAndQuery(ctx context.Context, u uuid.UUID, projectID string, datasetQuery string) (*bigquery.Client, *bigquery.Query, error) {
	bqClient, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}

	bqQuery := bqClient.Query(fmt.Sprintf("SELECT count(*) FROM (SELECT * FROM %s.%s WHERE labels.trace_id=? LIMIT 1)", projectID, datasetQuery))
	bqQuery.Parameters = []bigquery.QueryParameter{{Value: u.String()}}
	return bqClient, bqQuery, nil
}
