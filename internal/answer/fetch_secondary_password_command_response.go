package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FetchSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	state, err := orm.GetOrCreateSecondaryPasswordState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11604, err
	}
	response := protobuf.SC_11604{
		State:      proto.Uint32(state.State),
		FailCd:     proto.Uint32(state.FailCd),
		FailCount:  proto.Uint32(state.FailCount),
		Notice:     proto.String(state.Notice),
		SystemList: orm.ToUint32List(state.SystemList),
	}
	return client.SendMessage(11604, &response)
}
