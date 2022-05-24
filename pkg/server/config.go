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

package server

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

// Config includes all server configs.
type Config struct {
	Port string `env:"PORT, default=8080"`

	// If trace ratio is >= 1 will trace all requests.
	// If trace ratio <= 0 will not trace at all.
	TraceRatio float64 `env:"TRACE_RATIO, default=0.001"`
}

// NewConfig initializes a server config from environment vars.
func NewConfig(ctx context.Context) (*Config, error) {
	var c Config
	if err := envconfig.ProcessWith(ctx, &c, envconfig.OsLookuper()); err != nil {
		return nil, err
	}
	return &c, nil
}
