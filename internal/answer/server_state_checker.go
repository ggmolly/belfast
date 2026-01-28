package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/protobuf"
)

func Forge_SC10019(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_10019{}

	// Update server list
	statuses := getServerStatusCache(config.Current().Servers)
	Servers = buildServerInfo(config.Current().Servers, statuses)
	response.Serverlist = Servers
	logger.LogEvent("Server", "SC_10019", fmt.Sprintf("sending %d servers", len(response.Serverlist)), logger.LOG_LEVEL_WARN)
	return client.SendMessage(10019, &response)
}
