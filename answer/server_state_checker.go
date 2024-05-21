package answer

import (
	"fmt"
	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
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
	Servers = make([]*protobuf.SERVERINFO, len(belfastServers))
	for i, server := range belfastServers {
		Servers[i] = &protobuf.SERVERINFO{
			Ids:   []uint32{server.ID},
			Ip:    proto.String(server.IP),
			Port:  proto.Uint32(server.Port),
			State: proto.Uint32(*server.StateID - 1), // StateID is 0-based in Azur Lane, but 1-based in the database
			Name:  proto.String(server.Name),
			Sort:  proto.Uint32(uint32(i + 1)),
		}
	}
	response.Serverlist = Servers
	logger.LogEvent("Server", "SC_10019", fmt.Sprintf("sending %d servers", len(response.Serverlist)), logger.LOG_LEVEL_WARN)
	return client.SendMessage(10019, &response)
}
