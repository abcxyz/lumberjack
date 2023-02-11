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
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	AuditLogRequestTimeout  time.Duration `env:"AUDIT_CLIENT_TEST_AUDIT_LOG_REQUEST_TIMEOUT,default=30s"`
	AuditLogRequestWait     time.Duration `env:"AUDIT_CLIENT_TEST_AUDIT_LOG_REQUEST_WAIT,default=4s"`
	HTTPEndpoints           string        `env:"HTTP_ENDPOINTS,required"`
	GRPCEndpoints           string        `env:"GRPC_ENDPOINTS,required"`
	LogRoutingWait          time.Duration `env:"AUDIT_CLIENT_TEST_AUDIT_LOG_ROUTING_WAIT,default=25s"`
	MaxAuditLogRequestTries uint64        `env:"AUDIT_CLIENT_TEST_MAX_AUDIT_LOG_REQUEST_TRIES,default=4"`
	MaxDBQueryTries         uint64        `env:"AUDIT_CLIENT_TEST_MAX_DB_QUERY_TRIES,default=5"`
}

func newTestConfig(ctx context.Context) (*Config, error) {
	var c Config
	if err := envconfig.ProcessWith(ctx, &c, envconfig.OsLookuper()); err != nil {
		return nil, fmt.Errorf("failed to process environment: %w", err)
	}
	return &c, nil
}
