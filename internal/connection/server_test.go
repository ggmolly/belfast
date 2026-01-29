package connection

import (
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/smallnest/ringbuffer"
	"google.golang.org/protobuf/proto"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/region"
)

type testConn struct {
	closed     bool
	remoteAddr net.Addr
}

func (m *testConn) Close() error {
	m.closed = true
	return nil
}

func (m *testConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m *testConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (m *testConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *testConn) RemoteAddr() net.Addr {
	if m.remoteAddr != nil {
		return m.remoteAddr
	}
	return &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 54321}
}

func (m *testConn) SetDeadline(t time.Time) error     { return nil }
func (m *testConn) SetReadDeadline(t time.Time) error { return nil }
func (m *testConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func testToNetConn(m *testConn) net.Conn {
	return m
}

func initServerTest(t *testing.T) (*Server, func()) {
	t.Helper()

	cleanup := func() {
		if BelfastInstance != nil {
			BelfastInstance = nil
		}
	}

	t.Cleanup(cleanup)

	server := NewServer("127.0.0.1", 8080, func(pkt *[]byte, c *Client, size int) {})

	return server, cleanup
}

func TestServerNewServer(t *testing.T) {
	server, _ := initServerTest(t)

	if server.BindAddress != "127.0.0.1" {
		t.Fatalf("expected bind address 127.0.0.1, got %s", server.BindAddress)
	}
	if server.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", server.Port)
	}
	if server.Dispatcher == nil {
		t.Fatalf("expected dispatcher to be set")
	}
	if server.clients == nil {
		t.Fatalf("expected clients map to be initialized")
	}
	if server.rooms == nil {
		t.Fatalf("expected rooms map to be initialized")
	}
	if server.StartTime.IsZero() {
		t.Fatalf("expected start time to be set")
	}
	if !server.acceptingConnections.Load() {
		t.Fatalf("expected accepting connections to be true")
	}
	if server.maintenanceEnabled != 0 {
		t.Fatalf("expected maintenance disabled initially, got %d", server.maintenanceEnabled)
	}
}

func TestServerAddClient(t *testing.T) {
	server, _ := initServerTest(t)

	client := &Client{
		IP:   net.ParseIP("192.168.1.100"),
		Port: 12345,
		Hash: 12345,
	}

	server.AddClient(client)

	server.clientsMutex.RLock()
	defer server.clientsMutex.RUnlock()

	if len(server.clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(server.clients))
	}

	storedClient, ok := server.clients[client.Hash]
	if !ok {
		t.Fatalf("expected client to be in clients map")
	}
	if storedClient != client {
		t.Fatalf("expected stored client to match added client")
	}
}

func TestServerAddMultipleClients(t *testing.T) {
	server, _ := initServerTest(t)

	clients := []*Client{
		{Hash: 1001},
		{Hash: 1002},
		{Hash: 1003},
	}

	for _, client := range clients {
		server.AddClient(client)
	}

	server.clientsMutex.RLock()
	defer server.clientsMutex.RUnlock()

	if len(server.clients) != 3 {
		t.Fatalf("expected 3 clients, got %d", len(server.clients))
	}

	for _, client := range clients {
		if _, ok := server.clients[client.Hash]; !ok {
			t.Fatalf("expected client %d to be stored", client.Hash)
		}
	}
}

func TestServerRemoveClient(t *testing.T) {
	server, _ := initServerTest(t)

	client := &Client{
		Hash:       99999,
		IP:         net.ParseIP("10.0.0.1"),
		Port:       9999,
		Connection: mockToNetConn(&mockConn{}),
	}

	server.clients[client.Hash] = client

	server.RemoveClient(client)

	server.clientsMutex.RLock()
	defer server.clientsMutex.RUnlock()

	if _, ok := server.clients[client.Hash]; ok {
		t.Fatalf("expected client to be removed from clients map")
	}
}

