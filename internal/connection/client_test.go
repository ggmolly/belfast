package connection

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientInitQueues(t *testing.T) {
	client := &Client{}
	client.initQueues()

	if client.queueLimit != packetQueueSize {
		t.Fatalf("expected queue limit %d, got %d", packetQueueSize, client.queueLimit)
	}
	if client.packetQueue == nil {
		t.Fatalf("expected packet queue to be initialized")
	}
	if client.queueCond == nil {
		t.Fatalf("expected queue condition variable to be initialized")
	}
	if client.packetPool == nil {
		t.Fatalf("expected packet pool to be initialized")
	}
	if cap(client.packetPool) != packetPoolSize {
		t.Fatalf("expected packet pool capacity %d, got %d", packetPoolSize, cap(client.packetPool))
	}
}

func TestClientEnqueuePacket(t *testing.T) {
	client := &Client{}
	client.initQueues()

	packet := []byte{1, 2, 3, 4, 5}
	err := client.EnqueuePacket(packet)
	if err != nil {
		t.Fatalf("expected no error enqueuing packet, got %v", err)
	}

	if client.packetQueue.Length() != 1 {
		t.Fatalf("expected queue length 1, got %d", client.packetQueue.Length())
	}

	dequeued := client.packetQueue.Remove().([]byte)
	if !equalBytes(dequeued, packet) {
		t.Fatalf("expected dequeued packet %v, got %v", packet, dequeued)
	}
}

func TestClientEnqueueMultiplePackets(t *testing.T) {
	client := &Client{}
	client.initQueues()

	packets := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	for _, pkt := range packets {
		if err := client.EnqueuePacket(pkt); err != nil {
			t.Fatalf("expected no error enqueuing packet, got %v", err)
		}
	}

	if client.packetQueue.Length() != len(packets) {
		t.Fatalf("expected queue length %d, got %d", len(packets), client.packetQueue.Length())
	}
}

func TestClientEnqueueAfterClose(t *testing.T) {
	client := &Client{}
	client.initQueues()
	client.closed = true

	packet := []byte{1, 2, 3}
	err := client.EnqueuePacket(packet)
	if !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed, got %v", err)
	}
}

func TestClientEnqueueOverflow(t *testing.T) {
	client := &Client{}
	client.initQueues()

	largePacket := make([]byte, 100)

	enqueued := 0
	for i := 0; i < packetQueueSize+10; i++ {
		if err := client.EnqueuePacket(largePacket); err == nil {
			enqueued++
		}
	}

	if enqueued != packetQueueSize {
		t.Fatalf("expected %d packets to be enqueued, got %d", packetQueueSize, enqueued)
	}

	initialQueueBlocks := client.metrics.queueBlocks

	if err := client.EnqueuePacket(largePacket); err == nil {
		t.Fatalf("expected blocking behavior to prevent overflow")
	}

	if client.metrics.queueBlocks <= initialQueueBlocks {
		t.Fatalf("expected queue blocks to increase, got %d", client.metrics.queueBlocks)
	}
}

func TestClientClose(t *testing.T) {
	client := &Client{}
	client.initQueues()
	client.closed = false

	client.Close()

	if !client.closed {
		t.Fatalf("expected client to be closed")
	}

	if err := client.EnqueuePacket([]byte{1, 2}); !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed after close, got %v", err)
	}
}

func TestClientCloseWithError(t *testing.T) {
	client := &Client{}
	client.initQueues()
	client.closed = false

	testErr := errors.New("test error")
	client.CloseWithError(testErr)

	if !client.closed {
		t.Fatalf("expected client to be closed")
	}

	if err := client.EnqueuePacket([]byte{1, 2}); !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed after close with error, got %v", err)
	}
}

func TestClientCloseOnce(t *testing.T) {
	client := &Client{}
	client.initQueues()

	callCount := 0
	client.closeOnce = sync.Once{}
	client.closeOnce.Do(func() {
		callCount++
	})
	client.closeOnce.Do(func() {
		callCount++
	})

	if callCount != 1 {
		t.Fatalf("expected closeOnce to execute once, got %d", callCount)
	}
}

