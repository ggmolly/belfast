package answer

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestServerStatusCacheMappingFromProbeLoads(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		serverLoad  uint32
		dbLoad      uint32
		expected    uint32
	}{
		{name: "online", displayName: "Belfast", serverLoad: 10, dbLoad: 20, expected: SERVER_STATE_ONLINE},
		{name: "busy by server load", displayName: "Suffolk", serverLoad: 80, dbLoad: 0, expected: SERVER_STATE_BUSY},
		{name: "busy by db load", displayName: "Javelin", serverLoad: 0, dbLoad: 80, expected: SERVER_STATE_BUSY},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, stop := startProbeStatusServer(t, 1, tt.serverLoad, tt.dbLoad)
			defer stop()

			serverStatusCacheRefreshedAt = time.Time{}
			serverStatusCacheEntries = nil
			statuses := getServerStatusCache([]config.ServerConfig{{ID: 1, Name: tt.displayName, IP: host, Port: port}})

			status := statuses[1]
			if status.State != tt.expected {
				t.Fatalf("expected state %d, got %d", tt.expected, status.State)
			}
			if status.Name != tt.displayName {
				t.Fatalf("expected name %q, got %q", tt.displayName, status.Name)
			}
			if status.Commit != "" {
				t.Fatalf("expected empty commit, got %q", status.Commit)
			}
			if status.ServerLoad != tt.serverLoad {
				t.Fatalf("expected server load %d, got %d", tt.serverLoad, status.ServerLoad)
			}
			if status.DBLoad != tt.dbLoad {
				t.Fatalf("expected db load %d, got %d", tt.dbLoad, status.DBLoad)
			}

			serverInfo := buildServerInfo([]config.ServerConfig{{ID: 1, Name: tt.displayName, IP: host, Port: port}}, statuses)
			if got := serverInfo[0].GetName(); got != tt.displayName {
				t.Fatalf("expected formatted name %q, got %q", tt.displayName, got)
			}
		})
	}
}

func TestServerStatusCacheFallbackNameUsesIP(t *testing.T) {
	host, port, stop := startProbeStatusServer(t, 3, 5, 5)
	defer stop()

	serverStatusCacheRefreshedAt = time.Time{}
	serverStatusCacheEntries = nil
	statuses := getServerStatusCache([]config.ServerConfig{{ID: 3, IP: host, Port: port}})

	if status := statuses[3]; status.Name != host {
		t.Fatalf("expected fallback name %q, got %q", host, status.Name)
	}
}

func TestServerStatusCacheProbeErrorMapsToOffline(t *testing.T) {
	port := reserveUnusedPort(t)

	serverStatusCacheRefreshedAt = time.Time{}
	serverStatusCacheEntries = nil
	statuses := getServerStatusCache([]config.ServerConfig{{ID: 9, IP: "127.0.0.1", Port: port}})

	status := statuses[9]
	if status.State != SERVER_STATE_OFFLINE {
		t.Fatalf("expected offline state, got %d", status.State)
	}
}

func TestServerStatusCacheAssertOnlineSkipsProbe(t *testing.T) {
	previousProbeFn := serverStatusProbeFn
	probeCalled := false
	serverStatusProbeFn = func(_ config.ServerConfig) (serverStatusProbeData, error) {
		probeCalled = true
		return serverStatusProbeData{}, errors.New("unexpected probe")
	}
	defer func() { serverStatusProbeFn = previousProbeFn }()

	serverStatusCacheRefreshedAt = time.Time{}
	serverStatusCacheEntries = nil
	statuses := getServerStatusCache([]config.ServerConfig{{ID: 1, IP: "203.0.113.1", Port: 7000, AssertOnline: true}})

	if probeCalled {
		t.Fatalf("expected no status probe when assert_online is enabled")
	}
	if status := statuses[1]; status.State != SERVER_STATE_ONLINE {
		t.Fatalf("expected online state, got %d", status.State)
	}
}

func startProbeStatusServer(t *testing.T, expectedServerID uint32, serverLoad uint32, dbLoad uint32) (string, uint32, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen probe server: %v", err)
	}

	errCh := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				errCh <- err
			}
			return
		}
		defer conn.Close()

		packetData, err := readPacketForTest(conn)
		if err != nil {
			errCh <- err
			return
		}
		if packetID := packets.GetPacketId(0, &packetData); packetID != 10022 {
			errCh <- fmt.Errorf("expected packet id 10022, got %d", packetID)
			return
		}
		packetSize := packets.GetPacketSize(0, &packetData) + 2
		if packetSize < packets.HEADER_SIZE || len(packetData) < packetSize {
			errCh <- fmt.Errorf("invalid packet size %d", packetSize)
			return
		}

		var request protobuf.CS_10022
		if err := proto.Unmarshal(packetData[packets.HEADER_SIZE:packetSize], &request); err != nil {
			errCh <- err
			return
		}
		if request.GetAccountId() != 0 {
			errCh <- fmt.Errorf("unexpected account id %d", request.GetAccountId())
			return
		}
		if request.GetServerTicket() != serverTicketPrefix {
			errCh <- fmt.Errorf("unexpected server ticket %q", request.GetServerTicket())
			return
		}
		if request.GetPlatform() != "0" {
			errCh <- fmt.Errorf("unexpected platform %q", request.GetPlatform())
			return
		}
		if request.GetServerid() != expectedServerID {
			errCh <- fmt.Errorf("unexpected server id %d", request.GetServerid())
			return
		}
		if request.GetCheckKey() != "status_probe" {
			errCh <- fmt.Errorf("unexpected check key %q", request.GetCheckKey())
			return
		}
		if request.GetDeviceId() != "" {
			errCh <- fmt.Errorf("unexpected device id %q", request.GetDeviceId())
			return
		}

		response := protobuf.SC_10023{
			Result:       proto.Uint32(0),
			UserId:       proto.Uint32(0),
			ServerTicket: proto.String(serverTicketPrefix),
			ServerLoad:   proto.Uint32(serverLoad),
			DbLoad:       proto.Uint32(dbLoad),
		}
		data, err := proto.Marshal(&response)
		if err != nil {
			errCh <- err
			return
		}
		connection.InjectPacketHeader(10023, &data, 0)
		if _, err := conn.Write(data); err != nil {
			errCh <- err
		}
	}()

	host, port := parseListenerHostPort(t, listener.Addr().String())
	stop := func() {
		_ = listener.Close()
		<-done
		select {
		case err := <-errCh:
			t.Fatalf("probe server failed: %v", err)
		default:
		}
	}
	return host, port, stop
}

func readPacketForTest(conn net.Conn) ([]byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	size := int(header[0])<<8 | int(header[1])
	if size < 5 {
		return nil, fmt.Errorf("invalid packet size %d", size)
	}
	packetData := make([]byte, size+2)
	copy(packetData[:2], header)
	if _, err := io.ReadFull(conn, packetData[2:]); err != nil {
		return nil, err
	}
	return packetData, nil
}

func parseListenerHostPort(t *testing.T, address string) (string, uint32) {
	t.Helper()
	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return host, uint32(port)
}

func reserveUnusedPort(t *testing.T) uint32 {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	_, port := parseListenerHostPort(t, listener.Addr().String())
	_ = listener.Close()
	return port
}
