package answer_test

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
)

func TestMaintenanceBlocksNewConnections(t *testing.T) {
	server := connection.NewServer("127.0.0.1", 0, func(*[]byte, *connection.Client, int) {})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()
	server.SetMaintenance(true)
	done := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			close(done)
			return
		}
		server.HandleConnection(conn)
		close(done)
	}()
	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer clientConn.Close()
	_ = clientConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 1)
	n, err := clientConn.Read(buf)
	if err == nil {
		t.Fatalf("expected connection to close")
	}
	if n != 0 || err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected handler to finish")
	}
}
