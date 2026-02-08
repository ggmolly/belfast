package entrypoint

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestProxyBidirectionalPipesBothWays(t *testing.T) {
	clientSide, proxyClient := net.Pipe()
	proxyUpstream, upstreamSide := net.Pipe()
	defer clientSide.Close()
	defer upstreamSide.Close()

	done := make(chan struct{})
	go func() {
		proxyBidirectional(proxyClient, proxyUpstream)
		close(done)
	}()

	if _, err := clientSide.Write([]byte("hello")); err != nil {
		t.Fatalf("write client->proxy: %v", err)
	}
	buf := make([]byte, 5)
	if _, err := io.ReadFull(upstreamSide, buf); err != nil {
		t.Fatalf("read upstream: %v", err)
	}
	if string(buf) != "hello" {
		t.Fatalf("expected upstream to read %q, got %q", "hello", string(buf))
	}

	if _, err := upstreamSide.Write([]byte("world")); err != nil {
		t.Fatalf("write upstream->proxy: %v", err)
	}
	buf = make([]byte, 5)
	if _, err := io.ReadFull(clientSide, buf); err != nil {
		t.Fatalf("read client: %v", err)
	}
	if string(buf) != "world" {
		t.Fatalf("expected client to read %q, got %q", "world", string(buf))
	}

	_ = clientSide.Close()
	select {
	case <-done:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatalf("expected proxy to stop after close")
	}
}
