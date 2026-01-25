package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CommanderStoryProgress(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_13001{}
	state, err := orm.GetOrCreateRemasterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 13001, err
	}
	if orm.ApplyRemasterDailyReset(state, time.Now()) {
		if err := orm.GormDB.Save(state).Error; err != nil {
			return 0, 13001, err
		}
	}
	response.ReactChapter = &protobuf.REACTCHAPTER_INFO{
		Count:           proto.Uint32(state.TicketCount),
		ActiveTimestamp: proto.Uint32(uint32(state.LastDailyResetAt.Unix())),
		ActiveId:        proto.Uint32(0),
		DailyCount:      proto.Uint32(state.DailyCount),
	}
	return client.SendMessage(13001, &response)
}
