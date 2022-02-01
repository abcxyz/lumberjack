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

package grpcrunner

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"strings"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/abcxyz/lumberjack/integration/testrunner/talkerpb"
	"github.com/abcxyz/lumberjack/integration/testrunner/utils"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func TestGRPCEndpoint(t testing.TB, ctx context.Context, endpointURL string,
	idToken string, projectID string, datasetQuery string, cfg *utils.Config) {
	conn := createConnection(endpointURL, idToken)
	defer conn.Close()
	c := talkerpb.NewTalkerClient(conn)
	u := uuid.New()
	_, err := c.Hello(ctx, &talkerpb.HelloRequest{Message: "Some Message", Target: u.String()})
	if err != nil {
		t.Fatalf("could not greet: %v", err)
	}

	bqClient, err := utils.MakeClient(ctx, projectID)
	if err != nil {
		t.Fatalf("BigQuery request failed: %v.", err)
	}

	defer func() {
		if err := bqClient.Close(); err != nil {
			t.Logf("Failed to close the BQ client: %v.", err)
		}
	}()

	query := makeQueryForGrpc(*bqClient, u, projectID, datasetQuery)
	utils.QueryIfAuditLogExistsWithRetries(t, ctx, query, cfg)
}

func createConnection(addr string, idToken string) *grpc.ClientConn {
	rpcCreds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: idToken})

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("failed to load system cert pool: %w", err)
	}
	creds := credentials.NewClientTLSFromCert(pool, "")

	addr = strings.TrimPrefix(addr, "https://")
	addr = addr + ":443"
	conn, err := grpc.Dial(addr, grpc.WithPerRPCCredentials(rpcCreds), grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return conn
}

func makeQueryForGrpc(client bigquery.Client, u uuid.UUID, projectID string, datasetQuery string) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s.%s WHERE jsonPayload.resource_name=? LIMIT 1", projectID, datasetQuery)
	return utils.MakeQuery(client, u, queryString)
}
