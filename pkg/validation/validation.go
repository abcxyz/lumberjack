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
	"errors"
	"fmt"

	"github.com/abcxyz/lumberjack/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"

	lepb "cloud.google.com/go/logging/apiv2/loggingpb"
	cal "google.golang.org/genproto/googleapis/cloud/audit"
)

var requiredLabels = map[string]struct{}{
	"environment":            {},
	"accessing_process_name": {},
}

// Validate validates a json string representation of a lumberjack log.
func Validate(log string) error {
	var logEntry lepb.LogEntry
	if err := protojson.Unmarshal([]byte(log), &logEntry); err != nil {
		return fmt.Errorf("failed to parse log entry as JSON: %w", err)
	}

	var retErr error
	if err := validatePayload(&logEntry); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("failed to validate payload: %w", err))
	}

	if err := checkRequiredLabels(logEntry.Labels); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("missing required labels: %w", err))
	}
	return retErr
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
	if err := util.Validate(&al); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	return nil
}

func checkRequiredLabels(ls map[string]string) error {
	if ls == nil {
		return fmt.Errorf("labels is empty")
	}

	var retErr error
	for k := range requiredLabels {
		if _, ok := ls[k]; !ok {
			retErr = errors.Join(retErr, fmt.Errorf("missing label: %q", k))
		}
	}
	return retErr
}
