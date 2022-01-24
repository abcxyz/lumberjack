package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/abcxyz/lumberjack/clients/go/test/grpc-app/talkerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.Int("port", 8080, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	talkerpb.RegisterTalkerServer(s, &server{})
	// Register the reflection service makes it easier for some clients.
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	talkerpb.UnimplementedTalkerServer
}

func (s *server) Hello(ctx context.Context, req *talkerpb.HelloRequest) (*talkerpb.HelloResponse, error) {
	return &talkerpb.HelloResponse{
		Message: fmt.Sprintf("Hi, I'm %s!", req.Target),
	}, nil
}

func (s *server) Whisper(ctx context.Context, req *talkerpb.WhisperRequest) (*talkerpb.WhisperResponse, error) {
	return &talkerpb.WhisperResponse{
		Message: fmt.Sprintf("Shush, I'm %s.", req.Target),
	}, nil
}

func (s *server) Bye(ctx context.Context, req *talkerpb.ByeRequest) (*talkerpb.ByeResponse, error) {
	return &talkerpb.ByeResponse{
		Message: "Bye!",
	}, nil
}
