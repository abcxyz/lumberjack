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
	"github.com/sethvargo/go-retry"
	"google.golang.org/api/iterator"

	logging "cloud.google.com/go/logging/apiv2"
)

// Puller pulls log entries of GCP organizations, folders, projects, or
// billingAccounts.
type Puller struct {
	client *logging.Client
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

// NewPuller creates a new Puller with provided clients and options.
func NewPuller(ctx context.Context, c *logging.Client, opts ...Option) *Puller {
	p := &Puller{client: c}
	for _, opt := range opts {
		p = opt(p)
	}

	if p.retry == nil {
		p.retry = retry.WithMaxRetries(5, retry.NewFibonacci(500*time.Millisecond))
	}
	return p
}

// Pull pulls a list of log entries given request.
func (p *Puller) Pull(ctx context.Context, req *loggingpb.ListLogEntriesRequest) ([]*loggingpb.LogEntry, error) {
	var ls []*loggingpb.LogEntry
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
			ls = append(ls, l)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list log entries: %w", err)
	}
	return ls, nil
}
