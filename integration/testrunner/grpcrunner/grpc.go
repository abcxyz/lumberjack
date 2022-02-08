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

package grpcrunner

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
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

type GRPC struct {
	t            testing.TB
	ctx          context.Context
	endpointURL  string
	idToken      string
	projectID    string
	datasetQuery string
	cfg          *utils.Config
}

func NewGRPC(t testing.TB, ctx context.Context, endpointURL string, idToken string, projectID string, datasetQuery string, cfg *utils.Config) *GRPC {
	return &GRPC{
		t:            t,
		ctx:          ctx,
		endpointURL:  endpointURL,
		idToken:      idToken,
		projectID:    projectID,
		datasetQuery: datasetQuery,
		cfg:          cfg,
	}
}

func (g *GRPC) TestGRPCEndpoint() {
	conn := createConnection(g.t, g.endpointURL, g.idToken)
	defer conn.Close()
	talkerClient := talkerpb.NewTalkerClient(conn)

	bqClient, err := utils.MakeClient(g.ctx, g.projectID)
	if err != nil {
		g.t.Fatalf("BigQuery request failed: %v.", err)
	}

	defer func() {
		if err := bqClient.Close(); err != nil {
			g.t.Logf("Failed to close the BQ client: %v.", err)
		}
	}()

	g.runHelloCheck(talkerClient, bqClient)
	g.runFibonacciCheck(talkerClient, bqClient)
}

// End-to-end test for the fibonacci API, which is a test for server-side streaming.
func (g *GRPC) runFibonacciCheck(talkerClient talkerpb.TalkerClient, bqClient *bigquery.Client) {
	u := uuid.New()
	places := 5
	stream, err := talkerClient.Fibonacci(g.ctx, &talkerpb.FibonacciRequest{Places: uint32(places), Target: u.String()})
	if err != nil {
		g.t.Fatalf("fibonacci call failed: %v", err)
	}
	for {
		place, err := stream.Recv()
		if err == io.EOF {
			// stream is finished
			break
		}
		if err != nil {
			g.t.Fatalf("Err while reading fibonacci stream: %v", err)
		}
		g.t.Logf("Received value %s", place.Value)
	}
	query := makeQueryForGrpcStream(*bqClient, u, g.projectID, g.datasetQuery)
	utils.QueryIfAuditLogsExistWithRetries(g.t, g.ctx, query, g.cfg, int64(places))
}

// End-to-end test for the hello API, which is a test for unary requests.
func (g *GRPC) runHelloCheck(talkerClient talkerpb.TalkerClient, bqClient *bigquery.Client) {
	u := uuid.New()
	_, err := talkerClient.Hello(g.ctx, &talkerpb.HelloRequest{Message: "Some Message", Target: u.String()})
	if err != nil {
		g.t.Fatalf("could not greet: %v", err)
	}
	query := makeQueryForGrpcUnary(*bqClient, u, g.projectID, g.datasetQuery)
	utils.QueryIfAuditLogExistsWithRetries(g.t, g.ctx, query, g.cfg)
}

// Server is in cloud run. Example: https://cloud.google.com/run/docs/triggering/grpc#request-auth
// We are using token-based authentication to connect to the server, which will be passed through
// a JWT to the server. There, the server will be able to decipher the JWT to find the calling user.
func createConnection(t testing.TB, addr string, idToken string) *grpc.ClientConn {
	rpcCreds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: idToken})

	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("failed to load system cert pool: %w", err)
	}
	creds := credentials.NewClientTLSFromCert(pool, "")

	// The address is input in a way that causes an error within grpc.
	// example input: https://my-example-server.a.run.app
	// expected url format: my-example-server.a.run.app:443
	addr = strings.TrimPrefix(addr, "https://")
	addr = addr + ":443"
	conn, err := grpc.Dial(addr, grpc.WithPerRPCCredentials(rpcCreds), grpc.WithTransportCredentials(creds))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	return conn
}

// This query is used to find the relevant audit log in BigQuery, which we assume will be added by the server.
// We specifically look up the log using the UUID specified in the request as we know the server will add that
// as the resource name, and provides us a unique key to find logs with.
func makeQueryForGrpcUnary(client bigquery.Client, u uuid.UUID, projectID string, datasetQuery string) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s.%s WHERE jsonPayload.resource_name=? LIMIT 1", projectID, datasetQuery)
	return utils.MakeQuery(client, u, queryString)
}

// Similar to the above function, but can return multiple results, which is what we expect for streaming.
func makeQueryForGrpcStream(client bigquery.Client, u uuid.UUID, projectID string, datasetQuery string) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s.%s WHERE jsonPayload.resource_name=?", projectID, datasetQuery)
	return utils.MakeQuery(client, u, queryString)
}
