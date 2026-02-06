package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func FinishPhantomQuest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12210
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12211, err
	}
	response := protobuf.SC_12211{Result: proto.Uint32(0)}
	shipID := payload.GetShipId()
	shadowID := payload.GetSkinShadowId()
	ship, ok := client.Commander.OwnedShipsMap[shipID]
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	quest, err := orm.GetTechnologyShadowUnlockConfig(shadowID)
	if err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	skinID, err := orm.GetShipBaseSkinIDTx(orm.GormDB, ship.ShipID)
	if err != nil || skinID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	var existing orm.OwnedShipShadowSkin
	if err := tx.First(&existing, "commander_id = ? AND ship_id = ? AND shadow_id = ?", client.Commander.CommanderID, shipID, shadowID).Error; err == nil {
		if existing.SkinID != skinID {
			if err := orm.UpsertOwnedShipShadowSkin(tx, client.Commander.CommanderID, shipID, shadowID, skinID); err != nil {
				tx.Rollback()
				response.Result = proto.Uint32(1)
				return client.SendMessage(12211, &response)
			}
			if err := tx.Commit().Error; err != nil {
				response.Result = proto.Uint32(1)
				return client.SendMessage(12211, &response)
			}
		} else {
			tx.Rollback()
		}
		return client.SendMessage(12211, &response)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	if quest.Type == 5 {
		if err := client.Commander.ConsumeResourceTx(tx, 4, quest.TargetNum); err != nil {
			tx.Rollback()
			response.Result = proto.Uint32(1)
			return client.SendMessage(12211, &response)
		}
	}
	if err := orm.UpsertOwnedShipShadowSkin(tx, client.Commander.CommanderID, shipID, shadowID, skinID); err != nil {
		tx.Rollback()
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	if err := tx.Commit().Error; err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	return client.SendMessage(12211, &response)
}