func TestClientIsClosed(t *testing.T) {
	client := &Client{}
	client.initQueues()

	if client.IsClosed() {
		t.Fatalf("expected client to not be closed initially")
	}

	client.closed = true

	if !client.IsClosed() {
		t.Fatalf("expected client to be closed")
	}
}

func TestClientIsClosedThreadSafe(t *testing.T) {
	client := &Client{}
	client.initQueues()

	done := make(chan bool)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.IsClosed()
		}()
	}

	go func() {
		client.queueMu.Lock()
		client.closed = true
		client.queueMu.Unlock()
		done <- true
	}()

	<-done
	wg.Wait()

	if !client.IsClosed() {
		t.Fatalf("expected client to be closed")
	}
}

func TestClientRecordHandlerError(t *testing.T) {
	client := &Client{}
	client.initQueues()

	initialErrors := atomic.LoadUint64(&client.metrics.handlerErrors)
	client.RecordHandlerError()
	newErrors := atomic.LoadUint64(&client.metrics.handlerErrors)

	if newErrors != initialErrors+1 {
		t.Fatalf("expected handler errors to increment from %d to %d, got %d", initialErrors, initialErrors+1, newErrors)
	}
}

func TestClientRecordHandlerErrorThreadSafe(t *testing.T) {
	client := &Client{}
	client.initQueues()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.RecordHandlerError()
		}()
	}

	wg.Wait()

	finalErrors := atomic.LoadUint64(&client.metrics.handlerErrors)
	if finalErrors != 100 {
		t.Fatalf("expected 100 handler errors, got %d", finalErrors)
	}
}

func TestClientRecordWriteError(t *testing.T) {
	client := &Client{}
	client.initQueues()

	initialErrors := atomic.LoadUint64(&client.metrics.writeErrors)
	client.recordWriteError()
	newErrors := atomic.LoadUint64(&client.metrics.writeErrors)

	if newErrors != initialErrors+1 {
		t.Fatalf("expected write errors to increment from %d to %d, got %d", initialErrors, initialErrors+1, newErrors)
	}
}

func TestClientMetricsSnapshot(t *testing.T) {
	client := &Client{}
	client.initQueues()

	client.metrics.queueMax = 42
	client.metrics.queueBlocks = 3
	atomic.StoreUint64(&client.metrics.handlerErrors, 5)
	atomic.StoreUint64(&client.metrics.writeErrors, 7)
	atomic.StoreUint64(&client.metrics.packets, 99)

	snapshot := client.MetricsSnapshot()

	if snapshot.QueueMax != 42 {
		t.Fatalf("expected queue max 42, got %d", snapshot.QueueMax)
	}
	if snapshot.QueueBlocks != 3 {
		t.Fatalf("expected queue blocks 3, got %d", snapshot.QueueBlocks)
	}
	if snapshot.HandlerErrors != 5 {
		t.Fatalf("expected handler errors 5, got %d", snapshot.HandlerErrors)
	}
	if snapshot.WriteErrors != 7 {
		t.Fatalf("expected write errors 7, got %d", snapshot.WriteErrors)
	}
	if snapshot.Packets != 99 {
		t.Fatalf("expected packets 99, got %d", snapshot.Packets)
	}
}

func TestClientAcquirePacketBuffer(t *testing.T) {
	client := &Client{}
	client.initQueues()

	buf1 := client.acquirePacketBuffer(100)
	if len(buf1) != 100 {
		t.Fatalf("expected buffer size 100, got %d", len(buf1))
	}

	if cap(buf1) < 100 {
		t.Fatalf("expected buffer capacity >= 100, got %d", cap(buf1))
	}

	buf2 := client.acquirePacketBuffer(200)
	if len(buf2) != 200 {
		t.Fatalf("expected buffer size 200, got %d", len(buf2))
	}
}

func TestClientAcquirePacketBufferFromPool(t *testing.T) {
	client := &Client{}
	client.initQueues()

	sizedBuffer := make([]byte, 500)
	client.packetPool <- sizedBuffer[:0]

	buf := client.acquirePacketBuffer(400)
	if cap(buf) != 500 {
		t.Fatalf("expected buffer from pool with capacity 500, got %d", cap(buf))
	}
}

