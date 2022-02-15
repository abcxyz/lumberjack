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

type LabelProcessor struct {
	DefaultLabels map[string]string
}

func (p *LabelProcessor) Process(ctx context.Context, logReq *alpb.AuditLogRequest) error {
	if p.DefaultLabels == nil || len(p.DefaultLabels) == 0 {
		// shortcut if there are no labels to add
		return nil
	}

	for key, val := range p.DefaultLabels {
		if _, exists := logReq.Labels[key]; !exists {
			logReq.Labels[key] = val
		}
	}
	return nil
}
