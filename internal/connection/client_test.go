package connection

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
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
	client.queueLimit = 2

	largePacket := make([]byte, 100)

	for i := 0; i < client.queueLimit; i++ {
		if err := client.EnqueuePacket(largePacket); err != nil {
			t.Fatalf("expected enqueue to succeed, got %v", err)
		}
	}

	initialQueueBlocks := client.metrics.queueBlocks
	errCh := make(chan error, 1)
	go func() {
		errCh <- client.EnqueuePacket(largePacket)
	}()

	select {
	case err := <-errCh:
		t.Fatalf("expected enqueue to block, got %v", err)
	case <-time.After(10 * time.Millisecond):
	}

	client.Close()
	if err := <-errCh; !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed after close, got %v", err)
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

	dispatcherCalled := make(chan struct{}, 1)
	client.Server = &Server{
		Dispatcher: func(pkt *[]byte, c *Client, size int) {
			select {
			case dispatcherCalled <- struct{}{}:
			default:
			}
		},
	}

	client.StartDispatcher()
	if err := client.EnqueuePacket([]byte{1, 2, 3}); err != nil {
		t.Fatalf("enqueue packet: %v", err)
	}
	select {
	case <-dispatcherCalled:
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("expected dispatcher to be started")
	}
	client.Close()
}

func TestClientDispatchLoopProcessesPackets(t *testing.T) {
	client := &Client{}
	client.initQueues()

	processed := make(chan []byte, 2)
	client.Server = &Server{
		Dispatcher: func(pkt *[]byte, c *Client, size int) {
			copyBuf := make([]byte, len(*pkt))
			copy(copyBuf, *pkt)
			processed <- copyBuf
		},
	}

	done := make(chan struct{})
	go func() {
		client.dispatchLoop()
		close(done)
	}()

	if err := client.EnqueuePacket([]byte{1, 2}); err != nil {
		t.Fatalf("enqueue packet: %v", err)
	}
	if err := client.EnqueuePacket([]byte{3, 4, 5}); err != nil {
		t.Fatalf("enqueue packet: %v", err)
	}

	for i := 0; i < 2; i++ {
		select {
		case <-processed:
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("expected dispatcher to process packet %d", i+1)
		}
	}

	client.Close()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected dispatch loop to exit after close")
	}

	if client.MetricsSnapshot().Packets != 2 {
		t.Fatalf("expected packets metric 2, got %d", client.MetricsSnapshot().Packets)
	}
}

