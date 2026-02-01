package answer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/region"
	"google.golang.org/protobuf/proto"
)

func writeGatewayConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "gateway.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write gateway config: %v", err)
	}
	return path
}

func decodeGatewayPackResponse(t *testing.T, client *connection.Client) *protobuf.SC_10701 {
	t.Helper()
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetId := packets.GetPacketId(0, &buffer)
	if packetId != 10701 {
		t.Fatalf("expected packet 10701, got %d", packetId)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	var response protobuf.SC_10701
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
	return &response
}

func TestGatewayPackInfoResponse(t *testing.T) {
	serverStatusCacheEntries = nil
	serverStatusCacheRefreshedAt = time.Time{}
	versions = []string{"hash$cat$abc", "dTag-1"}

	path := writeGatewayConfig(t, "bind_address = \"0.0.0.0\"\nport = 80\n\n[[servers]]\nid = 1\nip = \"127.0.0.1\"\nport = 7000\napi_port = 0\nproxy_ip = \"127.0.0.2\"\nproxy_port = 7001\n\n[[servers]]\nid = 2\nip = \"10.0.0.1\"\nport = 7002\napi_port = 0\n")
	if _, err := config.LoadGateway(path); err != nil {
		t.Fatalf("load gateway config: %v", err)
	}

	payload := protobuf.CS_10700{
		Platform:    proto.String("1"),
		SubPlatform: proto.String(""),
		PackIndex:   proto.Uint32(0),
	}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client := &connection.Client{}
	if _, _, err := GatewayPackInfo(&data, client); err != nil {
		t.Fatalf("forge sc_10701 failed: %v", err)
	}
	response := decodeGatewayPackResponse(t, client)
	regionID := region.Current()
	url := consts.GamePlatformUrl[regionID][payload.GetPlatform()]
	if response.GetUrl() != url {
		t.Fatalf("expected url %q, got %q", url, response.GetUrl())
	}
	if len(response.GetVersion()) != len(versions) {
		t.Fatalf("expected version list size %d, got %d", len(versions), len(response.GetVersion()))
	}
	if response.GetVersion()[0] != versions[0] {
		t.Fatalf("expected version %q, got %q", versions[0], response.GetVersion()[0])
	}
	if response.GetTimestamp() == 0 {
		t.Fatalf("expected timestamp to be set")
	}
	if response.GetMonday_0OclockTimestamp() != consts.Monday_0OclockTimestamps[regionID] {
		t.Fatalf("expected monday_0oclock_timestamp to be set")
	}
	if len(response.GetCdnList()) != 0 {
		t.Fatalf("expected empty cdn_list")
	}
	addrList := response.GetAddrList()
	if len(addrList) != 2 {
		t.Fatalf("expected 2 addr entries, got %d", len(addrList))
	}
	if addrList[0].GetDesc() != "127.0.0.1" {
		t.Fatalf("expected desc 127.0.0.1, got %q", addrList[0].GetDesc())
	}
	if addrList[0].GetIp() != "127.0.0.1" || addrList[0].GetPort() != 7000 {
		t.Fatalf("expected addr 127.0.0.1:7000")
	}
	if addrList[0].GetProxyIp() != "127.0.0.2" || addrList[0].GetProxyPort() != 7001 {
		t.Fatalf("expected proxy 127.0.0.2:7001")
	}
	if addrList[1].GetDesc() != "10.0.0.1" {
		t.Fatalf("expected desc 10.0.0.1, got %q", addrList[1].GetDesc())
	}
	if addrList[1].GetProxyIp() != "" || addrList[1].GetProxyPort() != 0 {
		t.Fatalf("expected empty proxy for second server")
	}
	if addrList[0].GetType() != 0 || addrList[1].GetType() != 0 {
		t.Fatalf("expected addr type 0")
	}
}

func TestGatewayPackInfoUnknownPlatform(t *testing.T) {
	serverStatusCacheEntries = nil
	serverStatusCacheRefreshedAt = time.Time{}
	versions = []string{"hash$cat$abc", "dTag-1"}

	path := writeGatewayConfig(t, "bind_address = \"0.0.0.0\"\nport = 80\n\n[[servers]]\nid = 1\nip = \"127.0.0.1\"\nport = 7000\napi_port = 0\n")
	if _, err := config.LoadGateway(path); err != nil {
		t.Fatalf("load gateway config: %v", err)
	}

	payload := protobuf.CS_10700{
		Platform:    proto.String("99"),
		SubPlatform: proto.String(""),
		PackIndex:   proto.Uint32(0),
	}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client := &connection.Client{}
	if _, _, err := GatewayPackInfo(&data, client); err == nil {
		t.Fatalf("expected error for unknown platform")
	}
}
