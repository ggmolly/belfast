package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm/clause"
)

const (
	equipCodeLikeResultOK          uint32 = 0
	equipCodeLikeResultErr         uint32 = 1
	equipCodeLikeResultAlreadyLike uint32 = 7
)

func EquipCodeLike(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_17605
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 17606, err
	}

	response := protobuf.SC_17606{Result: proto.Uint32(equipCodeLikeResultOK)}
	shipGroupID := data.GetShipgroup()
	shareID := data.GetShareid()
	if shipGroupID == 0 || shareID == 0 {
		response.Result = proto.Uint32(equipCodeLikeResultErr)
		return client.SendMessage(17606, &response)
	}

	now := time.Now().UTC()
	day := uint32(now.Unix() / 86400)
	commanderID := client.Commander.CommanderID

	like := orm.EquipCodeLike{
		CommanderID: commanderID,
		ShipGroupID: shipGroupID,
		ShareID:     shareID,
		LikeDay:     day,
		CreatedAt:   now,
	}

	tx := orm.GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "ship_group_id"}, {Name: "share_id"}, {Name: "like_day"}},
		DoNothing: true,
	}).Create(&like)
	if tx.Error != nil {
		return 0, 17606, tx.Error
	}
	if tx.RowsAffected == 0 {
		response.Result = proto.Uint32(equipCodeLikeResultAlreadyLike)
	}

	return client.SendMessage(17606, &response)
}
