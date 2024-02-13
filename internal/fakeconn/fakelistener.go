package fakeconn

import (
	"net"
	"slices"
	"sync"
)

var _ net.Listener = (*FakeListener)(nil)

type FakeListener struct {
	addr      net.Addr
	mu        sync.Mutex
	done      chan struct{}
	closeOnce func()
	updateCh  chan struct{}
	queue     []net.Conn
}

// NewFakeListener returns a newly allocated *FakeListener.
func NewFakeListener(addr net.Addr) *FakeListener {
	done := make(chan struct{})
	return &FakeListener{
		addr:      addr,
		done:      done,
		closeOnce: sync.OnceFunc(func() { close(done) }),
		updateCh:  make(chan struct{}),
	}
}

// AddConn adds
func (l *FakeListener) AddConn(conn net.Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.queue = append(l.queue, conn)
	l.notify()
}

func (l *FakeListener) notifier() chan struct{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.updateCh
}

func (l *FakeListener) notify() {
	ch := l.updateCh
	l.updateCh = make(chan struct{})
	close(ch)
}

func (l *FakeListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	select {
	case <-l.done:
		l.mu.Unlock()
		// TODO: use a more appropriate error
		return nil, net.ErrClosed
	default:
	}
	if len(l.queue) > 0 {
		defer l.mu.Unlock()
		return l.consumeOne()
	}
	l.mu.Unlock()

	for {
		select {
		case <-l.done:
			return nil, net.ErrClosed
		case <-l.notifier():
			l.mu.Lock()
			if len(l.queue) > 0 {
				defer l.mu.Unlock()
				return l.consumeOne()
			}
			l.mu.Unlock()
		}
	}
}

func (l *FakeListener) consumeOne() (net.Conn, error) {
	conn := l.queue[0]
	l.queue = slices.Delete(l.queue, 0, 1)
	return conn, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *FakeListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closeOnce()
	return nil
}

// Addr returns the listener's network address.
func (l *FakeListener) Addr() net.Addr {
	return l.addr
}