func TestServerRemoveClientClosesConnection(t *testing.T) {
	mockConn := &mockConn{}
	client := &Client{
		Hash:       99999,
		IP:         net.ParseIP("10.0.0.1"),
		Port:       9999,
		Connection: mockToNetConn(mockConn),
	}

	server, _ := initServerTest(t)
	server.clients[client.Hash] = client

	server.RemoveClient(client)

	if !mockConn.closed {
		t.Fatalf("expected connection to be closed")
	}
	if !client.IsClosed() {
		t.Fatalf("expected client close flag to be set")
	}
}

func TestServerClientCount(t *testing.T) {
	server, _ := initServerTest(t)

	if server.ClientCount() != 0 {
		t.Fatalf("expected 0 clients initially, got %d", server.ClientCount())
	}

	for i := 0; i < 5; i++ {
		server.clients[uint32(1000+i)] = &Client{Hash: uint32(1000 + i)}
	}

	if server.ClientCount() != 5 {
		t.Fatalf("expected 5 clients, got %d", server.ClientCount())
	}
}

func TestServerClientCountThreadSafe(t *testing.T) {
	server, _ := initServerTest(t)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(hash uint32) {
			defer wg.Done()
			server.AddClient(&Client{Hash: hash})
		}(uint32(i))
	}

	wg.Wait()
	count := server.ClientCount()
	if count != 100 {
		t.Fatalf("expected 100 clients, got %d", count)
	}
}

func TestServerListClients(t *testing.T) {
	server, _ := initServerTest(t)

	clients := []*Client{
		{Hash: 1, IP: net.ParseIP("192.168.1.1")},
		{Hash: 2, IP: net.ParseIP("192.168.1.2")},
		{Hash: 3, IP: net.ParseIP("192.168.1.3")},
	}

	for _, client := range clients {
		server.clients[client.Hash] = client
	}

	listed := server.ListClients()

	if len(listed) != 3 {
		t.Fatalf("expected 3 clients listed, got %d", len(listed))
	}

	hashes := make(map[uint32]bool)
	for _, client := range listed {
		hashes[client.Hash] = true
	}

	for hash := range hashes {
		if !hashes[hash] {
			t.Fatalf("expected client hash %d to be listed", hash)
		}
	}
}

func TestServerFindClient(t *testing.T) {
	server, _ := initServerTest(t)

	targetClient := &Client{
		Hash: 55555,
		IP:   net.ParseIP("172.16.0.100"),
		Port: 1234,
	}

	server.clients[targetClient.Hash] = targetClient

	found, ok := server.FindClient(55555)
	if !ok {
		t.Fatalf("expected client to be found")
	}
	if found != targetClient {
		t.Fatalf("expected found client to match target")
	}

	_, ok = server.FindClient(99999)
	if ok {
		t.Fatalf("expected client not to be found")
	}
}

func TestServerFindClientByCommander(t *testing.T) {
	server, _ := initServerTest(t)

	client := &Client{
		Hash: 77777,
		Commander: &orm.Commander{
			CommanderID: 12345,
		},
	}

	server.clients[client.Hash] = client

	found, ok := server.FindClientByCommander(12345)
	if !ok {
		t.Fatalf("expected commander to be found")
	}
	if found != client {
		t.Fatalf("expected found client to match")
	}

	_, ok = server.FindClientByCommander(99999)
	if ok {
		t.Fatalf("expected commander not to be found")
	}
}

func TestServerFindClientByCommanderNilCommander(t *testing.T) {
	server, _ := initServerTest(t)

	client := &Client{Hash: 88888}
	server.clients[client.Hash] = client

	_, ok := server.FindClientByCommander(12345)
	if ok {
		t.Fatalf("expected commander not to be found when client has nil commander")
	}
}

