package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/ngicks/example-grpc-over-file/api/echoer"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 8902, "The server port")
)

type server struct {
	seq atomic.Int64
	echoer.UnimplementedEchoerServer
}

func (s *server) Echo(req echoer.Echoer_EchoServer) error {
	for req.Context().Err() == nil {
		msg, err := req.Recv()
		if err != nil {
			if errors.Is(err, req.Context().Err()) {
				err = nil
			}
			return err
		}
		newSeq := s.seq.Add(1)
		payload := msg.GetPayload()
		fmt.Printf("seq = %d, payload = %s\n", newSeq, payload.String())
		err = req.Send(&echoer.EchoResponse{Seq: newSeq, Payload: payload})
		if err != nil {
			if errors.Is(err, req.Context().Err()) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	echoer.RegisterEchoerServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}
