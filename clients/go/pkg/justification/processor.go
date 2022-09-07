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

// Package justification provides utils to validate justification tokens
// produced by JVS and populate audit logs with justification.
package justification

import (
	"context"
	"encoding/json"
	"fmt"

	zlogger "github.com/abcxyz/pkg/logging"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

const (
	TokenHeaderKey = "justification-token"
	LogMetadataKey = "justification"
)

// Validator validates justification token generated by JVS.
type Validator interface {
	// ValidateJWT validates justification token generated by JVS.
	ValidateJWT(jvsToken string) (*jwt.Token, error)
}

// Processor populates an audit log request with justification.
type Processor struct {
	validator Validator
}

// NewProcessor creates a new justification processor.
func NewProcessor(v Validator) *Processor {
	return &Processor{
		validator: v,
	}
}

// Process populates the given audit log request with the justification info from the given token.
// If the token is empty, this function does nothing.
func (p *Processor) Process(ctx context.Context, logReq *api.AuditLogRequest) error {
	logger := zlogger.FromContext(ctx)

	// TODO(#257): We will enforce the token in the future.
	jvsToken, ok := logReq.Context.GetFields()[TokenHeaderKey]
	if !ok || jvsToken.GetStringValue() == "" {
		logger.Info("no justification token found in AuditLogRequest")
		return nil
	}

	tok, err := p.validator.ValidateJWT(jvsToken.GetStringValue())
	if err != nil {
		return fmt.Errorf("failed to validate justification token: %w", err)
	}

	b, err := json.Marshal(*tok)
	if err != nil {
		return fmt.Errorf("failed to encode justification token: %w", err)
	}

	var tokStruct structpb.Struct
	if err := protojson.Unmarshal(b, &tokStruct); err != nil {
		return fmt.Errorf("failed to decode justification token: %w", err)
	}

	if logReq.Payload == nil {
		return fmt.Errorf("log request missing payload")
	}

	if logReq.Payload.Metadata == nil {
		logReq.Payload.Metadata = &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		}
	}

	logReq.Payload.Metadata.Fields[LogMetadataKey] = structpb.NewStructValue(&tokStruct)
	// TODO(#266): Also populate RequestAttribute.Reason.
	return nil
}
