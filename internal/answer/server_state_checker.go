package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

func Forge_SC10019(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_10019{}

	// Update server list
	var belfastServers []orm.Server
	// Decrease by 1 the state of all servers
	err := orm.GormDB.Order("id asc").Find(&belfastServers).Error
	if err != nil {
		logger.LogEvent("Server", "SC_10019", fmt.Sprintf("failed to fetch servers: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		return 0, 10019, err
	}
	Servers = buildServerInfo(belfastServers)
	response.Serverlist = Servers
	logger.LogEvent("Server", "SC_10019", fmt.Sprintf("sending %d servers", len(response.Serverlist)), logger.LOG_LEVEL_WARN)
	return client.SendMessage(10019, &response)
}
