package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func InstagramChatActivateTopic(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11722
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11723, err
	}
	if client.Commander == nil {
		return 0, 11723, errors.New("missing commander")
	}
	now := uint32(time.Now().Unix())
	chatGroupIDs := payload.GetChatGroupIdList()
	resultList := make([]uint32, len(chatGroupIDs))
	for i, chatGroupID := range chatGroupIDs {
		if chatGroupID == 0 {
			resultList[i] = 1
			continue
		}
		config, err := getJuustagramChatGroupConfig(chatGroupID)
		if err != nil || config.ShipGroup == 0 {
			resultList[i] = 1
			continue
		}
		if _, err := orm.EnsureJuustagramGroupExists(client.Commander.CommanderID, config.ShipGroup, chatGroupID); err != nil {
			resultList[i] = 1
			continue
		}
		if _, err := orm.GetJuustagramChatGroup(client.Commander.CommanderID, chatGroupID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if _, err := orm.CreateJuustagramChatGroup(client.Commander.CommanderID, config.ShipGroup, chatGroupID, now); err != nil {
					resultList[i] = 1
					continue
				}
			} else {
				resultList[i] = 1
				continue
			}
		}
		resultList[i] = 0
	}
	response := protobuf.SC_11723{
		ResultList: resultList,
		OpTime:     proto.Uint32(now),
	}
	return client.SendMessage(11723, &response)
}
