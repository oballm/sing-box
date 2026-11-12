package listener

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestTrackingListenerClosesAcceptedConnections(t *testing.T) {
	inner, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	l := newTrackingListener(inner)
	client, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	server, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := server.(*net.TCPConn); !ok {
		t.Fatalf("accepted conn is %T, want the unwrapped *net.TCPConn", server)
	}
	if err = l.Close(); err != nil {
		t.Fatal(err)
	}
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err = client.Read(make([]byte, 1)); err != io.EOF {
		t.Fatalf("client read after listener close = %v, want EOF", err)
	}
	if err = l.Close(); err != nil && !isClosed(err) {
		t.Fatalf("second close: %v", err)
	}
}

func TestTrackingListenerDropsClosedConnections(t *testing.T) {
	inner, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	l := newTrackingListener(inner)
	defer l.Close()
	for i := 0; i < 500; i++ {
		client, err := net.Dial("tcp", l.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		server, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}
		server.Close()
		client.Close()
	}
	l.access.Lock()
	n := len(l.conns)
	l.access.Unlock()
	if n >= 64 {
		t.Fatalf("tracked %d connections after all were closed; sweep did not run", n)
	}
}

func isClosed(err error) bool {
	ne, ok := err.(*net.OpError)
	return ok && ne.Err != nil && ne.Err.Error() == "use of closed network connection"
}
