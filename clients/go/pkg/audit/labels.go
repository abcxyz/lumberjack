// Copyright 2022 Lumberjack authors (see AUTHORS file)
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

package audit

import (
	"context"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

// LabelProcessor is a mutator that adds labels to each AuditLogRequest. These labels
// are specified through the configuration, and are intended to be defaults. They do
// not overwrite any labels that are already in the request, and can be overwritten by
// the server code.
type LabelProcessor struct {
	DefaultLabels map[string]string
}

// Process adds the configured labels to each passed in request, without overwriting
// existing labels.
func (p *LabelProcessor) Process(ctx context.Context, logReq *alpb.AuditLogRequest) error {
	if len(p.DefaultLabels) == 0 {
		// short circuit if there are no labels to add
		return nil
	}
	if logReq.Labels == nil {
		logReq.Labels = map[string]string{}
	}

	for key, val := range p.DefaultLabels {
		if _, exists := logReq.Labels[key]; !exists {
			logReq.Labels[key] = val
		}
	}
	return nil
}
