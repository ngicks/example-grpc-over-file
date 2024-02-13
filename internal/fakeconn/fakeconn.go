package fakeconn

import (
	"context"
	"io"
	"net"
	"sync"
	"time"
)

var _ net.Conn = (*FakeConn)(nil)

// FakeConn is a faked connection which reads from a reader and writes to a writer.
type FakeConn struct {
	name    string
	rFeeder *feeder
	wFeeder *feeder
	cancel  func()
}

func New(name string, r io.Reader, w io.Writer) *FakeConn {
	ctx, cancel := context.WithCancel(context.Background())
	return &FakeConn{
		name:    name,
		rFeeder: newFeeder(ctx, r.Read),
		wFeeder: newFeeder(ctx, w.Write),
		cancel:  cancel,
	}
}

// Run runs FakeConn and blocks the current goroutine until c is closed.
// Calling Run twice or more causes undefined behavior.
//
// Run starts 2 additional goroutines.
func (c *FakeConn) Run() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.rFeeder.run()
	}()
	go func() {
		defer wg.Done()
		c.wFeeder.run()
	}()
	wg.Wait()
}

func (c *FakeConn) Read(b []byte) (n int, err error)  { return c.rFeeder.do(b) }
func (c *FakeConn) Write(b []byte) (n int, err error) { return c.wFeeder.do(b) }
func (c *FakeConn) Close() error {
	c.cancel()
	return nil
}
func (c *FakeConn) LocalAddr() net.Addr                { return fakeAddr(c.name) }
func (c *FakeConn) RemoteAddr() net.Addr               { return fakeAddr(c.name) }
func (c *FakeConn) SetDeadline(t time.Time) error      { return nil } // TODO: maybe implement?
func (c *FakeConn) SetReadDeadline(t time.Time) error  { return nil } // TODO: maybe implement?
func (c *FakeConn) SetWriteDeadline(t time.Time) error { return nil } // TODO: maybe implement?

// feeder serializes io operations.
type feeder struct {
	ctx      context.Context
	fn       func(b []byte) (int, error)
	bufCh    chan []byte
	resultCh chan feederResult
}

type feederResult struct {
	n   int
	err error
}

func newFeeder(ctx context.Context, fn func(b []byte) (int, error)) *feeder {
	return &feeder{
		ctx:      ctx,
		fn:       fn,
		bufCh:    make(chan []byte),
		resultCh: make(chan feederResult),
	}
}

func (f *feeder) do(b []byte) (n int, err error) {
	select {
	case f.bufCh <- b:
	case <-f.ctx.Done():
		return 0, io.EOF
	}
	select {
	case result := <-f.resultCh:
		return result.n, result.err
	case <-f.ctx.Done():
		return 0, io.EOF
	}
}

func (f *feeder) run() {
	for {
		var buf []byte
		select {
		case <-f.ctx.Done():
			return
		case buf = <-f.bufCh:
		}
		n, err := f.fn(buf)
		select {
		case <-f.ctx.Done():
			return
		case f.resultCh <- feederResult{n, err}:
		}
	}
}

var _ net.Addr = fakeAddr("")

type fakeAddr string

func (a fakeAddr) Network() string { return "fake" }
func (a fakeAddr) String() string  { return string(a) }
