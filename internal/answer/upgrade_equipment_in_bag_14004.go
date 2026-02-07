package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	upgradeEquipmentInBagResultOK             uint32 = 0
	upgradeEquipmentInBagResultGenericFailure uint32 = 1
)

func UpgradeEquipmentInBag14004(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_14005{Result: proto.Uint32(upgradeEquipmentInBagResultGenericFailure)}

	if client == nil || client.Commander == nil {
		return client.SendMessage(14005, &response)
	}

	var payload protobuf.CS_14004
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return client.SendMessage(14005, &response)
	}

	equipID := payload.GetEquipId()
	lv := payload.GetLv()
	if equipID == 0 || lv == 0 {
		return client.SendMessage(14005, &response)
	}

	if client.Commander.OwnedEquipmentMap == nil {
		if err := client.Commander.Load(); err != nil {
			return client.SendMessage(14005, &response)
		}
	}

	owned := client.Commander.GetOwnedEquipment(equipID)
	if owned == nil || owned.Count < 1 {
		return client.SendMessage(14005, &response)
	}

	upgradedID, ok := resolveEquipmentUpgradeChain(equipID, lv)
	if !ok {
		return client.SendMessage(14005, &response)
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(14005, &response)
	}
	if err := client.Commander.RemoveOwnedEquipmentTx(tx, equipID, 1); err != nil {
		tx.Rollback()
		return client.SendMessage(14005, &response)
	}
	if err := client.Commander.AddOwnedEquipmentTx(tx, upgradedID, 1); err != nil {
		tx.Rollback()
		return client.SendMessage(14005, &response)
	}
	if err := tx.Commit().Error; err != nil {
		return client.SendMessage(14005, &response)
	}

	response.Result = proto.Uint32(upgradeEquipmentInBagResultOK)
	return client.SendMessage(14005, &response)
}

func resolveEquipmentUpgradeChain(startID uint32, lv uint32) (uint32, bool) {
	currentID := startID
	for i := uint32(0); i < lv; i++ {
		var current orm.Equipment
		if err := orm.GormDB.First(&current, currentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, false
			}
			return 0, false
		}
		if current.Next == 0 {
			return 0, false
		}
		currentID = uint32(current.Next)
	}
	return currentID, true
}
