package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func NewEducateRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29001
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29002, err
	}
	tbState, err := orm.GetCommanderTB(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 29002, err
		}
		entry, err := orm.NewCommanderTB(client.Commander.CommanderID, tbInfoPlaceholder(), tbPermanentPlaceholder())
		if err != nil {
			return 0, 29002, err
		}
		if err := orm.GormDB.Create(entry).Error; err != nil {
			return 0, 29002, err
		}
		tbState = entry
	}
	info, permanent, err := tbState.Decode()
	if err != nil {
		return 0, 29002, err
	}
	response := protobuf.SC_29002{
		Result:    proto.Uint32(0),
		Tb:        info,
		Permanent: permanent,
	}
	response.Tb.Id = proto.Uint32(payload.GetId())
	return client.SendMessage(29002, &response)
}
