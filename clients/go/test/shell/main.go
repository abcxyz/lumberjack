// Copyright 2021 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main contains a minimal Cloud Run HTTP server that emits an
// application audit log using the audit client.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
	cal "google.golang.org/genproto/googleapis/cloud/audit"

	"github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/auditopt"
	"github.com/abcxyz/lumberjack/clients/go/test/util"
	"github.com/abcxyz/pkg/logging"
)

const (
	traceIDKey  = "trace_id"
	serviceName = "go-shell-app"
)

// handler implements ServeHTTP by using the audit client.
type handler struct {
	logger *zap.SugaredLogger
	client *audit.Client
}

// ServeHTTP emits an application audit log with a traceID. To verify that our
// audit logging solution works end-to-end, we can check that a log entry with
// the same traceID successfully reached the final log storage (e.g. BigQuery
// log sink).
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger

	logger.Debugw("received request", "url", r.URL)

	// Get the traceID from URL.
	traceID := r.URL.Query().Get(traceIDKey)
	if traceID == "" {
		http.Error(w, "missing trace_id in the request, add one by appending the URL with ?trace_id=$TRACE_ID.", http.StatusBadRequest)
		return
	}

	logger = logger.With("trace_id", traceID)
	logger.Debugw("found trace id")

	// Parse the JWT. We do not verify the JWT because:
	//
	//   - This app is only for testing purposes. It should never be used for anything else.
	//   - This runs as a Cloud Run service, and Cloud IAM verifies the JWT.
	//   - Cloud Run actually strips the signature, so there is no signature to verify.
	idToken := idTokenFrom(r.Header.Get("Authorization"))
	token, err := jwt.ParseString(idToken, jwt.WithVerify(false))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id token: %s", err), http.StatusBadRequest)
		return
	}

	// Extract the email claim.
	emailRaw, ok := token.Get("email")
	if !ok {
		http.Error(w, "email claim is missing", http.StatusBadRequest)
		return
	}
	email, ok := emailRaw.(string)
	if !ok {
		http.Error(w, fmt.Sprintf("email claim is not %T (got %T)", "", emailRaw), http.StatusBadRequest)
		return
	}
	if email == "" {
		http.Error(w, "email claim cannot be blank", http.StatusBadRequest)
		return
	}

	jvsToken := strings.TrimSpace(r.Header.Get("justification-token"))
	if jvsToken == "" {
		http.Error(w, "justification-token header cannot be blank", http.StatusBadRequest)
		return
	}

	// Generate a minimal and valid AuditLogRequest that stores the traceID in the
	// labels.
	auditLogRequest := &v1alpha1.AuditLogRequest{
		Type: v1alpha1.AuditLogRequest_DATA_ACCESS,
		Payload: &cal.AuditLog{
			ServiceName:  serviceName,
			ResourceName: traceID,
			AuthenticationInfo: &cal.AuthenticationInfo{
				PrincipalEmail: email,
			},
			MethodName: "loggingShell",
		},
		Labels:             map[string]string{traceIDKey: traceID},
		JustificationToken: jvsToken,
	}

	// Use the audit client to emit an application audit
	// log synchronously.
	if err := h.client.Log(ctx, auditLogRequest); err != nil {
		http.Error(w, fmt.Sprintf("failed to emit application audit log: %s", err), http.StatusInternalServerError)
		return
	}
	success := fmt.Sprintf("Successfully emitted application audit log with trace ID %v.", traceID)

	logger.Debugw("finished request", "success", success)

	fmt.Fprint(w, success) // automatically calls `w.WriteHeader(http.StatusOK)`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := realMain(ctx); err != nil {
		done()
		log.Fatal(err)
	}
}

// realMain creates an HTTP server that emits an application audit log
// with a traceID on the `/` handle. This server supports graceful
// stopping and cancellation by:
//   - using a cancellable context
//   - listening to incoming requests in a goroutine
func realMain(ctx context.Context) error {
	pubKeyEndpoint, shutdown, err := util.StartLocalPublicKeyServer()
	if err != nil {
		return fmt.Errorf("failed to start local public key server: %w", err)
	}
	defer shutdown()

	// Override JVS public key endpoint since we start a local one here in test.
	if err := os.Setenv("AUDIT_CLIENT_JUSTIFICATION_PUBLIC_KEYS_ENDPOINT", pubKeyEndpoint); err != nil {
		return fmt.Errorf("failed to set env: %w", err)
	}

	// Create a ServeMux with a handler containing the audit client.
	client, err := audit.NewClient(ctx, auditopt.FromConfigFile(""))
	if err != nil {
		return fmt.Errorf("failed to init audit client: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", &handler{
		logger: logging.NewFromEnv("SHELL_"),
		client: client,
	})

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Create the server and listen in a goroutine.
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}
	serverErrCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case serverErrCh <- err:
			default:
			}
		}
	}()

	// Wait for shutdown signal or error from the listener.
	select {
	case err := <-serverErrCh:
		return fmt.Errorf("error from server listener: %w", err)
	case <-ctx.Done():
	}

	// Gracefully shut down the server.
	shutdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	return nil
}

// idTokenFrom extracts the ID token from the given input string. It assumes the
// input string is from a header that might include the "Bearer" prefix.
func idTokenFrom(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 7 && strings.EqualFold(s[:7], "bearer ") {
		return strings.TrimSpace(s[7:])
	}
	return ""
}
