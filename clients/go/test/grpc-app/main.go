// Copyright 2022 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/auditopt"
	"github.com/abcxyz/lumberjack/clients/go/test/grpc-app/talkerpb"
)

var port = flag.Int("port", 8080, "The server port")

// Matching private key here: https://github.com/abcxyz/lumberjack/pull/261/files#diff-499009010e3b24ec9e364f0c66e4f3e88898ea2788f81f8a866a81944a8655fbR57
const pubKey = "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEhBWj8vw5LkPRWbCr45k0cOarIcWg\nApM03mSYF911de5q1wGOL7R9N8pC7jo2xbS+i1wGsMiz+AWnhhZIQcNTKg==\n-----END PUBLIC KEY-----\n"

func main() {
	if err := realMain(); err != nil {
		log.Fatal(err)
	}
}

func realMain() (outErr error) {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	pubKeyEndpoint, shutdown, err := startLocalPublicKeyServer()
	if err != nil {
		return err
	}
	defer shutdown()

	// Override JVS public key endpoint since we start a local one here in test.
	if err := os.Setenv("AUDIT_CLIENT_JUSTIFICATION_PUBLIC_KEYS_ENDPOINT", pubKeyEndpoint); err != nil {
		return err
	}

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to create jvs client: %w", err)
	}
	interceptor, err := audit.NewInterceptor(
		auditopt.InterceptorFromConfigFile(ctx, auditopt.DefaultConfigFilePath))
	if err != nil {
		return fmt.Errorf("failed to setup audit interceptor: %w", err)
	}
	defer func() {
		if err := interceptor.Stop(); err != nil {
			outErr = fmt.Errorf("failed to stop interceptor: %w", err)
		}
	}()
	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryInterceptor), grpc.StreamInterceptor(interceptor.StreamInterceptor))
	talkerpb.RegisterTalkerServer(s, &server{})
	// Register the reflection service makes it easier for some clients.
	reflection.Register(s)

	// Gracefully stop the server on ctrl-c.
	intrCh := make(chan os.Signal, 1)
	signal.Notify(intrCh, os.Interrupt)
	go func() {
		<-intrCh
		log.Println("gracefully stopping...")
		s.GracefulStop()
	}()

	log.Printf("server listening at %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	log.Println("server stopped.")

	return nil
}

// Parse pre-made key and set up a server to host it in JWKS format.
// This is intended to stand in for the JVS in the integration tests.
func startLocalPublicKeyServer() (string, func(), error) {
	block, _ := pem.Decode([]byte(pubKey))
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Printf("Err when parsing key %v", err)
		return "", nil, err
	}
	ecdsaKey, err := jwk.FromRaw(key)
	if err != nil {
		log.Printf("Err when converting key to jwk %v", err)
		return "", nil, err
	}
	if err := ecdsaKey.Set(jwk.KeyIDKey, "integ-key"); err != nil {
		log.Printf("Err when setting key id %v", err)
		return "", nil, err
	}

	jwks := make(map[string][]jwk.Key)
	jwks["keys"] = []jwk.Key{ecdsaKey}
	j, err := json.MarshalIndent(jwks, "", " ")
	if err != nil {
		log.Printf("Err when creating jwks json %v", err)
		return "", nil, err
	}
	path := "/.well-known/jwks"
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", j)
	})
	svr := httptest.NewServer(mux)
	return svr.URL + path, func() { svr.Close() }, nil
}

type server struct {
	talkerpb.UnimplementedTalkerServer
}

func (s *server) Hello(ctx context.Context, req *talkerpb.HelloRequest) (*talkerpb.HelloResponse, error) {
	if logReq, ok := audit.LogReqFromCtx(ctx); ok {
		logReq.Payload.ResourceName = req.Target
	}
	return &talkerpb.HelloResponse{
		Message: fmt.Sprintf("Hi, I'm %s!", req.Target),
	}, nil
}

func (s *server) Whisper(ctx context.Context, req *talkerpb.WhisperRequest) (*talkerpb.WhisperResponse, error) {
	if logReq, ok := audit.LogReqFromCtx(ctx); ok {
		logReq.Payload.ResourceName = req.Target
	}
	return &talkerpb.WhisperResponse{
		Message: fmt.Sprintf("Shush, I'm %s.", req.Target),
	}, nil
}

func (s *server) Bye(ctx context.Context, req *talkerpb.ByeRequest) (*talkerpb.ByeResponse, error) {
	if logReq, ok := audit.LogReqFromCtx(ctx); ok {
		logReq.Payload.ResourceName = req.Target
	}
	return &talkerpb.ByeResponse{
		Message: "Bye!",
	}, nil
}

func (s *server) Fibonacci(req *talkerpb.FibonacciRequest, svr talkerpb.Talker_FibonacciServer) error {
	if logReq, ok := audit.LogReqFromCtx(svr.Context()); ok {
		logReq.Payload.ResourceName = req.Target
	}

	var x, y uint32 = 0, 1
	for i := uint32(1); i <= req.Places; i++ {
		z := uint32(0)
		if i == 2 {
			z = 1
		} else if i > 2 {
			z = x + y
			x = y
			y = z
		}

		if err := svr.Send(&talkerpb.FibonacciResponse{
			Position: i,
			Value:    z,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) Addition(svr talkerpb.Talker_AdditionServer) error {
	sum := 0

	for {
		req, err := svr.Recv()
		if errors.Is(err, io.EOF) {
			// End of stream. Send the sum.
			if err := svr.SendAndClose(&talkerpb.AdditionResponse{
				Sum: uint64(sum),
			}); err != nil {
				return err
			}
			break
		}
		if err != nil {
			return err
		}

		if logReq, ok := audit.LogReqFromCtx(svr.Context()); ok {
			logReq.Payload.ResourceName = req.Target
		}
		sum += int(req.Addend)
	}

	return nil
}

func (s *server) Fail(ctx context.Context, req *talkerpb.FailRequest) (*talkerpb.FailResponse, error) {
	if logReq, ok := audit.LogReqFromCtx(ctx); ok {
		logReq.Payload.ResourceName = req.Target
	}
	return nil, status.Errorf(codes.ResourceExhausted, "this call will always fail")
}

func (s *server) FailOnFour(svr talkerpb.Talker_FailOnFourServer) error {
	for {
		req, err := svr.Recv()
		if errors.Is(err, io.EOF) {
			svr.SendAndClose(&talkerpb.FailOnFourResponse{
				Message: "closing...",
			})
			break
		}
		if err != nil {
			return err
		}

		if logReq, ok := audit.LogReqFromCtx(svr.Context()); ok {
			logReq.Payload.ResourceName = req.Target
		}

		if req.Value == 4 {
			return status.Errorf(codes.InvalidArgument, "this call will always fail on four")
		}
	}
	return nil
}
