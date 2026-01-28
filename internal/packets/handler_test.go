package packets

import (
	"net"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/region"
)

type mockConn struct{}

func (m *mockConn) Read(b []byte) (int, error)         { return 0, nil }
func (m *mockConn) Write(b []byte) (int, error)        { return len(b), nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func newTestClient() *connection.Client {
	var conn net.Conn = &mockConn{}
	return &connection.Client{Connection: &conn}
}

func initPacketTests(t *testing.T) {
	t.Helper()
	PacketDecisionFn = make(map[int][]PacketHandler)
}

func TestRegisterPacketHandler(t *testing.T) {
	initPacketTests(t)
	packetID := 99999

	handlerCalled := false
	handlers := []PacketHandler{
		func(pkt *[]byte, c *connection.Client) (int, int, error) {
			handlerCalled = true
			return 0, 0, nil
		},
	}

	RegisterPacketHandler(packetID, handlers)

	stored, ok := PacketDecisionFn[packetID]
	if !ok {
		t.Fatalf("expected handler to be registered")
	}
	if len(stored) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(stored))
	}
	stored[0](nil, nil)

	if !handlerCalled {
		t.Fatalf("expected handler to be callable")
	}
}

func TestRegisterPacketHandlerMultiple(t *testing.T) {
	initPacketTests(t)
	packetID := 88888

	handlers := []PacketHandler{
		func(pkt *[]byte, c *connection.Client) (int, int, error) { return 0, 0, nil },
		func(pkt *[]byte, c *connection.Client) (int, int, error) { return 0, 0, nil },
		func(pkt *[]byte, c *connection.Client) (int, int, error) { return 0, 0, nil },
	}

	RegisterPacketHandler(packetID, handlers)

	stored, ok := PacketDecisionFn[packetID]
	if !ok {
		t.Fatalf("expected handlers to be registered")
	}
	if len(stored) != 3 {
		t.Fatalf("expected 3 handlers, got %d", len(stored))
	}
}

func TestRegisterPacketHandlerOverwrites(t *testing.T) {
	initPacketTests(t)
	packetID := 77777

	original := []PacketHandler{
		func(pkt *[]byte, c *connection.Client) (int, int, error) { return 1, 1, nil },
	}

	RegisterPacketHandler(packetID, original)

	replacement := []PacketHandler{
		func(pkt *[]byte, c *connection.Client) (int, int, error) { return 2, 2, nil },
	}

	RegisterPacketHandler(packetID, replacement)

	stored, ok := PacketDecisionFn[packetID]
	if !ok {
		t.Fatalf("expected handler to be registered")
	}

	result, _, _ := stored[0](nil, nil)
	if result != 2 {
		t.Fatalf("expected handler to be overwritten")
	}
}

