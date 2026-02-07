package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func ChapterBaseSync(_ *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_13000{
		DailyRepairCount: proto.Uint32(0),
	}

	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return client.SendMessage(13000, &response)
		}
		return 0, 13000, err
	}
	if len(state.State) == 0 {
		return client.SendMessage(13000, &response)
	}

	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		return 0, 13000, err
	}
	response.CurrentChapter = &current

	return client.SendMessage(13000, &response)
}
