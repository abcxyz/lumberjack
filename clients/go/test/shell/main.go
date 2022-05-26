// Copyright 2021 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
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
	"syscall"
	"time"

	"github.com/golang-jwt/jwt"
	cal "google.golang.org/genproto/googleapis/cloud/audit"

	"github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/auditopt"
)

const (
	traceIDKey  = "trace_id"
	serviceName = "go-shell-app"
)

// handler implements ServeHTTP by using the audit client.
type handler struct {
	client *audit.Client
}

// ServeHTTP emits an application audit log with a traceID. To verify
// that our audit logging solution works end-to-end, we can check
// that a log entry with the same traceID successfully reached the
// final log storage (e.g. BigQuery log sink). For context, see
// go/lumberjack-ci-design.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request at %s", r.URL)
	// Get the traceID from URL.
	traceID := r.URL.Query().Get(traceIDKey)
	if traceID == "" {
		http.Error(w, "Missing traceID in the request, add one by appending the URL with ?trace_id=$TRACE_ID.", 400)
		return
	}
	log.Printf("Using trace ID: %v.", traceID)

	// "Bearer " has 7 characters. Trim that.
	idToken := r.Header.Get("Authorization")[7:]
	p := &jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err := p.ParseUnverified(idToken, claims)
	if err != nil {
		http.Error(w, fmt.Sprintf("malformated ID token: %v", err), http.StatusBadRequest)
		return
	}
	email, ok := claims["email"].(string)
	if !ok {
		http.Error(w, fmt.Sprintf("email claim is not a string (got %T)", claims["email"]), http.StatusBadRequest)
		return
	}
	if email == "" {
		http.Error(w, fmt.Sprintf("ID token doesn't have email in the claims: %v", err), http.StatusBadRequest)
		return
	}

	// Generate a minimal and valid AuditLogRequest that
	// stores the traceID in the labels.
	auditLogRequest := &v1alpha1.AuditLogRequest{
		Type: v1alpha1.AuditLogRequest_DATA_ACCESS,
		Payload: &cal.AuditLog{
			ServiceName: serviceName,
			AuthenticationInfo: &cal.AuthenticationInfo{
				PrincipalEmail: email,
			},
		},
		Labels: map[string]string{traceIDKey: traceID},
	}

	// Use the audit client to emit an application audit
	// log synchronously.
	if err := h.client.Log(r.Context(), auditLogRequest); err != nil {
		msg := fmt.Sprintf("Error emiting application audit log: %v.", err)
		http.Error(w, msg, 500)
		return
	}
	success := fmt.Sprintf("Successfully emitted application audit log with trace ID %v.", traceID)
	log.Print(success)
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
	// Create a ServeMux with a handler containing the audit client.
	client, err := audit.NewClient(auditopt.FromConfigFile(""))
	if err != nil {
		return fmt.Errorf("failed to init audit client: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", &handler{
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
		Addr:    ":" + port,
		Handler: mux,
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
