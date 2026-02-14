package answer

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	serverStatusCacheTTL = 30 * time.Second
	serverStatusTimeout  = 2 * time.Second
)

var (
	serverStatusCacheMu          sync.Mutex
	serverStatusCacheRefreshedAt time.Time
	serverStatusCacheEntries     map[uint32]serverStatusEntry
	serverStatusProbeFn          = probeServerStatus
)

type serverStatusProbeData struct {
	ServerLoad uint32
	DBLoad     uint32
}

type serverStatusEntry struct {
	Name       string
	Commit     string
	State      uint32
	ServerLoad uint32
	DBLoad     uint32
}

func getServerStatusCache(servers []config.ServerConfig) map[uint32]serverStatusEntry {
	serverStatusCacheMu.Lock()
	defer serverStatusCacheMu.Unlock()
	if time.Since(serverStatusCacheRefreshedAt) < serverStatusCacheTTL && serverStatusCacheEntries != nil {
		return serverStatusCacheEntries
	}
	entries := make(map[uint32]serverStatusEntry, len(servers))
	for i := range servers {
		server := servers[i]
		entries[server.ID] = resolveServerStatus(server)
	}
	serverStatusCacheEntries = entries
	serverStatusCacheRefreshedAt = time.Now().UTC()
	return serverStatusCacheEntries
}

func resolveServerStatus(server config.ServerConfig) serverStatusEntry {
	entry := serverStatusEntry{
		Name:       server.IP,
		Commit:     "",
		State:      SERVER_STATE_OFFLINE,
		ServerLoad: 0,
		DBLoad:     0,
	}
	if server.AssertOnline {
		entry.State = SERVER_STATE_ONLINE
		return entry
	}
	probe, err := serverStatusProbeFn(server)
	if err != nil {
		logger.LogEvent("Server", "StatusRefresh", fmt.Sprintf("status probe failed for %s:%d: %s", server.IP, server.Port, err.Error()), logger.LOG_LEVEL_WARN)
		return entry
	}
	entry.ServerLoad = probe.ServerLoad
	entry.DBLoad = probe.DBLoad
	if probe.ServerLoad >= 80 || probe.DBLoad >= 80 {
		entry.State = SERVER_STATE_BUSY
		return entry
	}
	entry.State = SERVER_STATE_ONLINE
	return entry
}

func probeServerStatus(server config.ServerConfig) (serverStatusProbeData, error) {
	requestPayload := protobuf.CS_10022{
		AccountId:    proto.Uint32(0),
		ServerTicket: proto.String(serverTicketPrefix),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(server.ID),
		CheckKey:     proto.String("status_probe"),
		DeviceId:     proto.String(""),
	}
	data, err := proto.Marshal(&requestPayload)
	if err != nil {
		return serverStatusProbeData{}, err
	}
	connection.InjectPacketHeader(10022, &data, 0)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port), serverStatusTimeout)
	if err != nil {
		return serverStatusProbeData{}, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(serverStatusTimeout)); err != nil {
		return serverStatusProbeData{}, err
	}
	if _, err := conn.Write(data); err != nil {
		return serverStatusProbeData{}, err
	}

	packetData, err := readSinglePacket(conn)
	if err != nil {
		return serverStatusProbeData{}, err
	}
	if packets.GetPacketId(0, &packetData) != 10023 {
		return serverStatusProbeData{}, fmt.Errorf("unexpected packet id %d", packets.GetPacketId(0, &packetData))
	}
	packetSize := packets.GetPacketSize(0, &packetData) + 2
	if packetSize < packets.HEADER_SIZE || len(packetData) < packetSize {
		return serverStatusProbeData{}, fmt.Errorf("invalid packet size %d", packetSize)
	}

	var responsePayload protobuf.SC_10023
	if err := proto.Unmarshal(packetData[packets.HEADER_SIZE:packetSize], &responsePayload); err != nil {
		return serverStatusProbeData{}, err
	}
	return serverStatusProbeData{ServerLoad: responsePayload.GetServerLoad(), DBLoad: responsePayload.GetDbLoad()}, nil
}

func readSinglePacket(conn net.Conn) ([]byte, error) {
	sizeHeader := make([]byte, 2)
	if _, err := io.ReadFull(conn, sizeHeader); err != nil {
		return nil, err
	}
	size := int(sizeHeader[0])<<8 | int(sizeHeader[1])
	if size < 5 {
		return nil, fmt.Errorf("invalid packet size %d", size)
	}
	packetData := make([]byte, size+2)
	copy(packetData[:2], sizeHeader)
	if _, err := io.ReadFull(conn, packetData[2:]); err != nil {
		return nil, err
	}
	return packetData, nil
}