func TestClientDispatchLoopDrainsOnClose(t *testing.T) {
	client := &Client{}
	client.initQueues()

	client.queueMu.Lock()
	client.packetQueue.Add([]byte{1})
	client.packetQueue.Add([]byte{2})
	client.closed = true
	client.queueMu.Unlock()

	done := make(chan struct{})
	go func() {
		client.dispatchLoop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected dispatch loop to exit")
	}

	if client.packetQueue.Length() != 0 {
		t.Fatalf("expected queue to be drained, got %d", client.packetQueue.Length())
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
	closed     bool
	remoteAddr net.Addr
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
	if m.remoteAddr != nil {
		return m.remoteAddr
	}
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

type writeErrConn struct {
	closed bool
}

func (c *writeErrConn) Close() error {
	c.closed = true
	return nil
}

func (c *writeErrConn) Write(b []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

func (c *writeErrConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (c *writeErrConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (c *writeErrConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (c *writeErrConn) SetDeadline(t time.Time) error     { return nil }
func (c *writeErrConn) SetReadDeadline(t time.Time) error { return nil }
func (c *writeErrConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestClientFlushErrorClosesClient(t *testing.T) {
	errConn := &writeErrConn{}
	var conn net.Conn = errConn
	client := &Client{
		Connection: &conn,
	}
	client.initQueues()
	client.Buffer.Write([]byte{1, 2, 3})

	if err := client.Flush(); err == nil {
		t.Fatalf("expected flush error")
	}
	if !client.IsClosed() {
		t.Fatalf("expected client to be closed after flush error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected buffer to be reset after flush error")
	}
	if client.MetricsSnapshot().WriteErrors != 1 {
		t.Fatalf("expected write error metric to increment")
	}
}

func TestClientCreateCommander(t *testing.T) {
	withTestDB(t, &orm.YostarusMap{}, &orm.Commander{}, &orm.OwnedShip{}, &orm.CommanderItem{}, &orm.OwnedResource{}, &orm.Fleet{})

	client := &Client{}
	accountID, err := client.CreateCommander(321)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var mapping orm.YostarusMap
	if err := orm.GormDB.Where("arg2 = ?", 321).First(&mapping).Error; err != nil {
		t.Fatalf("expected yostarus map entry: %v", err)
	}
	if mapping.AccountID != accountID {
		t.Fatalf("expected mapping account %d, got %d", accountID, mapping.AccountID)
	}

	var commander orm.Commander
	if err := orm.GormDB.Where("account_id = ?", accountID).First(&commander).Error; err != nil {
		t.Fatalf("expected commander: %v", err)
	}

	var ships []orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ?", accountID).Find(&ships).Error; err != nil {
		t.Fatalf("expected owned ships: %v", err)
	}
	if len(ships) != 2 {
		t.Fatalf("expected 2 owned ships, got %d", len(ships))
	}
	var hasSecretary bool
	var hasLongIsland bool
	for _, ship := range ships {
		if ship.ShipID == 202124 && ship.IsSecretary {
			hasSecretary = true
		}
		if ship.ShipID == 106011 {
			hasLongIsland = true
		}
	}
	if !hasSecretary || !hasLongIsland {
		t.Fatalf("expected Belfast secretary ship and Long Island")
	}

	var fleets []orm.Fleet
	if err := orm.GormDB.Where("commander_id = ?", accountID).Find(&fleets).Error; err != nil {
		t.Fatalf("expected fleets: %v", err)
	}
	if len(fleets) != 1 {
		t.Fatalf("expected 1 fleet, got %d", len(fleets))
	}
	if fleets[0].GameID != 1 {
		t.Fatalf("expected fleet game_id 1, got %d", fleets[0].GameID)
	}
	var longIslandOwnedID uint32
	for _, ship := range ships {
		if ship.ShipID == 106011 {
			longIslandOwnedID = ship.ID
			break
		}
	}
	if longIslandOwnedID == 0 {
		t.Fatalf("expected Long Island owned ship id")
	}
	var inFleet bool
	for _, id := range fleets[0].ShipList {
		if uint32(id) == longIslandOwnedID {
			inFleet = true
			break
		}
	}
	if !inFleet {
		t.Fatalf("expected Long Island to be in fleet 1")
	}

	var items []orm.CommanderItem
	if err := orm.GormDB.Where("commander_id = ?", accountID).Find(&items).Error; err != nil {
		t.Fatalf("expected commander items: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	var resources []orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ?", accountID).Find(&resources).Error; err != nil {
		t.Fatalf("expected resources: %v", err)
	}
	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources))
	}
}

func TestClientCreateCommanderWithStarter(t *testing.T) {
	withTestDB(t, &orm.YostarusMap{}, &orm.Commander{}, &orm.OwnedShip{}, &orm.CommanderItem{}, &orm.OwnedResource{}, &orm.Fleet{})

	client := &Client{}
	accountID, err := client.CreateCommanderWithStarter(654, "Test", 101)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var mapping orm.YostarusMap
	if err := orm.GormDB.Where("arg2 = ?", 654).First(&mapping).Error; err != nil {
		t.Fatalf("expected yostarus map entry: %v", err)
	}
	if mapping.AccountID != accountID {
		t.Fatalf("expected mapping account %d, got %d", accountID, mapping.AccountID)
	}

	var ships []orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ?", accountID).Find(&ships).Error; err != nil {
		t.Fatalf("expected owned ships: %v", err)
	}
	if len(ships) != 3 {
		t.Fatalf("expected 3 owned ships, got %d", len(ships))
	}
	var hasStarter bool
	var hasSecretary bool
	var hasLongIsland bool
	for _, ship := range ships {
		if ship.ShipID == 101 {
			hasStarter = true
		}
		if ship.ShipID == 202124 && ship.IsSecretary {
			hasSecretary = true
		}
		if ship.ShipID == 106011 {
			hasLongIsland = true
		}
	}
	if !hasStarter || !hasSecretary || !hasLongIsland {
		t.Fatalf("expected starter, secretary, and Long Island ships")
	}

	var fleets []orm.Fleet
	if err := orm.GormDB.Where("commander_id = ?", accountID).Find(&fleets).Error; err != nil {
		t.Fatalf("expected fleets: %v", err)
	}
	if len(fleets) != 1 {
		t.Fatalf("expected 1 fleet, got %d", len(fleets))
	}
	if fleets[0].GameID != 1 {
		t.Fatalf("expected fleet game_id 1, got %d", fleets[0].GameID)
	}
	var longIslandOwnedID uint32
	for _, ship := range ships {
		if ship.ShipID == 106011 {
			longIslandOwnedID = ship.ID
			break
		}
	}
	if longIslandOwnedID == 0 {
		t.Fatalf("expected Long Island owned ship id")
	}
	var inFleet bool
	for _, id := range fleets[0].ShipList {
		if uint32(id) == longIslandOwnedID {
			inFleet = true
			break
		}
	}
	if !inFleet {
		t.Fatalf("expected Long Island to be in fleet 1")
	}
}
