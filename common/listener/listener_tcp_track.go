package listener

// fork: close accepted TCP connections when the listener closes.
//
// Closing a box only closed the listen socket. Connections a protocol keeps
// open across requests (an anytls session, a gRPC HTTP/2 connection) stayed
// open after the box was gone: the client kept sending new requests into
// them, each one failed at once, and the connection never idled out, so the
// client was stuck until it restarted. In xboard-node the box is restarted
// in-process, so this happened on every kernel restart.
//
// Connections are returned unwrapped (code that relies on the concrete type,
// e.g. splice, keeps working) and only remembered here. Closed ones are
// dropped from the set by a sweep that runs as the set grows, using
// syscall.RawConn.Control, which fails once a connection is closed.

import (
	"net"
	"sync"
	"syscall"
)

type trackingListener struct {
	net.Listener
	access    sync.Mutex
	conns     map[net.Conn]struct{}
	sweepAt   int
	closed    bool
	closeOnce sync.Once
}

func newTrackingListener(l net.Listener) *trackingListener {
	return &trackingListener{Listener: l, conns: make(map[net.Conn]struct{}), sweepAt: 64}
}

func (l *trackingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	if _, ok := conn.(syscall.Conn); !ok {
		return conn, nil
	}
	l.access.Lock()
	if l.closed {
		l.access.Unlock()
		conn.Close()
		return nil, net.ErrClosed
	}
	l.conns[conn] = struct{}{}
	if len(l.conns) >= l.sweepAt {
		for c := range l.conns {
			if connClosed(c) {
				delete(l.conns, c)
			}
		}
		l.sweepAt = max(64, 2*len(l.conns))
	}
	l.access.Unlock()
	return conn, nil
}

func (l *trackingListener) Close() error {
	err := l.Listener.Close()
	l.closeOnce.Do(func() {
		l.access.Lock()
		l.closed = true
		conns := l.conns
		l.conns = nil
		l.access.Unlock()
		for conn := range conns {
			conn.Close()
		}
	})
	return err
}

func connClosed(conn net.Conn) bool {
	raw, err := conn.(syscall.Conn).SyscallConn()
	if err != nil {
		return true
	}
	return raw.Control(func(uintptr) {}) != nil
}
