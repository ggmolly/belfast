package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func ChapterBattleResultRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13106
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13105, err
	}
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response := protobuf.SC_13105{
				MapUpdate:    []*protobuf.CHAPTERCELLINFO_P13{},
				AiList:       []*protobuf.CHAPTERCELLINFO_P13{},
				AddFlagList:  []uint32{},
				DelFlagList:  []uint32{},
				BuffList:     []uint32{},
				CellFlagList: []*protobuf.CELLFLAG{},
			}
			return client.SendMessage(13105, &response)
		}
		return 0, 13105, err
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		return 0, 13105, err
	}
	response := protobuf.SC_13105{
		MapUpdate:    current.GetCellList(),
		AiList:       current.GetAiList(),
		AddFlagList:  []uint32{},
		DelFlagList:  []uint32{},
		BuffList:     current.GetBuffList(),
		CellFlagList: current.GetCellFlagList(),
	}
	return client.SendMessage(13105, &response)
}
