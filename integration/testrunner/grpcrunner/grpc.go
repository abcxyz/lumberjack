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
	"github.com/abcxyz/lumberjack/integration/testrunner/utils"
	"github.com/abcxyz/lumberjack/internal/talkerpb"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
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

// Matching public key here: https://github.com/abcxyz/lumberjack/blob/92782c326681157221df37e0897ba234c5a22240/clients/go/test/grpc-app/main.go#L47
const privateKeyString = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIITZ4357UsTCbhxXu8w8cY54ZLlsAIJj/Aej9ylb/ZfBoAoGCCqGSM49
AwEHoUQDQgAEhBWj8vw5LkPRWbCr45k0cOarIcWgApM03mSYF911de5q1wGOL7R9
N8pC7jo2xbS+i1wGsMiz+AWnhhZIQcNTKg==
-----END EC PRIVATE KEY-----
`

var (
	privateKeyPEM, _ = pem.Decode([]byte(strings.TrimSpace(privateKeyString)))
	privateKey, _    = x509.ParseECPrivateKey(privateKeyPEM.Bytes)
)

// TestGRPCEndpoint runs tests against a GRPC endpoint. The given GRPC must
// define a projectID and datasetQuery. If a TalkerClient or BigQueryClient are
// not provided, they are instantiated via the defaults.
//
//nolint:thelper // Not really a helper.
func CheckGRPCEndpoint(ctx context.Context, t *testing.T, g *GRPC) {
	// Don't mark t.Helper().
	// Here locates the actual test logic so we want to be able to locate the
	// actual line of error here instead of the main test.

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

		conn := createConnection(ctx, t, g.EndpointURL, g.IDToken)
		t.Cleanup(func() {
			conn.Close()
		})
		g.TalkerClient = talkerpb.NewTalkerClient(conn)
	}

	signedToken, err := justificationToken()
	if err != nil {
		t.Fatalf("couldn't generate justification token: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	md.Set("justification-token", signedToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	if g.BigQueryClient == nil {
		bqClient, err := utils.MakeClient(ctx, g.ProjectID)
		if err != nil {
			t.Errorf("BigQuery request failed: %v", err)
		}
		t.Cleanup(func() {
			if err := bqClient.Close(); err != nil {
				t.Errorf("Failed to close the BQ client: %v", err)
			}
		})
		g.BigQueryClient = bqClient
	}

	t.Run("hello_req_unary_success", func(t *testing.T) {
		t.Parallel()
		id := uuid.New().String()
		_, err := g.TalkerClient.Hello(ctx, &talkerpb.HelloRequest{Message: "Some Message", Target: id})
		if err != nil {
			t.Errorf("could not greet: %v", err)
		}
		query := g.makeQueryForGRPCUnary(id)
		utils.QueryIfAuditLogExistsWithRetries(ctx, t, query, g.Config, "unary_hello")
	})

	t.Run("fail_req_unary_failure", func(t *testing.T) {
		t.Parallel()
		id := uuid.New().String()
		reply, err := g.TalkerClient.Fail(ctx, &talkerpb.FailRequest{Message: "Some Message", Target: id})

		if err != nil {
			returnStatus, ok := status.FromError(err)
			if !ok {
				t.Errorf("Could not convert err to status %v", err)
			}
			if int32(returnStatus.Code()) != int32(rpccode.Code_RESOURCE_EXHAUSTED) {
				t.Errorf("Got unexpected Err. Got code %d but expected %d", int32(returnStatus.Code()),
					int32(rpccode.Code_RESOURCE_EXHAUSTED))
			}
			t.Logf("Got Error as expected: %v", err)
		} else {
			t.Errorf("Did not get err as expected. Instead got reply: %v", reply)
		}

		query := g.makeQueryForGRPCUnary(id)
		utils.QueryIfAuditLogExistsWithRetries(ctx, t, query, g.Config, "unary_fail")
	})

	t.Run("fibonacci_req_server_streaming_success", func(t *testing.T) {
		t.Parallel()
		id := uuid.New().String()
		places := 5
		stream, err := g.TalkerClient.Fibonacci(ctx, &talkerpb.FibonacciRequest{Places: uint32(places), Target: id})
		if err != nil {
			t.Errorf("fibonacci call failed: %v", err)
		}
		for {
			place, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				// stream is finished
				break
			}
			if err != nil {
				t.Errorf("failed to read fibonacci stream: %v", err)
				break
			}
			t.Logf("Received value %v", place.Value)
		}
		query := g.makeQueryForGRPCStream(id)
		utils.QueryIfAuditLogsExistWithRetries(ctx, t, query, g.Config, "server_stream_fibonacci", int64(places))
	})

	t.Run("addition_req_client_streaming_success", func(t *testing.T) {
		t.Parallel()
		id := uuid.New().String()
		stream, err := g.TalkerClient.Addition(ctx)
		if err != nil {
			t.Errorf("addition call failed: %v", err)
		}
		totalNumbers := 5
		for i := 0; i < totalNumbers; i++ {
			if err := stream.Send(&talkerpb.AdditionRequest{
				Addend: uint32(i),
				Target: id,
			}); err != nil {
				t.Errorf("sending value to addition failed: %v", err)
			}
		}
		reply, err := stream.CloseAndRecv()
		if err != nil {
			t.Errorf("failed getting result from addition: %v", err)
		}
		t.Logf("Value returned: %d", reply.Sum)

		query := g.makeQueryForGRPCStream(id)
		utils.QueryIfAuditLogsExistWithRetries(ctx, t, query, g.Config, "client_stream_addition", int64(totalNumbers))
	})

	t.Run("fail_on_four_req_client_stream_failure", func(t *testing.T) {
		t.Parallel()
		id := uuid.New().String()
		stream, err := g.TalkerClient.FailOnFour(ctx)
		if err != nil {
			t.Errorf("addition call failed: %v", err)
		}
		totalNumbers := 5
		for i := 1; i <= totalNumbers; i++ {
			if err := stream.Send(&talkerpb.FailOnFourRequest{
				Value:  uint32(i),
				Target: id,
			}); err != nil {
				t.Errorf("sending value to addition failed: %v", err)
			}
		}
		reply, err := stream.CloseAndRecv()
		if err != nil {
			returnStatus, ok := status.FromError(err)
			if !ok {
				t.Errorf("Could not convert err to status %v", err)
			}
			if int32(returnStatus.Code()) != int32(rpccode.Code_INVALID_ARGUMENT) {
				t.Errorf("Got unexpected Err. Got code %d but expected %d", int32(returnStatus.Code()),
					int32(rpccode.Code_INVALID_ARGUMENT))
			}
			t.Logf("Got Error as expected: %v", err)
		} else {
			t.Errorf("Did not get err as expected. Instead got reply: %v", reply)
		}

		query := g.makeQueryForGRPCStream(id)
		// we expect to have 4 audit logs - the last sent number (5) will be after the err ocurred.
		utils.QueryIfAuditLogsExistWithRetries(ctx, t, query, g.Config, "stream_fail_on_four", int64(4))
	})
}

// create a justification token to pass in the call to services.
func justificationToken() (string, error) {
	now := time.Now().UTC()

	token, err := jwt.NewBuilder().
		Audience([]string{"talker-app"}).
		Expiration(now.Add(time.Hour)).
		JwtID(uuid.New().String()).
		IssuedAt(now).
		Issuer("lumberjack-test-runner").
		NotBefore(now).
		Subject("lumberjack-integ").
		Build()
	if err != nil {
		return "", fmt.Errorf("failed to build justification token: %w", err)
	}

	if err := jvspb.SetJustifications(token, []*jvspb.Justification{
		{
			Category: "test",
			Value:    "test",
		},
	}); err != nil {
		return "", fmt.Errorf("failed to set justifications: %w", err)
	}

	// Build custom headers and set the "kid" as the signer ID.
	headers := jws.NewHeaders()
	if err := headers.Set(jws.KeyIDKey, "integ-key"); err != nil {
		return "", fmt.Errorf("failed to set header: %w", err)
	}

	// Sign the token.
	b, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, privateKey,
		jws.WithProtectedHeaders(headers)))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return string(b), nil
}

// Server is in cloud run. Example: https://cloud.google.com/run/docs/triggering/grpc#request-auth
// We are using token-based authentication to connect to the server, which will be passed through
// a JWT to the server. There, the server will be able to decipher the JWT to find the calling user.
func createConnection(ctx context.Context, t *testing.T, addr string, idToken string) *grpc.ClientConn {
	t.Helper()

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
	conn, err := grpc.DialContext(ctx, addr, grpc.WithPerRPCCredentials(rpcCreds), grpc.WithTransportCredentials(creds))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	return conn
}

// This query is used to find the relevant audit log in BigQuery, which we assume will be added by the server.
// We specifically look up the log using the UUID specified in the request as we know the server will add that
// as the resource name, and provides us a unique key to find logs with.
func (g *GRPC) makeQueryForGRPCUnary(id string) *bigquery.Query {
	// Cast to int64 because the result checker expects a number.
	queryString := fmt.Sprintf("SELECT CAST(EXISTS (SELECT * FROM `%s.%s` WHERE jsonPayload.resource_name=?", g.ProjectID, g.DatasetQuery)
	if g.RequireJustification {
		// TODO(#265): For back-compatibility, we need to ensure justification in either metadata key.
		queryString += ` AND jsonPayload.metadata.justification IS NOT NULL`
	}
	queryString += ` AND jsonPayload.service_name IS NOT NULL`
	queryString += ` AND jsonPayload.method_name IS NOT NULL`
	queryString += ` AND jsonPayload.authentication_info.principal_email IS NOT NULL`
	queryString += ") AS INT64)"
	return utils.MakeQuery(*g.BigQueryClient, id, queryString)
}

// Similar to the above function, but can return multiple results, which is what we expect for streaming.
func (g *GRPC) makeQueryForGRPCStream(id string) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT count(*) FROM `%s.%s` WHERE jsonPayload.resource_name=?", g.ProjectID, g.DatasetQuery)
	if g.RequireJustification {
		// TODO(#265): For back-compatibility, we need to ensure justification in either metadata key.
		queryString += ` AND jsonPayload.metadata.justification IS NOT NULL`
	}
	queryString += ` AND jsonPayload.service_name IS NOT NULL`
	queryString += ` AND jsonPayload.method_name IS NOT NULL`
	queryString += ` AND jsonPayload.authentication_info.principal_email IS NOT NULL`
	return utils.MakeQuery(*g.BigQueryClient, id, queryString)
}
