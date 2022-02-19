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

// Package server implements the gRPC server of the audit log agent.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/compute/metadata"
	dlp "cloud.google.com/go/dlp/apiv2"
	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/zlogger"
	dlp2 "google.golang.org/genproto/googleapis/privacy/dlp/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuditLogAgent is the implementation of the audit log agent server.
type AuditLogAgent struct {
	alpb.UnimplementedAuditLogAgentServer
	dlpClient *dlp.Client

	client *audit.Client
}

// NewAuditLogAgent creates a new AuditLogAgent.
func NewAuditLogAgent(client *audit.Client, dlpClient *dlp.Client) (*AuditLogAgent, error) {
	return &AuditLogAgent{client: client, dlpClient: dlpClient}, nil
}

// ProcessLog processes the log requests by calling the internal client.
func (a *AuditLogAgent) ProcessLog(ctx context.Context, logReq *alpb.AuditLogRequest) (*alpb.AuditLogResponse, error) {
	if err := a.redactUsingDLP(ctx, logReq); err != nil {
		return nil, codifyErr(err)
	}

	if err := a.client.Log(ctx, logReq); err != nil {
		return nil, codifyErr(err)
	}

	return &alpb.AuditLogResponse{
		Result: logReq,
	}, nil
}

func codifyErr(err error) error {
	if errors.Is(err, audit.ErrInvalidRequest) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	// TODO: Handle other well-known errors if we have more.
	return status.Error(codes.Internal, err.Error())
}

func (a *AuditLogAgent) redactUsingDLP(ctx context.Context, logReq *alpb.AuditLogRequest) error {
	zlogger := zlogger.FromContext(ctx)
	auditLog := *logReq.GetPayload()
	request := auditLog.GetRequest()
	if request == nil {
		zlogger.Debug("Request was nil, not calling DLP.")
		return nil
	}

	projectID, err := metadata.ProjectID()
	if err != nil {
		return fmt.Errorf("failed to get project ID from metadata server: %w", err)
	}

	stringRequest, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("err when converting to json: %w", err)
	}

	req := &dlp2.DeidentifyContentRequest{
		Parent: fmt.Sprintf("projects/%s", projectID),
		Item:   &dlp2.ContentItem{DataItem: &dlp2.ContentItem_Value{Value: string(stringRequest)}},
		DeidentifyConfig: &dlp2.DeidentifyConfig{
			Transformation: &dlp2.DeidentifyConfig_InfoTypeTransformations{
				InfoTypeTransformations: &dlp2.InfoTypeTransformations{
					Transformations: []*dlp2.InfoTypeTransformations_InfoTypeTransformation{
						{
							InfoTypes: []*dlp2.InfoType{}, // Match all info types.
							PrimitiveTransformation: &dlp2.PrimitiveTransformation{
								Transformation: &dlp2.PrimitiveTransformation_ReplaceWithInfoTypeConfig{},
							},
						},
					},
				},
			},
		},
		// The InspectConfig is used to identify the DATE fields.
		InspectConfig: &dlp2.InspectConfig{
			InfoTypes: []*dlp2.InfoType{
				{
					Name: "DATE",
				},
				{
					Name: "EMAIL_ADDRESS",
				},
			},
		},
	}
	resp, err := a.dlpClient.DeidentifyContent(ctx, req)
	if err != nil {
		return fmt.Errorf("Err when calling dlp: %w", err)
	}
	// TODO: Use resp.
	_ = resp
	deIdentifiedRequest := resp.GetItem().GetDataItem()
	trimmed := trimFirstRune(fmt.Sprintf("%v", deIdentifiedRequest))

	zlogger.Warnf("Response before dlp: %s", stringRequest)
	zlogger.Warnf("Response after dlp: %s", trimmed)

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &m); err != nil {
		zlogger.Warnf("Unmarshal failed: %v", err)
	}

	if logReq.Payload.Request, err = audit.ToProtoStruct(m); err != nil {
		zlogger.Warnf("Err converting: %v", err)
		return fmt.Errorf("err when converting %v to struct: %w", request.String(), err)
	}

	return nil
}

func trimFirstRune(s string) string {
	return s[2 : len(s)-1]
}