func TestServerSetAcceptingConnections(t *testing.T) {
	server, _ := initServerTest(t)

	if !server.acceptingConnections.Load() {
		t.Fatalf("expected accepting connections initially")
	}

	server.SetAcceptingConnections(true)

	if !server.acceptingConnections.Load() {
		t.Fatalf("expected accepting connections to be true after SetAcceptingConnections(true)")
	}

	server.SetAcceptingConnections(false)

	if server.acceptingConnections.Load() {
		t.Fatalf("expected accepting connections to be false after SetAcceptingConnections(false)")
	}
}

func TestServerIsAcceptingConnections(t *testing.T) {
	server, _ := initServerTest(t)

	if !server.IsAcceptingConnections() {
		t.Fatalf("expected accepting connections initially")
	}

	server.acceptingConnections.Store(false)

	if server.IsAcceptingConnections() {
		t.Fatalf("expected not accepting connections after Store(false)")
	}
}

func TestServerMaintenanceEnabled(t *testing.T) {
	server, _ := initServerTest(t)

	server.SetMaintenance(true)

	if !server.MaintenanceEnabled() {
		t.Fatalf("expected maintenance to be enabled")
	}

	server.SetMaintenance(false)

	if server.MaintenanceEnabled() {
		t.Fatalf("expected maintenance to be disabled")
	}
}

func TestServerGetClient(t *testing.T) {
	server, _ := initServerTest(t)

	mockConn := &mockConn{
		remoteAddr: &net.TCPAddr{
			IP:   net.ParseIP("10.0.0.1"),
			Port: 54321,
		},
	}
	var conn net.Conn = mockConn

	client, err := server.GetClient(&conn)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.IP == nil {
		t.Fatalf("expected client IP to be set")
	}
	if client.Port != 54321 {
		t.Fatalf("expected client port 54321, got %d", client.Port)
	}
	if client.Server != server {
		t.Fatalf("expected client server to be set")
	}
	if client.ConnectedAt.IsZero() {
		t.Fatalf("expected connected at to be set")
	}
	if client.Hash == 0 {
		t.Fatalf("expected client hash to be calculated")
	}
	if client.packetQueue == nil {
		t.Fatalf("expected packet queue to be initialized")
	}
}

func TestServerGetClientHashCalculation(t *testing.T) {
	server, _ := initServerTest(t)

	tests := []struct {
		addr     string
		port     int
		expected bool
	}{
		{"127.0.0.1", 8080, false},
		{"10.0.0.1", 9000, true},
		{"172.16.0.1", 5000, true},
		{"8.8.8.8", 80, false},
		{"1.1.1.1", 443, false},
	}

	for _, tt := range tests {
		testConn := &testConn{
			remoteAddr: &net.TCPAddr{
				IP:   net.ParseIP(tt.addr),
				Port: tt.port,
			},
		}
		var conn net.Conn = testToNetConn(testConn)

		client, _ := server.GetClient(&conn)

		isPrivate := client.IP.IsPrivate()
		if isPrivate != tt.expected {
			t.Fatalf("expected IP %s to be private=%v, got %v", tt.addr, tt.expected, isPrivate)
		}
	}
}

func TestServerRegionSet(t *testing.T) {
	region.SetCurrent("EN")
	server, _ := initServerTest(t)

	if server.Region != "EN" {
		t.Fatalf("expected region EN, got %s", server.Region)
	}
}

func TestServerStartTimeSet(t *testing.T) {
	before := time.Now().UTC()
	time.Sleep(10 * time.Millisecond)

	server, _ := initServerTest(t)

	if server.StartTime.Before(before) {
		t.Fatalf("expected start time to be after %v", before)
	}
}

func TestServerClientsMapThreadSafety(t *testing.T) {
	server, _ := initServerTest(t)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(hash uint32) {
			defer wg.Done()
			server.AddClient(&Client{Hash: hash})
		}(uint32(i))
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(hash uint32) {
			defer wg.Done()
			server.clientsMutex.RLock()
			_ = server.clients[hash]
			server.clientsMutex.RUnlock()
		}(uint32(i))
	}

	wg.Wait()

	if len(server.clients) != 50 {
		t.Fatalf("expected 50 clients, got %d", len(server.clients))
	}
}

