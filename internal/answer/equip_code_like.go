package answer

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
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

	res, err := db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO equip_code_likes (commander_id, ship_group_id, share_id, like_day, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (commander_id, ship_group_id, share_id, like_day)
DO NOTHING
`, int64(like.CommanderID), int64(like.ShipGroupID), int64(like.ShareID), int64(like.LikeDay), like.CreatedAt)
	if err != nil {
		return 0, 17606, err
	}
	if res.RowsAffected() == 0 {
		response.Result = proto.Uint32(equipCodeLikeResultAlreadyLike)
	}

	return client.SendMessage(17606, &response)
}
