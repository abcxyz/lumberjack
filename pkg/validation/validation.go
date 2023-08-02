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

// Package validation provides untils for lumberjack/data access logs
// validation.
package validation

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	lepb "cloud.google.com/go/logging/apiv2/loggingpb"
	cal "google.golang.org/genproto/googleapis/cloud/audit"
)

// Validate validates a json string representation of a lumberjack log.
func Validate(log string) error {
	var logEntry lepb.LogEntry
	if err := protojson.Unmarshal([]byte(log), &logEntry); err != nil {
		return fmt.Errorf("failed to parse log entry as JSON: %w", err)
	}

	if err := validatePayload(&logEntry); err != nil {
		return fmt.Errorf("failed to validate payload: %w", err)
	}

	// TODO (#427): add required fields check.
	return nil
}

func validatePayload(logEntry *lepb.LogEntry) error {
	payload := logEntry.GetJsonPayload()
	if payload == nil {
		return fmt.Errorf("missing audit log payload")
	}

	var al cal.AuditLog
	val, err := payload.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to extract audit log from JSON payload: %w", err)
	}
	if err := protojson.Unmarshal(val, &al); err != nil {
		return fmt.Errorf("failed to parse JSON payload: %w", err)
	}
	return nil
}
