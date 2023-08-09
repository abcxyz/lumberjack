// Copyright 2023 The Authors (see AUTHORS file)
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

// Package cloudlogging pulls lumberjack/data access logs from GCP cloud
// logging.
package cloudlogging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/googleapis/gax-go"
	"github.com/sethvargo/go-retry"
	"google.golang.org/api/iterator"

	logging "cloud.google.com/go/logging/apiv2"
)

// LoggingClient interface for streaming read of log entries as they are ingested.
type LoggingClient interface {
	ListLogEntries(context.Context, *loggingpb.ListLogEntriesRequest, ...gax.CallOption) *logging.LogEntryIterator
}

// Puller pulls log entries of GCP organizations, folders, projects, and
// billingAccounts.
type Puller struct {
	client LoggingClient
	// Required. Name of a parent resource from which to retrieve log entries:
	//
	// *  `projects/[PROJECT_ID]`
	// *  `organizations/[ORGANIZATION_ID]`
	// *  `billingAccounts/[BILLING_ACCOUNT_ID]`
	// *  `folders/[FOLDER_ID]`
	//
	// May alternatively be one or more views:
	//
	//   - `projects/[PROJECT_ID]/locations/[LOCATION_ID]/buckets/[BUCKET_ID]/views/[VIEW_ID]`
	//   - `organizations/[ORGANIZATION_ID]/locations/[LOCATION_ID]/buckets/[BUCKET_ID]/views/[VIEW_ID]`
	//   - `billingAccounts/[BILLING_ACCOUNT_ID]/locations/[LOCATION_ID]/buckets/[BUCKET_ID]/views/[VIEW_ID]`
	//   - `folders/[FOLDER_ID]/locations/[LOCATION_ID]/buckets/[BUCKET_ID]/views/[VIEW_ID]`
	resource string
	// Optional retry backoff strategy, default is 5 attempts with fibonacci
	// backoff that starts at 500ms.
	retry retry.Backoff
}

// Option is the option to set up an log puller.
type Option func(h *Puller) *Puller

// WithRetry provides retry strategy to the log puller.
func WithRetry(b retry.Backoff) Option {
	return func(p *Puller) *Puller {
		p.retry = b
		return p
	}
}

// NewPuller creates a new Puller with provided clients, resource, and options.
func NewPuller(ctx context.Context, c LoggingClient, resource string, opts ...Option) *Puller {
	p := &Puller{client: c, resource: resource}
	for _, opt := range opts {
		p = opt(p)
	}

	if p.retry == nil {
		p.retry = retry.WithMaxRetries(5, retry.NewFibonacci(500*time.Millisecond))
	}
	return p
}

// Pull pulls log entries and send them to the result channel.
func (p *Puller) Pull(ctx context.Context, filter string, result chan<- *loggingpb.LogEntry) error {
	req := &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{p.resource},
		Filter:        filter,
		// Set descending time order so that newest logs will be returned.
		OrderBy: "timestamp desc",
		// Set a large pagesize to optimize read efficiency, see ref
		// https://cloud.google.com/logging/docs/reference/api-overview#entries-list.
		PageSize: 1000,
	}
	if err := retry.Do(ctx, p.retry, func(ctx context.Context) error {
		it := p.client.ListLogEntries(ctx, req)
		for {
			l, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return retry.RetryableError(fmt.Errorf("failed to get next log entry: %w, retrying", err))
			}
			result <- l
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to list log entries: %w", err)
	}
	return nil
}