func TestClientReleasePacketBuffer(t *testing.T) {
	client := &Client{}
	client.initQueues()

	buf := make([]byte, 100)
	client.releasePacketBuffer(buf)

	if client.packetQueue.Length() != 0 {
		t.Fatalf("expected queue to be empty, got length %d", client.packetQueue.Length())
	}
}

func TestClientReleasePacketBufferToPool(t *testing.T) {
	client := &Client{}
	client.initQueues()

	buf := make([]byte, 500)
	client.releasePacketBuffer(buf[:0])

	select {
	case <-client.packetPool:
	default:
		t.Fatalf("expected buffer to be released to pool")
	}
}

func TestClientReleasePacketBufferNil(t *testing.T) {
	client := &Client{}
	client.initQueues()

	client.releasePacketBuffer(nil)
}

func TestClientReleasePacketBufferPoolFull(t *testing.T) {
	client := &Client{}
	client.packetPool = make(chan []byte, 1)
	client.packetPool <- make([]byte, 100)[:0]

	buf1 := make([]byte, 100)
	client.releasePacketBuffer(buf1[:0])

	buf2 := make([]byte, 100)
	client.releasePacketBuffer(buf2[:0])
}

func TestClientDrainQueueLocked(t *testing.T) {
	client := &Client{}
	client.initQueues()

	for i := 0; i < 10; i++ {
		client.packetQueue.Add([]byte{byte(i)})
	}

	if client.packetQueue.Length() != 10 {
		t.Fatalf("expected queue length 10, got %d", client.packetQueue.Length())
	}

	client.drainQueueLocked()

	if client.packetQueue.Length() != 0 {
		t.Fatalf("expected queue to be drained, got length %d", client.packetQueue.Length())
	}
}

func TestClientStartDispatcher(t *testing.T) {
	client := &Client{}
	client.initQueues()

	dispatcherCalled := false
	client.Server = &Server{
		Dispatcher: func(pkt *[]byte, c *Client, size int) {
			dispatcherCalled = true
		},
	}

	client.StartDispatcher()
	client.Close()
	time.Sleep(10 * time.Millisecond)

	if !dispatcherCalled {
		t.Fatalf("expected dispatcher to be started")
	}
}

func TestClientMetricsThreadSafety(t *testing.T) {
	client := &Client{}
	client.initQueues()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.RecordHandlerError()
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.recordWriteError()
		}()
	}

	wg.Wait()

	handlerErrors := atomic.LoadUint64(&client.metrics.handlerErrors)
	writeErrors := atomic.LoadUint64(&client.metrics.writeErrors)

	if handlerErrors != 50 {
		t.Fatalf("expected 50 handler errors, got %d", handlerErrors)
	}
	if writeErrors != 50 {
		t.Fatalf("expected 50 write errors, got %d", writeErrors)
	}
}

func TestClientQueueMetrics(t *testing.T) {
	client := &Client{}
	client.initQueues()

	if client.metrics.queueMax != 0 {
		t.Fatalf("expected initial queue max 0, got %d", client.metrics.queueMax)
	}

	for i := 0; i < 100; i++ {
		client.packetQueue.Add([]byte{byte(i)})
	}

	client.queueMu.Lock()
	depth := client.packetQueue.Length()
	if depth > client.metrics.queueMax {
		client.metrics.queueMax = depth
	}
	client.queueMu.Unlock()

	if client.metrics.queueMax != 100 {
		t.Fatalf("expected queue max 100, got %d", client.metrics.queueMax)
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type mockConn struct {
	closed bool
}

func mockToNetConn(m *mockConn) *net.Conn {
	var nc net.Conn = m
	return &nc
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *mockConn) SetDeadline(t time.Time) error     { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestClientCloseWithConnection(t *testing.T) {
	mockConn := &mockConn{}
	client := &Client{
		Connection: mockToNetConn(mockConn),
	}
	client.initQueues()

	client.Close()

	if !mockConn.closed {
		t.Fatalf("expected connection to be closed")
	}
}
