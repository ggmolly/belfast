package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func TransformEquipmentOnShip14013(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14013
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14013, err
	}

	response := protobuf.SC_14014{Result: proto.Uint32(0)}
	ship, ok := client.Commander.OwnedShipsMap[data.GetShipId()]
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14014, &response)
	}
	config, err := orm.GetShipEquipConfig(ship.ShipID)
	if err != nil {
		return 0, 14013, err
	}
	pos := data.GetPos()
	slotCount := config.SlotCount()
	slotTypes := config.SlotTypes(pos)
	if pos == 0 || pos > slotCount || len(slotTypes) == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14014, &response)
	}

	entries, err := orm.ListOwnedShipEquipment(orm.GormDB, client.Commander.CommanderID, ship.ID)
	if err != nil {
		return 0, 14013, err
	}
	current := findShipEquipment(entries, pos)
	if current == nil {
		current = &orm.OwnedShipEquipment{
			OwnerID: client.Commander.CommanderID,
			ShipID:  ship.ID,
			Pos:     pos,
			EquipID: 0,
			SkinID:  0,
		}
	}
	if current.EquipID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14014, &response)
	}

	upgrade, err := orm.GetEquipUpgradeDataTx(orm.GormDB, data.GetUpgradeId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14014, &response)
		}
		return 0, 14013, err
	}
	if upgrade.UpgradeFrom != current.EquipID {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14014, &response)
	}
	targetEquipID := upgrade.TargetID
	if targetEquipID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14014, &response)
	}

	if client.Commander.GetResourceCount(1) < upgrade.CoinConsume {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14014, &response)
	}
	for _, cost := range upgrade.MaterialCost {
		if client.Commander.GetItemCount(cost.ItemID) < cost.Count {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14014, &response)
		}
	}

	allowed, err := equipmentAllowedAtPos(entries, pos, ship, slotTypes, targetEquipID)
	if err != nil {
		return 0, 14013, err
	}
	if !allowed {
		if client.Commander.EquipmentBagCount() >= equipBagMax {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14014, &response)
		}
	}

	tx := orm.GormDB.Begin()
	if upgrade.CoinConsume != 0 {
		if err := client.Commander.ConsumeResourceTx(tx, 1, upgrade.CoinConsume); err != nil {
			tx.Rollback()
			response.Result = proto.Uint32(1)
			return client.SendMessage(14014, &response)
		}
	}
	for _, cost := range upgrade.MaterialCost {
		if err := client.Commander.ConsumeItemTx(tx, cost.ItemID, cost.Count); err != nil {
			tx.Rollback()
			response.Result = proto.Uint32(1)
			return client.SendMessage(14014, &response)
		}
	}

	if allowed {
		current.EquipID = targetEquipID
		current.SkinID = 0
	} else {
		current.EquipID = 0
		current.SkinID = 0
		if err := client.Commander.AddOwnedEquipmentTx(tx, targetEquipID, 1); err != nil {
			tx.Rollback()
			return 0, 14013, err
		}
	}
	if err := orm.UpsertOwnedShipEquipmentTx(tx, current); err != nil {
		tx.Rollback()
		return 0, 14013, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 14013, err
	}
	applyShipEquipmentUpdate(ship, current)
	return client.SendMessage(14014, &response)
}

func equipmentAllowedAtPos(entries []orm.OwnedShipEquipment, pos uint32, ship *orm.OwnedShip, slotTypes []uint32, equipmentID uint32) (bool, error) {
	cache := make(map[uint32]*orm.Equipment)
	equipConfig, err := resolveEquipmentConfig(cache, equipmentID)
	if err != nil {
		return false, err
	}
	if !containsUint32(slotTypes, equipConfig.Type) {
		return false, nil
	}
	if isForbiddenShipType(equipConfig.ShipTypeForbidden, ship.Ship.Type) {
		return false, nil
	}
	if equipConfig.EquipLimit != 0 {
		for _, entry := range entries {
			if entry.Pos == pos || entry.EquipID == 0 {
				continue
			}
			otherConfig, err := resolveEquipmentConfig(cache, entry.EquipID)
			if err != nil {
				return false, err
			}
			if otherConfig.EquipLimit == equipConfig.EquipLimit {
				return false, nil
			}
		}
	}
	return true, nil
}