func TestRegisterLocalizedPacketHandlerCN(t *testing.T) {
	initPacketTests(t)
	packetID := 66666
	region.SetCurrent("CN")

	cnHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 1, 0, nil }
	otherHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 2, 0, nil }

	localized := LocalizedHandler{
		CN:      &[]PacketHandler{cnHandler},
		Default: &[]PacketHandler{otherHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored, ok := PacketDecisionFn[packetID]
	if !ok {
		t.Fatalf("expected handler to be registered")
	}

	result, _, _ := stored[0](nil, nil)
	if result != 1 {
		t.Fatalf("expected CN handler (1), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerEN(t *testing.T) {
	initPacketTests(t)
	packetID := 65555
	region.SetCurrent("EN")

	enHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 3, 0, nil }
	otherHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 4, 0, nil }

	localized := LocalizedHandler{
		EN:      &[]PacketHandler{enHandler},
		Default: &[]PacketHandler{otherHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored, ok := PacketDecisionFn[packetID]
	if !ok {
		t.Fatalf("expected handler to be registered")
	}

	result, _, _ := stored[0](nil, nil)
	if result != 3 {
		t.Fatalf("expected EN handler (3), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerDefault(t *testing.T) {
	initPacketTests(t)
	packetID := 64444
	region.SetCurrent("EN")

	defaultHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 5, 0, nil }

	localized := LocalizedHandler{
		Default: &[]PacketHandler{defaultHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored, ok := PacketDecisionFn[packetID]
	if !ok {
		t.Fatalf("expected handler to be registered")
	}

	result, _, _ := stored[0](nil, nil)
	if result != 5 {
		t.Fatalf("expected default handler (5), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerNilRegion(t *testing.T) {
	initPacketTests(t)
	packetID := 63333
	region.SetCurrent("EN")

	cnHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 6, 0, nil }

	localized := LocalizedHandler{
		CN: &[]PacketHandler{cnHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	_, ok := PacketDecisionFn[packetID]
	if ok {
		t.Fatalf("expected no handler for unregistered region")
	}
}

func TestGetPacketId(t *testing.T) {
	buffer := []byte{
		0x12, 0x34,
		0x00,
		0x56, 0x78,
		0x00, 0x00,
	}

	packetID := GetPacketId(0, &buffer)

	if packetID != 0x5678 {
		t.Fatalf("expected packet ID 0x5678, got 0x%04x", packetID)
	}
}

func TestGetPacketIdWithOffset(t *testing.T) {
	buffer := []byte{
		0xAA, 0xBB,
		0xCC, 0xDD,
		0x00, 0xEE, 0x78, 0x9A,
		0x12, 0x34,
	}

	packetID := GetPacketId(3, &buffer)

	if packetID != 0x789A {
		t.Fatalf("expected packet ID 0x789A, got 0x%04x", packetID)
	}
}

func TestGetPacketSize(t *testing.T) {
	buffer := []byte{
		0x01, 0x00,
		0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
	}

	size := GetPacketSize(0, &buffer)

	if size != 0x0100 {
		t.Fatalf("expected packet size 0x0100, got 0x%04x", size)
	}
}

func TestGetPacketSizeWithOffset(t *testing.T) {
	buffer := []byte{
		0xFF, 0xFF,
		0xFF, 0xFE,
		0x00, 0xFF, 0xFE, 0xFD,
		0x12, 0x34,
	}

	size := GetPacketSize(2, &buffer)

	if size != 0xFFFE {
		t.Fatalf("expected packet size 0xFFFE, got 0x%04x", size)
	}
}

func TestGetPacketIndex(t *testing.T) {
	buffer := []byte{
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0xAB, 0xCD,
	}

	index := GetPacketIndex(0, &buffer)

	if index != 0xABCD {
		t.Fatalf("expected packet index 0xABCD, got 0x%04x", index)
	}
}

func TestGetPacketIndexWithOffset(t *testing.T) {
	buffer := []byte{
		0x11, 0x22,
		0x33, 0x44,
		0x00, 0x55, 0x66, 0x77,
		0x55, 0x66,
	}

	index := GetPacketIndex(3, &buffer)

	if index != 0x5566 {
		t.Fatalf("expected packet index 0x5566, got 0x%04x", index)
	}
}

func TestPacketHeaderConstants(t *testing.T) {
	if HEADER_SIZE != 7 {
		t.Fatalf("expected HEADER_SIZE 7, got %d", HEADER_SIZE)
	}
}

func TestDispatchWithHandler(t *testing.T) {
	initPacketTests(t)

	dispatchCalled := false
	var dispatchedPacketID int
	var dispatchedClient *connection.Client

	handlers := []PacketHandler{
		func(pkt *[]byte, c *connection.Client) (int, int, error) {
			dispatchCalled = true
			dispatchedPacketID = 12345
			dispatchedClient = c
			return 5, 10, nil
		},
	}

	PacketDecisionFn[12345] = handlers

	client := newTestClient()

	buffer := []byte{
		0x30, 0x39,
		0x00,
		0x30, 0x39,
		0x00, 0x00,
		0x01, 0x02, 0x03, 0x04, 0x05,
	}

	Dispatch(&buffer, client, len(buffer))

	if !dispatchCalled {
		t.Fatalf("expected dispatch to call handler")
	}

	if dispatchedPacketID != 12345 {
		t.Fatalf("expected packet ID 12345, got %d", dispatchedPacketID)
	}

	if dispatchedClient != client {
		t.Fatalf("expected correct client to be passed")
	}
}

func TestDispatchWithoutHandler(t *testing.T) {
	initPacketTests(t)

	buffer := []byte{
		0x00, 0x10,
		0x00,
		0x27, 0x18,
		0x00, 0x00,
		0x01, 0x02, 0x03, 0x04, 0x05,
	}

	client := newTestClient()

	Dispatch(&buffer, client, len(buffer))
}

func TestDispatchMultiplePackets(t *testing.T) {
	initPacketTests(t)

	callCount := 0
	handlers := []PacketHandler{
		func(pkt *[]byte, c *connection.Client) (int, int, error) {
			callCount++
			return 0, 0, nil
		},
	}

	PacketDecisionFn[54321] = handlers

	buffer := []byte{
		0x00, 0x0A,
		0x00,
		0xD4, 0x31,
		0x00, 0x00,
		0x01, 0x02, 0x03, 0x04, 0x05,

		0x00, 0x0A,
		0x00,
		0xD4, 0x31,
		0x00, 0x00,
		0x06, 0x07, 0x08, 0x09, 0x0A,
	}

	client := newTestClient()

	Dispatch(&buffer, client, len(buffer))

	if callCount != 2 {
		t.Fatalf("expected 2 packets to be dispatched, got %d", callCount)
	}
}
