package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	pb "github.com/abcxyz/lumberjack/clients/go/test/grpc-app/talkerpb"
)

const (
	address        = "grpc-talker-server-kszciupxha-uw.a.run.app:443" // unary server cloud run address
	defaultTarget  = "target"
	defaultMessage = "message"
	insecure       = false
)

func main() {
	opts := []grpc.DialOption{
		grpc.WithAuthority(address),
		grpc.WithBlock(),
	}
	if insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			log.Fatal(err)
		}
		cred := credentials.NewTLS(&tls.Config{
			RootCAs: systemRoots,
		})
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTalkerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// identity-token can be printed with `gcloud auth print-identity-token`
	if len(os.Args) < 2 {
		log.Fatalf("usage %s $(gcloud auth print-identity-token)", os.Args[0])
	}
	identityToken := os.Args[1]
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+identityToken)

	callMethods(ctx, c)
}

func callMethods(ctx context.Context, c pb.TalkerClient) {
	helloResp, err := c.Hello(ctx, &pb.HelloRequest{Target: defaultTarget, Message: defaultMessage})
	if err != nil {
		log.Fatalf("could not c.Hello: %v", err)
	}
	log.Printf("c.Hello: %s", helloResp.GetMessage())

	whisperResp, err := c.Whisper(ctx, &pb.WhisperRequest{Target: defaultTarget, Message: defaultMessage})
	if err != nil {
		log.Fatalf("could not c.Whisper: %v", err)
	}
	log.Printf("c.Whisper: %s", whisperResp.GetMessage())

	byeResp, err := c.Bye(ctx, &pb.ByeRequest{Target: defaultTarget, Message: defaultMessage})
	if err != nil {
		log.Fatalf("could not c.Bye: %v", err)
	}
	log.Printf("c.Bye: %s", byeResp.GetMessage())
}
