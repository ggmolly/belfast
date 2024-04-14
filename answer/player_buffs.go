package answer

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/orm"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func PlayerBuffs(buffer *[]byte, client *connection.Client) (int, int, error) {
	// NOTE: This seems to be completely unused by the client, but it's here anyway
	// NOTE2: Some effects are visible in the dorm
	// TODO: Load a commander's buff with the Load() function, and send it to the client
	// currently, there are no 'applied_buffs' table entries, so this will return all buffs
	var buffs []orm.Buff
	if err := orm.GormDB.Find(&buffs).Error; err != nil {
		return 0, 11015, fmt.Errorf("failed to get buffs: %v", err)
	}

	var response protobuf.SC_11015
	response.BuffList = make([]*protobuf.BENEFITBUFF, len(buffs))
	for i, buff := range buffs {
		response.BuffList[i] = &protobuf.BENEFITBUFF{
			Id:        proto.Uint32(uint32(buff.ID)),
			Timestamp: proto.Uint32(uint32(time.Now().Add(time.Hour * 24 * 30).Unix())),
		}
	}
	logger.LogEvent("Server", "SC_11015", fmt.Sprintf("Sending %d buffs to the user", len(buffs)), logger.LOG_LEVEL_WARN)
	return client.SendMessage(11015, &response)
}
