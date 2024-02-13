package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/ngicks/example-grpc-over-file/api/echoer"
	"github.com/ngicks/example-grpc-over-file/internal/fakeconn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	p          = flag.Bool("p", false, "whether to use stdio or os.Pipe")
	subCmdPath = flag.String("c", "./sub", "path to built sub command")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cmdCtx, cmdCancel := context.WithCancel(ctx)
	defer cmdCancel()
	cmd := subCmd(cmdCtx, *subCmdPath, *p)
	stderr, _ := cmd.StderrPipe()

	var (
		r io.ReadCloser
		w io.WriteCloser
	)
	if *p {
		pr1, pw1, err := os.Pipe()
		if err != nil {
			panic(err)
		}
		pr2, pw2, err := os.Pipe()
		if err != nil {
			panic(err)
		}
		defer func() {
			for _, f := range []*os.File{pr1, pw1, pr2, pw2} {
				_ = f.Close()
			}
		}()
		cmd.ExtraFiles = []*os.File{pr1, pw2}
		r, w = pr2, pw1
	} else {
		r, _ = cmd.StdoutPipe()
		w, _ = cmd.StdinPipe()
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for ctx.Err() == nil && scanner.Scan() {
			fmt.Printf("cmd stderr: %s\n", scanner.Text())
		}
		err := scanner.Err()
		if err != nil && !errors.Is(err, ctx.Err()) {
			fmt.Printf("cmd err: %s\n", err)
		}
	}()

	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	defer func() { _ = cmd.Wait() }()

	fakeConn := fakeconn.New("fake", r, w)
	go fakeConn.Run()
	defer func() { _ = fakeConn.Close() }()

	// Set up a connection to the server.
	conn, err := grpc.DialContext(
		ctx,
		"",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return fakeConn, nil
		}),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := echoer.NewEchoerClient(conn)

	runCtx, runCancel := context.WithTimeout(ctx, 30*time.Second)
	defer runCancel()

	stream, err := c.Echo(runCtx)
	if err != nil {
		panic(err)
	}

	closeOnce := sync.OnceValue(func() error { return stream.CloseSend() })
	defer func() { _ = closeOnce() }()

	for _, msg := range []string{"foo", "bar", "baz"} {
		fmt.Printf("sending %s\n", msg)
		err = stream.Send(&echoer.EchoRequest{Payload: must(anypb.New(&wrappers.StringValue{Value: msg}))})
		if err != nil {
			panic(err)
		}
		fmt.Printf("sent %s\n", msg)
		fmt.Printf("receiving %s\n", msg)
		resp, err := stream.Recv()
		if err != nil {
			panic(err)
		}
		fmt.Printf("received %s\n", msg)
		fmt.Printf("seq = %d, payload = %s\n", resp.GetSeq(), resp.GetPayload().String())
	}

	if err := closeOnce(); err != nil {
		panic(err)
	}
	fmt.Printf("closed\n")

	if runtime.GOOS == "windows" {
		// I'm not sure why but sending SIGTERM on the process blocks long on Windows.
		err = cmd.Process.Kill()
	} else {
		err = cmd.Process.Signal(syscall.SIGTERM)
	}
	if err != nil {
		panic(err)
	}
	err = cmd.Wait()
	fmt.Printf("wait error: %v\n", err)
	fmt.Printf("exit code = %d\n", cmd.ProcessState.ExitCode())
}

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}

func subCmd(ctx context.Context, cmdPath string, usePipe bool) *exec.Cmd {
	args := []string{}
	if usePipe {
		args = append(args, []string{"-r", "3", "-w", "4"}...)
	}
	return exec.CommandContext(ctx, cmdPath, args...)
}
