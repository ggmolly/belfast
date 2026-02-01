package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func PermanentActivites(buffer *[]byte, client *connection.Client) (int, int, error) {
	state, err := orm.GetOrCreatePermanentActivityState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11210, err
	}
	permanentIDs, err := loadPermanentActivityIDSet()
	if err != nil {
		return 0, 11210, err
	}
	finished := filterPermanentActivityIDs(orm.ToUint32List(state.FinishedActivityIDs), permanentIDs)
	response := protobuf.SC_11210{
		PermanentActivity: finished,
	}
	if state.PermanentNow != 0 {
		if _, ok := permanentIDs[state.PermanentNow]; ok {
			response.PermanentNow = proto.Uint32(state.PermanentNow)
		} else {
			response.PermanentNow = proto.Uint32(0)
		}
	} else {
		response.PermanentNow = proto.Uint32(0)
	}
	return client.SendMessage(11210, &response)
}
