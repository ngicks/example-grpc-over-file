package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/ngicks/example-grpc-over-file/api/echoer"
	"github.com/ngicks/example-grpc-over-file/internal/fakeconn"
	"google.golang.org/grpc"
)

var (
	r = flag.Int("r", -1, "fd for read. defaults to stdin")
	w = flag.Int("w", -1, "fd for write. defaults to stdout")
)

type server struct {
	printf func(format string, args ...any)
	seq    atomic.Int64
	echoer.UnimplementedEchoerServer
}

func (s *server) Echo(req echoer.Echoer_EchoServer) error {
	s.printf("receiving on echo method\n")
	defer func() {
		s.printf("echo exiting\n")
	}()

	for req.Context().Err() == nil {
		s.printf("receiving on msg\n")
		msg, err := req.Recv()
		if err != nil {
			if errors.Is(err, req.Context().Err()) {
				err = nil
			}
			return err
		}
		s.printf("received on msg\n")
		newSeq := s.seq.Add(1)
		payload := msg.GetPayload()
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// logFile, err := os.OpenFile("./log", os.O_APPEND|os.O_CREATE|os.O_RDWR, fs.ModePerm)
	// if err != nil {
	// 	panic(err)
	// }
	// defer func() {
	// 	_ = logFile.Sync()
	// 	_ = logFile.Close()
	// }()

	log := func(s string, args ...any) {
		// _, _ = fmt.Fprintf(logFile, s, args...)
		_, _ = fmt.Fprintf(os.Stderr, s, args...)
	}

	var rFile, wFile *os.File
	if *r < 0 {
		rFile = os.Stdin
	} else {
		rFile = os.NewFile(uintptr(*r), "in")
	}

	if *w < 0 {
		wFile = os.Stdout
	} else {
		wFile = os.NewFile(uintptr(*w), "out")
	}

	defer func() {
		_ = rFile.Close()
		_ = wFile.Close()
	}()

	log("cmd staring, r = %v, w = %v\n", rFile.Fd(), wFile.Fd())

	fakeConn := fakeconn.New("sub-fake", rFile, wFile)
	go fakeConn.Run()
	defer func() { _ = fakeConn.Close() }()

	fakeListener := fakeconn.NewFakeListener(fakeConn.LocalAddr())
	fakeListener.AddConn(fakeConn)

	s := grpc.NewServer()
	echoer.RegisterEchoerServer(s, &server{printf: log})

	go func() {
		<-ctx.Done()
		log("context signaled\n")
		_ = fakeListener.Close()
		s.Stop()
	}()

	if err := s.Serve(fakeListener); err != nil {
		log("serve error = %v\n", err)
	}
	log("done\n")
}
