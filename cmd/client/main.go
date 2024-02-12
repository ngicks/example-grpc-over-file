package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/ngicks/example-grpc-over-file/api/echoer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	addr = flag.String("addr", "localhost:8902", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := echoer.NewEchoerClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stream, err := c.Echo(ctx)
	if err != nil {
		panic(err)
	}
	defer func() { _ = stream.CloseSend() }()

	for _, msg := range []string{"foo", "bar", "baz"} {
		err = stream.Send(&echoer.EchoRequest{Payload: must(anypb.New(&wrappers.StringValue{Value: msg}))})
		if err != nil {
			panic(err)
		}
		resp, err := stream.Recv()
		if err != nil {
			panic(err)
		}
		fmt.Printf("seq = %d, payload = %s\n", resp.GetSeq(), resp.GetPayload().String())
	}
}

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}
