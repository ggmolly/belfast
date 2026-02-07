package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func TransformEquipmentInBag14015(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14015
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14015, err
	}

	response := protobuf.SC_14016{Result: proto.Uint32(0)}
	equipID := data.GetEquipId()
	if equipID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}
	owned := client.Commander.GetOwnedEquipment(equipID)
	if owned == nil || owned.Count == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}

	upgrade, err := orm.GetEquipUpgradeDataTx(orm.GormDB, data.GetUpgradeId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14016, &response)
		}
		return 0, 14015, err
	}
	if upgrade.UpgradeFrom != equipID {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}
	targetEquipID := upgrade.TargetID
	if targetEquipID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}

	if client.Commander.GetResourceCount(1) < upgrade.CoinConsume {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}
	for _, cost := range upgrade.MaterialCost {
		if client.Commander.GetItemCount(cost.ItemID) < cost.Count {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14016, &response)
		}
	}

	tx := orm.GormDB.Begin()
	if upgrade.CoinConsume != 0 {
		if err := client.Commander.ConsumeResourceTx(tx, 1, upgrade.CoinConsume); err != nil {
			tx.Rollback()
			response.Result = proto.Uint32(1)
			return client.SendMessage(14016, &response)
		}
	}
	for _, cost := range upgrade.MaterialCost {
		if err := client.Commander.ConsumeItemTx(tx, cost.ItemID, cost.Count); err != nil {
			tx.Rollback()
			response.Result = proto.Uint32(1)
			return client.SendMessage(14016, &response)
		}
	}
	if err := client.Commander.RemoveOwnedEquipmentTx(tx, equipID, 1); err != nil {
		tx.Rollback()
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}
	if err := client.Commander.AddOwnedEquipmentTx(tx, targetEquipID, 1); err != nil {
		tx.Rollback()
		return 0, 14015, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 14015, err
	}
	return client.SendMessage(14016, &response)
}
