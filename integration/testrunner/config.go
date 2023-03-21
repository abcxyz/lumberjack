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

package testrunner

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/abcxyz/lumberjack/internal/talkerpb"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sethvargo/go-envconfig"

	jvspb "github.com/abcxyz/jvs/apis/v0"
)

// Config is the global configuration for integration tests.
type Config struct {
	AuditLogRequestTimeout  time.Duration `env:"AUDIT_CLIENT_TEST_AUDIT_LOG_REQUEST_TIMEOUT,default=30s"`
	AuditLogRequestWait     time.Duration `env:"AUDIT_CLIENT_TEST_AUDIT_LOG_REQUEST_WAIT,default=4s"`
	HTTPEndpoints           string        `env:"HTTP_ENDPOINTS,required"`
	GRPCEndpoints           string        `env:"GRPC_ENDPOINTS,required"`
	LogRoutingWait          time.Duration `env:"AUDIT_CLIENT_TEST_AUDIT_LOG_ROUTING_WAIT,default=5s"`
	MaxAuditLogRequestTries uint64        `env:"AUDIT_CLIENT_TEST_MAX_AUDIT_LOG_REQUEST_TRIES,default=4"`
	MaxDBQueryTries         uint64        `env:"AUDIT_CLIENT_TEST_MAX_DB_QUERY_TRIES,default=60"`
	JustificationSubject    string        `env:"AUDIT_CLIENT_TEST_JUSTIFICATION_SUB,required"`
	IDToken                 string        `env:"AUDIT_CLIENT_TEST_IDTOKEN"`
	ProjectID               string        `env:"AUDIT_CLIENT_TEST_PROJECT_ID,required"`
	BigQueryDataset         string        `env:"AUDIT_CLIENT_TEST_BIGQUERY_DATASET,required"`
	PrivateKeyFilePath      string        `env:"AUDIT_CLIENT_TEST_PRIVATE_KEY_PATH,required"`
	PrivateKey              ecdsa.PrivateKey
}

// TestCaseConfig contains all configuration needed in a test case.
type TestCaseConfig struct {
	*Config

	Endpoint string
	TraceID  string

	BigQueryClient *bigquery.Client

	// For gRPC endpoint testing only.
	TalkerClient talkerpb.TalkerClient
}

type privateKeyJSONData struct {
	Encoded string
}

func parsePrivateKey(path string) (*ecdsa.PrivateKey, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w from file: %s", err, path)
	}
	b, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from key file: %w", err)
	}
	var data privateKeyJSONData
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to privateKeyJSONData: %w", err)
	}
	privateKeyPEM, _ := pem.Decode([]byte(strings.TrimSpace(data.Encoded)))
	privateKey, err := x509.ParseECPrivateKey(privateKeyPEM.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}
	return privateKey, nil
}

func newTestConfig(ctx context.Context) (*Config, error) {
	var c Config
	if err := envconfig.ProcessWith(ctx, &c, envconfig.OsLookuper()); err != nil {
		return nil, fmt.Errorf("failed to process environment: %w", err)
	}
	PrivateKey, err := parsePrivateKey(c.PrivateKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	c.PrivateKey = *PrivateKey

	return &c, nil
}

// create a justification token to pass in the call to services.
func justificationToken(audience, subject string, key *ecdsa.PrivateKey) (string, error) {
	now := time.Now().UTC()

	token, err := jwt.NewBuilder().
		Audience([]string{audience}).
		Expiration(now.Add(time.Hour)).
		JwtID(uuid.New().String()).
		IssuedAt(now).
		Issuer("lumberjack-test-runner").
		NotBefore(now).
		Subject(subject).
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
	b, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, key,
		jws.WithProtectedHeaders(headers)))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return string(b), nil
}
