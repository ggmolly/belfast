package packets

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/region"
)

type mockConnError struct{}

func (m *mockConnError) Read(b []byte) (int, error)  { return 0, nil }
func (m *mockConnError) Write(b []byte) (int, error) { return 0, errors.New("write failed") }
func (m *mockConnError) Close() error                { return nil }
func (m *mockConnError) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1}
}
func (m *mockConnError) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2}
}
func (m *mockConnError) SetDeadline(t time.Time) error      { return nil }
func (m *mockConnError) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConnError) SetWriteDeadline(t time.Time) error { return nil }

func newTestClientWithConn(conn net.Conn) *connection.Client {
	return &connection.Client{Connection: &conn}
}

func TestRegisterLocalizedPacketHandlerJP(t *testing.T) {
	initPacketTests(t)
	packetID := 62222
	region.SetCurrent("JP")

	jpHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 7, 0, nil }
	defaultHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 8, 0, nil }

	localized := LocalizedHandler{
		JP:      &[]PacketHandler{jpHandler},
		Default: &[]PacketHandler{defaultHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored := PacketDecisionFn[packetID]
	result, _, _ := stored[0](nil, nil)
	if result != 7 {
		t.Fatalf("expected JP handler (7), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerJPDefault(t *testing.T) {
	initPacketTests(t)
	packetID := 62221
	region.SetCurrent("JP")

	defaultHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 12, 0, nil }
	localized := LocalizedHandler{
		Default: &[]PacketHandler{defaultHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored := PacketDecisionFn[packetID]
	result, _, _ := stored[0](nil, nil)
	if result != 12 {
		t.Fatalf("expected default handler (12), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerKR(t *testing.T) {
	initPacketTests(t)
	packetID := 61111
	region.SetCurrent("KR")

	krHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 9, 0, nil }

	localized := LocalizedHandler{
		KR: &[]PacketHandler{krHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored := PacketDecisionFn[packetID]
	result, _, _ := stored[0](nil, nil)
	if result != 9 {
		t.Fatalf("expected KR handler (9), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerKRDefault(t *testing.T) {
	initPacketTests(t)
	packetID := 61110
	region.SetCurrent("KR")

	defaultHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 13, 0, nil }
	localized := LocalizedHandler{
		Default: &[]PacketHandler{defaultHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored := PacketDecisionFn[packetID]
	result, _, _ := stored[0](nil, nil)
	if result != 13 {
		t.Fatalf("expected default handler (13), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerTW(t *testing.T) {
	initPacketTests(t)
	packetID := 60001
	region.SetCurrent("TW")

	twHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 14, 0, nil }
	localized := LocalizedHandler{
		TW: &[]PacketHandler{twHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored := PacketDecisionFn[packetID]
	result, _, _ := stored[0](nil, nil)
	if result != 14 {
		t.Fatalf("expected TW handler (14), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerTWDefault(t *testing.T) {
	initPacketTests(t)
	packetID := 60000
	region.SetCurrent("TW")

	defaultHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 10, 0, nil }
	localized := LocalizedHandler{
		Default: &[]PacketHandler{defaultHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored := PacketDecisionFn[packetID]
	result, _, _ := stored[0](nil, nil)
	if result != 10 {
		t.Fatalf("expected default handler (10), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerCNDefault(t *testing.T) {
	initPacketTests(t)
	packetID := 66665
	region.SetCurrent("CN")

	defaultHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 15, 0, nil }
	localized := LocalizedHandler{
		Default: &[]PacketHandler{defaultHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored := PacketDecisionFn[packetID]
	result, _, _ := stored[0](nil, nil)
	if result != 15 {
		t.Fatalf("expected default handler (15), got %d", result)
	}
}

func TestRegisterLocalizedPacketHandlerUnknownRegionDefault(t *testing.T) {
	initPacketTests(t)
	packetID := 59999
	region.ResetCurrentForTest()
	t.Setenv("AL_REGION", "ZZ")

	defaultHandler := func(pkt *[]byte, c *connection.Client) (int, int, error) { return 11, 0, nil }
	localized := LocalizedHandler{
		Default: &[]PacketHandler{defaultHandler},
	}

	RegisterLocalizedPacketHandler(packetID, localized)

	stored := PacketDecisionFn[packetID]
	result, _, _ := stored[0](nil, nil)
	if result != 11 {
		t.Fatalf("expected default handler (11), got %d", result)
	}
}

func TestDispatchHandlerError(t *testing.T) {
	initPacketTests(t)
	PacketDecisionFn[22222] = []PacketHandler{
		func(pkt *[]byte, c *connection.Client) (int, int, error) {
			return 0, 22222, errors.New("handler failed")
		},
	}

	var conn net.Conn = &mockConn{}
	client := &connection.Client{Connection: &conn}

	buffer := []byte{
		0x00, 0x0A,
		0x00,
		0x56, 0xCE,
		0x00, 0x00,
		0x01, 0x02, 0x03, 0x04, 0x05,
	}

	Dispatch(&buffer, client, len(buffer))
	if client.MetricsSnapshot().HandlerErrors == 0 {
		t.Fatalf("expected handler errors to be recorded")
	}
	if !client.IsClosed() {
		t.Fatalf("expected client to be closed on handler error")
	}
}

func TestDispatchFlushError(t *testing.T) {
	initPacketTests(t)
	PacketDecisionFn[33333] = []PacketHandler{
		func(pkt *[]byte, c *connection.Client) (int, int, error) { return 0, 0, nil },
	}

	var conn net.Conn = &mockConnError{}
	client := newTestClientWithConn(conn)

	buffer := []byte{
		0x00, 0x0A,
		0x00,
		0x82, 0x35,
		0x00, 0x00,
		0x01, 0x02, 0x03, 0x04, 0x05,
	}

	Dispatch(&buffer, client, len(buffer))
	if client.MetricsSnapshot().WriteErrors == 0 {
		t.Fatalf("expected write errors to be recorded")
	}
}
