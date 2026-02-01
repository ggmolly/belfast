package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func ActivityPermanentFinish(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11208
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11209, err
	}
	response := protobuf.SC_11209{Result: proto.Uint32(0)}

	permanentIDs, err := loadPermanentActivityIDSet()
	if err != nil {
		return 0, 11209, err
	}
	if _, ok := permanentIDs[payload.GetActivityId()]; !ok {
		return client.SendMessage(11209, &response)
	}
	if _, err := loadActivityTemplate(payload.GetActivityId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return client.SendMessage(11209, &response)
		}
		return 0, 11209, err
	}

	state, err := orm.GetOrCreatePermanentActivityState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11209, err
	}
	finished := orm.ToUint32List(state.FinishedActivityIDs)
	finished = appendUniqueUint32(finished, payload.GetActivityId())
	state.FinishedActivityIDs = orm.ToInt64List(finished)
	state.PermanentNow = 0
	if err := orm.GormDB.Save(state).Error; err != nil {
		return 0, 11209, err
	}

	return client.SendMessage(11209, &response)
}