func TestServerRoomsMapInitialized(t *testing.T) {
	server, _ := initServerTest(t)

	if server.rooms == nil {
		t.Fatalf("expected rooms map to be initialized")
	}
}

func TestServerDisconnectCommander(t *testing.T) {
	server, _ := initServerTest(t)

	commander := &orm.Commander{CommanderID: 99988}
	client := &Client{
		Hash:       88888,
		Commander:  commander,
		Connection: mockToNetConn(&mockConn{}),
	}

	server.clients[client.Hash] = client

	kicked := server.DisconnectCommander(99988, 1, nil)

	if !kicked {
		t.Fatalf("expected commander to be kicked")
	}

	if _, ok := server.clients[client.Hash]; ok {
		t.Fatalf("expected client to be removed from clients map")
	}
}

func TestServerDisconnectCommanderNotFound(t *testing.T) {
	server, _ := initServerTest(t)

	kicked := server.DisconnectCommander(99999, 1, nil)

	if kicked {
		t.Fatalf("expected commander not to be kicked")
	}
}

func TestServerDisconnectCommanderExcludeClient(t *testing.T) {
	server, _ := initServerTest(t)

	commander := &orm.Commander{CommanderID: 99988}
	client := &Client{
		Hash:       88888,
		Commander:  commander,
		Connection: mockToNetConn(&mockConn{}),
	}

	server.clients[client.Hash] = client

	kicked := server.DisconnectCommander(99988, 1, client)

	if kicked {
		t.Fatalf("expected commander not to be kicked when excluded")
	}

	if _, ok := server.clients[client.Hash]; !ok {
		t.Fatalf("expected client to still be in clients map when excluded")
	}
}

func TestDisconnectAllRemovesClients(t *testing.T) {
	server, _ := initServerTest(t)

	for i := 0; i < 2; i++ {
		client := &Client{
			Hash:       uint32(i + 1),
			IP:         net.ParseIP("10.0.0.1"),
			Port:       1000 + i,
			Connection: mockToNetConn(&mockConn{}),
		}
		client.initQueues()
		server.clients[client.Hash] = client
	}

	server.DisconnectAll(consts.DR_SERVER_MAINTENANCE)

	if len(server.clients) != 0 {
		t.Fatalf("expected clients map to be empty")
	}
}

func TestRoomJoinLeaveChange(t *testing.T) {
	server, _ := initServerTest(t)

	clientA := &Client{Hash: 1}
	clientB := &Client{Hash: 2}

	server.JoinRoom(10, clientA)
	server.JoinRoom(10, clientB)

	if len(server.rooms[10]) != 2 {
		t.Fatalf("expected 2 clients in room")
	}

	server.LeaveRoom(10, clientA)
	if len(server.rooms[10]) != 1 {
		t.Fatalf("expected 1 client after leave")
	}
	if server.rooms[10][0] != clientB {
		t.Fatalf("expected remaining client to be clientB")
	}

	server.ChangeRoom(10, 20, clientB)
	if len(server.rooms[10]) != 0 {
		t.Fatalf("expected old room to be empty")
	}
	if len(server.rooms[20]) != 1 {
		t.Fatalf("expected new room to have client")
	}
	if server.rooms[20][0] != clientB {
		t.Fatalf("expected clientB in new room")
	}
}

