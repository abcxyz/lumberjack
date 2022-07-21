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
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	jvspb "github.com/abcxyz/jvs/apis/v0"
	"github.com/abcxyz/jvs/pkg/jvscrypto"
	"github.com/abcxyz/lumberjack/integration/testrunner/talkerpb"
	"github.com/abcxyz/lumberjack/integration/testrunner/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	rpccode "google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPC struct {
	ProjectID    string
	DatasetQuery string

	// IDToken and endpoint URL are required unless a custom TalkerClient is provided.
	IDToken      string
	EndpointURL  string
	TalkerClient talkerpb.TalkerClient

	Config               *utils.Config
	BigQueryClient       *bigquery.Client
	RequireJustification bool
}

// Matching public key here: https://github.com/abcxyz/lumberjack/pull/261/files#diff-f06321655121106c1e25ed2dd8cf773af6d7d5fb9b129abb2e1c04ba4d6dea5eR48
const privateKey = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIITZ4357UsTCbhxXu8w8cY54ZLlsAIJj/Aej9ylb/ZfBoAoGCCqGSM49\nAwEHoUQDQgAEhBWj8vw5LkPRWbCr45k0cOarIcWgApM03mSYF911de5q1wGOL7R9\nN8pC7jo2xbS+i1wGsMiz+AWnhhZIQcNTKg==\n-----END EC PRIVATE KEY-----"

