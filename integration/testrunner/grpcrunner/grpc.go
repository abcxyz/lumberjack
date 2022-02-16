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
	ProjectID    string
	DatasetQuery string

	// IDToken and endpoint URL are required unless a custom TalkerClient is provided.
	IDToken      string
	EndpointURL  string
	TalkerClient talkerpb.TalkerClient

	Config         *utils.Config
	BigQueryClient *bigquery.Client
}

// TestGRPCEndpoint runs tests against a GRPC endpoint. The given GRPC must
// define a projectID and datasetQuery. If a TalkerClient or BigQueryClient are
// not provided, they are instantiated via the defaults.
func TestGRPCEndpoint(t testing.TB, ctx context.Context, g *GRPC) {
	if g.ProjectID == "" {
		t.Fatal("ProjectID must be set")
	}

	if g.DatasetQuery == "" {
		t.Fatal("DatasetQuery must be set")
	}

	if g.TalkerClient == nil {
		if g.IDToken == "" || g.EndpointURL == "" {
			t.Fatal("IDToken and EndpointURL are required to create a TalkerClient")
		}

		conn := createConnection(t, g.EndpointURL, g.IDToken)
		t.Cleanup(func() {
			conn.Close()
		})
		g.TalkerClient = talkerpb.NewTalkerClient(conn)
	}

	if g.BigQueryClient == nil {
		bqClient, err := utils.MakeClient(ctx, g.ProjectID)
		if err != nil {
			t.Fatalf("BigQuery request failed: %v", err)
		}
		t.Cleanup(func() {
			if err := bqClient.Close(); err != nil {
				t.Errorf("Failed to close the BQ client: %v", err)
			}
		})
		g.BigQueryClient = bqClient
	}

	g.runHelloCheck(t, ctx)

	// TODO(#149): Reenable stream interception tests once Go adds stream handling.
	if strings.Contains(g.EndpointURL, "java") {
		g.runFailCheck(t, ctx)
		g.runFibonacciCheck(t, ctx)
		g.runAdditionCheck(t, ctx)
		g.runFailOnFourCheck(t, ctx)
	}
}

// End-to-end test for the fibonacci API, which is a test for server-side streaming.
func (g *GRPC) runFibonacciCheck(t testing.TB, ctx context.Context) {
	u := uuid.New()
	places := 5
	stream, err := g.TalkerClient.Fibonacci(ctx, &talkerpb.FibonacciRequest{Places: uint32(places), Target: u.String()})
	if err != nil {
		t.Fatalf("fibonacci call failed: %v", err)
	}
	for {
		place, err := stream.Recv()
		if err == io.EOF {
			// stream is finished
			break
		}
		if err != nil {
			t.Fatalf("Err while reading fibonacci stream: %v", err)
		}
		t.Logf("Received value %v", place.Value)
	}
	query := g.makeQueryForGRPCStream(u)
	utils.QueryIfAuditLogsExistWithRetries(t, ctx, query, g.Config, "fibonacciCheck", int64(places))
}

// End-to-end test for the addition API, which is a test for client-side streaming.
func (g *GRPC) runAdditionCheck(t testing.TB, ctx context.Context) {
	u := uuid.New()
	stream, err := g.TalkerClient.Addition(ctx)
	if err != nil {
		t.Fatalf("addition call failed: %v", err)
	}
	totalNumbers := 5
	for i := 0; i < totalNumbers; i++ {
		if err := stream.Send(&talkerpb.AdditionRequest{
			Addend: uint32(i),
			Target: u.String(),
		}); err != nil {
			t.Fatalf("sending value to addition failed: %v", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("failed getting result from addition: %v", err)
	}
	t.Logf("Value returned: %d", reply.Sum)

	query := g.makeQueryForGRPCStream(u)
	utils.QueryIfAuditLogsExistWithRetries(t, ctx, query, g.Config, "additionCheck", int64(totalNumbers))
}

// End-to-end test for the hello API, which is a test for unary requests.
func (g *GRPC) runHelloCheck(t testing.TB, ctx context.Context) {
	u := uuid.New()
	_, err := g.TalkerClient.Hello(ctx, &talkerpb.HelloRequest{Message: "Some Message", Target: u.String()})
	if err != nil {
		t.Fatalf("could not greet: %v", err)
	}
	query := g.makeQueryForGRPCUnary(u)
	utils.QueryIfAuditLogExistsWithRetries(t, ctx, query, g.Config, "helloCheck")
}

// End-to-end test for the fail API, which is a test for unary failures.
func (g *GRPC) runFailCheck(t testing.TB, ctx context.Context) {
	u := uuid.New()
	_, err := g.TalkerClient.Fail(ctx, &talkerpb.FailRequest{Message: "Some Message", Target: u.String()})
	if err == nil {
		t.Fatalf("expected err but did not get one: %v", err)
	}
	query := g.makeQueryForGRPCUnary(u)
	t.Logf("querying with: %v", query)
	utils.QueryIfAuditLogExistsWithRetries(t, ctx, query, g.Config, "failCheck")
}

// End-to-end test for the failOnFour API, which is a test for failures during client-side streaming.
func (g *GRPC) runFailOnFourCheck(t testing.TB, ctx context.Context) {
	u := uuid.New()
	stream, err := g.TalkerClient.FailOnFour(ctx)
	if err != nil {
		t.Fatalf("addition call failed: %v", err)
	}
	totalNumbers := 5
	for i := 1; i <= totalNumbers; i++ {
		if err := stream.Send(&talkerpb.FailOnFourRequest{
			Value:  uint32(i),
			Target: u.String(),
		}); err != nil {
			t.Fatalf("sending value to addition failed: %v", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		t.Logf("Got Error as expectd: %v", err)
	} else {
		t.Fatalf("Did not get err as expected: %v", reply)
	}

	query := g.makeQueryForGRPCStream(u)
	// we expect to have 4 audit logs - the last sent number (5) will be after the err ocurred.
	utils.QueryIfAuditLogsExistWithRetries(t, ctx, query, g.Config, "failOnFourCheck", int64(4))
}

// Server is in cloud run. Example: https://cloud.google.com/run/docs/triggering/grpc#request-auth
// We are using token-based authentication to connect to the server, which will be passed through
// a JWT to the server. There, the server will be able to decipher the JWT to find the calling user.
func createConnection(t testing.TB, addr string, idToken string) *grpc.ClientConn {
	rpcCreds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: idToken})

	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("failed to load system cert pool: %v", err)
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
func (g *GRPC) makeQueryForGRPCUnary(u uuid.UUID) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s.%s WHERE jsonPayload.resource_name=? LIMIT 1", g.ProjectID, g.DatasetQuery)
	return utils.MakeQuery(*g.BigQueryClient, u, queryString)
}

// Similar to the above function, but can return multiple results, which is what we expect for streaming.
func (g *GRPC) makeQueryForGRPCStream(u uuid.UUID) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s.%s WHERE jsonPayload.resource_name=?", g.ProjectID, g.DatasetQuery)
	return utils.MakeQuery(*g.BigQueryClient, u, queryString)
}
