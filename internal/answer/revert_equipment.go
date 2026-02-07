package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const revertEquipmentItemID uint32 = 15007

func RevertEquipment(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14010
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14010, err
	}

	response := protobuf.SC_14011{Result: proto.Uint32(0)}
	equipID := data.GetEquipId()
	if equipID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}
	owned := client.Commander.GetOwnedEquipment(equipID)
	if owned == nil || owned.Count == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}
	if !client.Commander.HasEnoughItem(revertEquipmentItemID, 1) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}

	rootEquipID, refundItems, refundCoins, ok, err := computeRevertEquipmentRefunds(orm.GormDB, equipID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14011, &response)
		}
		return 0, 14010, err
	}
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}

	tx := orm.GormDB.Begin()
	if err := client.Commander.ConsumeItemTx(tx, revertEquipmentItemID, 1); err != nil {
		tx.Rollback()
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}
	if err := client.Commander.RemoveOwnedEquipmentTx(tx, equipID, 1); err != nil {
		tx.Rollback()
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}
	if err := client.Commander.AddOwnedEquipmentTx(tx, rootEquipID, 1); err != nil {
		tx.Rollback()
		return 0, 14010, err
	}
	for itemID, count := range refundItems {
		if err := client.Commander.AddItemTx(tx, itemID, count); err != nil {
			tx.Rollback()
			return 0, 14010, err
		}
	}
	if refundCoins != 0 {
		if err := client.Commander.AddResourceTx(tx, 1, refundCoins); err != nil {
			tx.Rollback()
			return 0, 14010, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 14010, err
	}
	return client.SendMessage(14011, &response)
}

func computeRevertEquipmentRefunds(db *gorm.DB, equipID uint32) (uint32, map[uint32]uint32, uint32, bool, error) {
	var current orm.Equipment
	if err := db.First(&current, equipID).Error; err != nil {
		return 0, nil, 0, false, err
	}
	if current.Prev == 0 || current.Level <= 1 {
		return 0, nil, 0, false, nil
	}

	refundItems := make(map[uint32]uint32)
	var refundCoins uint32
	for current.Prev != 0 {
		prevID := uint32(current.Prev)
		var prev orm.Equipment
		if err := db.First(&prev, prevID).Error; err != nil {
			return 0, nil, 0, false, err
		}
		refundCoins += prev.TransUseGold
		if err := addTransUseItems(refundItems, prev.TransUseItem); err != nil {
			return 0, nil, 0, false, err
		}
		current = prev
	}

	return current.ID, refundItems, refundCoins, true, nil
}