// TestGRPCEndpoint runs tests against a GRPC endpoint. The given GRPC must
// define a projectID and datasetQuery. If a TalkerClient or BigQueryClient are
// not provided, they are instantiated via the defaults.
func TestGRPCEndpoint(ctx context.Context, tb testing.TB, g *GRPC) {
	tb.Helper()

	if g.ProjectID == "" {
		tb.Fatal("ProjectID must be set")
	}

	if g.DatasetQuery == "" {
		tb.Fatal("DatasetQuery must be set")
	}

	if g.TalkerClient == nil {
		if g.IDToken == "" || g.EndpointURL == "" {
			tb.Fatal("IDToken and EndpointURL are required to create a TalkerClient")
		}

		conn := createConnection(tb, g.EndpointURL, g.IDToken)
		tb.Cleanup(func() {
			conn.Close()
		})
		g.TalkerClient = talkerpb.NewTalkerClient(conn)
	}

	signedToken, err := justificationToken()
	if err != nil {
		tb.Fatalf("couldn't generate justification token: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	md.Set("justification_token", signedToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	if g.BigQueryClient == nil {
		bqClient, err := utils.MakeClient(ctx, g.ProjectID)
		if err != nil {
			tb.Errorf("BigQuery request failed: %v", err)
		}
		tb.Cleanup(func() {
			if err := bqClient.Close(); err != nil {
				tb.Errorf("Failed to close the BQ client: %v", err)
			}
		})
		g.BigQueryClient = bqClient
	}

	justificationRequired := strings.Contains(g.EndpointURL, "go")

	g.runHelloCheck(ctx, tb, justificationRequired)
	g.runFailCheck(ctx, tb, justificationRequired)
	g.runFibonacciCheck(ctx, tb, justificationRequired)
	g.runAdditionCheck(ctx, tb, justificationRequired)
	g.runFailOnFourCheck(ctx, tb, justificationRequired)
}

// create a justification token to pass in the call to services.
func justificationToken() (string, error) {
	now := time.Now().UTC()
	claims := jvspb.JVSClaims{
		StandardClaims: &jwt.StandardClaims{
			Audience:  "talker-app",
			ExpiresAt: now.Add(time.Hour).Unix(),
			Id:        uuid.New().String(),
			IssuedAt:  now.Unix(),
			Issuer:    "lumberjack-test-runner",
			NotBefore: now.Unix(),
			Subject:   "lumberjack-integ",
		},
		Justifications: []*jvspb.Justification{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = "integ-key"

	block, _ := pem.Decode([]byte(privateKey))
	key, _ := x509.ParseECPrivateKey(block.Bytes)
	return jvscrypto.SignToken(token, key)
}

// End-to-end test for the fibonacci API, which is a test for server-side streaming.
func (g *GRPC) runFibonacciCheck(ctx context.Context, tb testing.TB, justificationRequired bool) {
	tb.Helper()

	u := uuid.New()
	places := 5
	stream, err := g.TalkerClient.Fibonacci(ctx, &talkerpb.FibonacciRequest{Places: uint32(places), Target: u.String()})
	if err != nil {
		tb.Errorf("fibonacci call failed: %v", err)
	}
	for {
		place, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// stream is finished
			break
		}
		if err != nil {
			tb.Errorf("Err while reading fibonacci stream: %v", err)
		}
		tb.Logf("Received value %v", place.Value)
	}
	query := g.makeQueryForGRPCStream(u, justificationRequired)
	utils.QueryIfAuditLogsExistWithRetries(ctx, tb, query, g.Config, "server_stream_fibonacci", int64(places))
}

// End-to-end test for the addition API, which is a test for client-side streaming.
func (g *GRPC) runAdditionCheck(ctx context.Context, tb testing.TB, justificationRequired bool) {
	tb.Helper()

	u := uuid.New()
	stream, err := g.TalkerClient.Addition(ctx)
	if err != nil {
		tb.Errorf("addition call failed: %v", err)
	}
	totalNumbers := 5
	for i := 0; i < totalNumbers; i++ {
		if err := stream.Send(&talkerpb.AdditionRequest{
			Addend: uint32(i),
			Target: u.String(),
		}); err != nil {
			tb.Errorf("sending value to addition failed: %v", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		tb.Errorf("failed getting result from addition: %v", err)
	}
	tb.Logf("Value returned: %d", reply.Sum)

	query := g.makeQueryForGRPCStream(u, justificationRequired)
	utils.QueryIfAuditLogsExistWithRetries(ctx, tb, query, g.Config, "client_stream_addition", int64(totalNumbers))
}

// End-to-end test for the hello API, which is a test for unary requests.
func (g *GRPC) runHelloCheck(ctx context.Context, tb testing.TB, justificationRequired bool) {
	tb.Helper()

	u := uuid.New()
	_, err := g.TalkerClient.Hello(ctx, &talkerpb.HelloRequest{Message: "Some Message", Target: u.String()})
	if err != nil {
		tb.Errorf("could not greet: %v", err)
	}
	query := g.makeQueryForGRPCUnary(u, justificationRequired)
	utils.QueryIfAuditLogExistsWithRetries(ctx, tb, query, g.Config, "unary_hello")
}

// End-to-end test for the fail API, which is a test for unary failures.
func (g *GRPC) runFailCheck(ctx context.Context, tb testing.TB, justificationRequired bool) {
	tb.Helper()

	u := uuid.New()
	reply, err := g.TalkerClient.Fail(ctx, &talkerpb.FailRequest{Message: "Some Message", Target: u.String()})

	if err != nil {
		returnStatus, ok := status.FromError(err)
		if !ok {
			tb.Errorf("Could not convert err to status %v", err)
		}
		if int32(returnStatus.Code()) != int32(rpccode.Code_RESOURCE_EXHAUSTED) {
			tb.Errorf("Got unexpected Err. Got code %d but expected %d", int32(returnStatus.Code()),
				int32(rpccode.Code_RESOURCE_EXHAUSTED))
		}
		tb.Logf("Got Error as expected: %v", err)
	} else {
		tb.Errorf("Did not get err as expected. Instead got reply: %v", reply)
	}

	query := g.makeQueryForGRPCUnary(u, justificationRequired)
	utils.QueryIfAuditLogExistsWithRetries(ctx, tb, query, g.Config, "unary_fail")
}

// End-to-end test for the failOnFour API, which is a test for failures during client-side streaming.
func (g *GRPC) runFailOnFourCheck(ctx context.Context, tb testing.TB, justificationRequired bool) {
	tb.Helper()

	u := uuid.New()
	stream, err := g.TalkerClient.FailOnFour(ctx)
	if err != nil {
		tb.Errorf("addition call failed: %v", err)
	}
	totalNumbers := 5
	for i := 1; i <= totalNumbers; i++ {
		if err := stream.Send(&talkerpb.FailOnFourRequest{
			Value:  uint32(i),
			Target: u.String(),
		}); err != nil {
			tb.Errorf("sending value to addition failed: %v", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		returnStatus, ok := status.FromError(err)
		if !ok {
			tb.Errorf("Could not convert err to status %v", err)
		}
		if int32(returnStatus.Code()) != int32(rpccode.Code_INVALID_ARGUMENT) {
			tb.Errorf("Got unexpected Err. Got code %d but expected %d", int32(returnStatus.Code()),
				int32(rpccode.Code_INVALID_ARGUMENT))
		}
		tb.Logf("Got Error as expected: %v", err)
	} else {
		tb.Errorf("Did not get err as expected. Instead got reply: %v", reply)
	}

	query := g.makeQueryForGRPCStream(u, justificationRequired)
	// we expect to have 4 audit logs - the last sent number (5) will be after the err ocurred.
	utils.QueryIfAuditLogsExistWithRetries(ctx, tb, query, g.Config, "stream_fail_on_four", int64(4))
}

// Server is in cloud run. Example: https://cloud.google.com/run/docs/triggering/grpc#request-auth
// We are using token-based authentication to connect to the server, which will be passed through
// a JWT to the server. There, the server will be able to decipher the JWT to find the calling user.
func createConnection(tb testing.TB, addr string, idToken string) *grpc.ClientConn {
	tb.Helper()

	rpcCreds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: idToken})

	pool, err := x509.SystemCertPool()
	if err != nil {
		tb.Fatalf("failed to load system cert pool: %v", err)
	}
	creds := credentials.NewClientTLSFromCert(pool, "")

	// The address is input in a way that causes an error within grpc.
	// example input: https://my-example-server.a.run.app
	// expected url format: my-example-server.a.run.app:443
	addr = strings.TrimPrefix(addr, "https://")
	addr = addr + ":443"
	conn, err := grpc.Dial(addr, grpc.WithPerRPCCredentials(rpcCreds), grpc.WithTransportCredentials(creds))
	if err != nil {
		tb.Fatalf("did not connect: %v", err)
	}
	return conn
}

// This query is used to find the relevant audit log in BigQuery, which we assume will be added by the server.
// We specifically look up the log using the UUID specified in the request as we know the server will add that
// as the resource name, and provides us a unique key to find logs with.
func (g *GRPC) makeQueryForGRPCUnary(u uuid.UUID) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s.%s WHERE jsonPayload.resource_name=?", g.ProjectID, g.DatasetQuery)
	if g.RequireJustification {
		queryString += ` AND jsonPayload.metadata.justification_token != ""`
	}
	queryString += " LIMIT 1"
	return utils.MakeQuery(*g.BigQueryClient, u, queryString)
}

// Similar to the above function, but can return multiple results, which is what we expect for streaming.
func (g *GRPC) makeQueryForGRPCStream(u uuid.UUID) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s.%s WHERE jsonPayload.resource_name=?", g.ProjectID, g.DatasetQuery)
	if g.RequireJustification {
		queryString += ` AND jsonPayload.metadata.justification_token != ""`
	}
	return utils.MakeQuery(*g.BigQueryClient, u, queryString)
}
