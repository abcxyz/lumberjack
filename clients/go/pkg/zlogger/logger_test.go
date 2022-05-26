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

package zlogger

import (
	"context"
	"testing"
)

func TestNewFromEnv(t *testing.T) {
	t.Parallel()

	if logger := NewFromEnv(""); logger == nil {
		t.Fatal("expected logger to never be nil")
	}
}

func TestDefault(t *testing.T) {
	t.Parallel()

	logger1 := Default()
	if logger1 == nil {
		t.Fatal("expected logger to never be nil")
	}

	logger2 := Default()
	if logger2 == nil {
		t.Fatal("expected logger to never be nil")
	}

	// Intentionally comparing identities here
	if logger1 != logger2 {
		t.Errorf("expected %#v to be %#v", logger1, logger2)
	}
}

func TestContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger1 := FromContext(ctx)
	if logger1 == nil {
		t.Fatal("expected logger to never be nil")
	}

	ctx = WithLogger(ctx, logger1)

	logger2 := FromContext(ctx)
	if logger1 != logger2 {
		t.Errorf("expected %#v to be %#v", logger1, logger2)
	}
	_ = logger2
}