func TestSendMessagePayload(t *testing.T) {
	server, _ := initServerTest(t)
	roomID := uint32(99)

	sender := &Client{Commander: &orm.Commander{CommanderID: 1234, Name: "Alice", Level: 11}}
	recipient := &Client{}
	recipient.initQueues()

	server.JoinRoom(roomID, recipient)

	msg := orm.Message{RoomID: roomID, Content: "hello"}
	server.SendMessage(sender, msg)

	data := recipient.Buffer.Bytes()
	if len(data) == 0 {
		t.Fatalf("expected recipient buffer to be populated")
	}
	packetID := int(data[3])<<8 | int(data[4])
	if packetID != 50101 {
		t.Fatalf("expected packet id 50101, got %d", packetID)
	}
	payload := data[7:]
	var parsed protobuf.SC_50101
	if err := proto.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if parsed.GetPlayer().GetId() != 1234 {
		t.Fatalf("expected player id 1234, got %d", parsed.GetPlayer().GetId())
	}
	if parsed.GetPlayer().GetName() != "Alice" {
		t.Fatalf("expected name Alice, got %s", parsed.GetPlayer().GetName())
	}
	if parsed.GetPlayer().GetLv() != 11 {
		t.Fatalf("expected level 11, got %d", parsed.GetPlayer().GetLv())
	}
	if parsed.GetType() != orm.MSG_TYPE_NORMAL {
		t.Fatalf("expected message type normal")
	}
	if parsed.GetContent() != "hello" {
		t.Fatalf("expected content hello, got %s", parsed.GetContent())
	}
}

func TestGeneratePacketHeader(t *testing.T) {
	payload := []byte{0x01, 0x02}
	header := GeneratePacketHeader(0x1234, &payload, 0x0001)

	expected := []byte{0x00, 0x07, 0x00, 0x12, 0x34, 0x00, 0x01}
	if !equalBytes(header, expected) {
		t.Fatalf("expected header %v, got %v", expected, header)
	}
}

func TestInjectPacketHeader(t *testing.T) {
	payload := []byte{0xAA, 0xBB}
	InjectPacketHeader(0x1234, &payload, 0x0001)

	expectedHeader := []byte{0x00, 0x07, 0x00, 0x12, 0x34, 0x00, 0x01}
	if len(payload) != len(expectedHeader)+2 {
		t.Fatalf("expected payload length %d, got %d", len(expectedHeader)+2, len(payload))
	}
	if !equalBytes(payload[:len(expectedHeader)], expectedHeader) {
		t.Fatalf("expected header %v, got %v", expectedHeader, payload[:len(expectedHeader)])
	}
	if !equalBytes(payload[len(expectedHeader):], []byte{0xAA, 0xBB}) {
		t.Fatalf("expected payload to be preserved")
	}
}

func TestReadPacketSizeValid(t *testing.T) {
	ring := ringbuffer.New(8).SetBlocking(true)
	_, _ = ring.Write([]byte{0x00, 0x05})
	ring.CloseWriter()

	header, size, err := readPacketSize(ring)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if header != [2]byte{0x00, 0x05} {
		t.Fatalf("expected header to match")
	}
	if size != 7 {
		t.Fatalf("expected size 7, got %d", size)
	}
}

func TestReadPacketSizeInvalid(t *testing.T) {
	ring := ringbuffer.New(8).SetBlocking(true)
	_, _ = ring.Write([]byte{0x00, 0x04})
	ring.CloseWriter()

	_, _, err := readPacketSize(ring)
	if err == nil {
		t.Fatalf("expected error for invalid size")
	}
}

func TestReadPacketBodyEmpty(t *testing.T) {
	ring := ringbuffer.New(8).SetBlocking(true)
	ring.CloseWriter()

	if err := readPacketBody(ring, nil); err != nil {
		t.Fatalf("expected nil error for empty packet, got %v", err)
	}
}

func TestReadPacketBodyEOF(t *testing.T) {
	ring := ringbuffer.New(8).SetBlocking(true)
	ring.CloseWriter()

	err := readPacketBody(ring, make([]byte, 2))
	if err == nil || (!errors.Is(err, io.EOF) && !errors.Is(err, io.ErrClosedPipe)) {
		t.Fatalf("expected EOF or closed pipe, got %v", err)
	}
}
