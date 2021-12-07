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
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func makeAuditLogRequest(u uuid.UUID, endpointUrl string, requestTimeout time.Duration, authToken string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, endpointUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Insert the UUID used in tracing the log as a query parameter.
	q := req.URL.Query()
	q.Add("trace_id", u.String())
	req.URL.RawQuery = q.Encode()

	httpClient := &http.Client{Timeout: requestTimeout}
	return httpClient.Do(req)
}
