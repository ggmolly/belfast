package answer

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func PlayerBuffs(buffer *[]byte, client *connection.Client) (int, int, error) {
	// Some effects are visible in the dorm, so this needs to be accurate.
	now := time.Now().UTC()
	buffs, err := orm.ListCommanderActiveBuffs(client.Commander.CommanderID, now)
	if err != nil {
		return 0, 11015, fmt.Errorf("failed to get buffs: %v", err)
	}
	var response protobuf.SC_11015
	response.BuffList = make([]*protobuf.BENEFITBUFF, len(buffs))
	for i, buff := range buffs {
		response.BuffList[i] = &protobuf.BENEFITBUFF{
			Id:        proto.Uint32(buff.BuffID),
			Timestamp: proto.Uint32(uint32(buff.ExpiresAt.UTC().Unix())),
		}
	}
	logger.LogEvent("Server", "SC_11015", fmt.Sprintf("Sending %d buffs to the user", len(buffs)), logger.LOG_LEVEL_WARN)
	return client.SendMessage(11015, &response)
}
