package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const equipBagMax = 250

func EquipToShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12006
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12006, err
	}
	response := protobuf.SC_12007{Result: proto.Uint32(0)}
	if data.GetType() != 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12007, &response)
	}
	ship, ok := client.Commander.OwnedShipsMap[data.GetShipId()]
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12007, &response)
	}
	config, err := orm.GetShipEquipConfig(ship.ShipID)
	if err != nil {
		return 0, 12006, err
	}
	pos := data.GetPos()
	slotCount := config.SlotCount()
	slotTypes := config.SlotTypes(pos)
	if pos == 0 || pos > slotCount || len(slotTypes) == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12007, &response)
	}
	entries, err := orm.ListOwnedShipEquipment(orm.GormDB, client.Commander.CommanderID, ship.ID)
	if err != nil {
		return 0, 12006, err
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
	equipID := data.GetEquipId()
	if equipID == 0 {
		if current.EquipID == 0 {
			return client.SendMessage(12007, &response)
		}
		if client.Commander.EquipmentBagCount() >= equipBagMax {
			response.Result = proto.Uint32(1)
			return client.SendMessage(12007, &response)
		}
		tx := orm.GormDB.Begin()
		if err := client.Commander.AddOwnedEquipmentTx(tx, current.EquipID, 1); err != nil {
			tx.Rollback()
			return 0, 12006, err
		}
		current.EquipID = 0
		current.SkinID = 0
		if err := orm.UpsertOwnedShipEquipmentTx(tx, current); err != nil {
			tx.Rollback()
			return 0, 12006, err
		}
		if err := tx.Commit().Error; err != nil {
			return 0, 12006, err
		}
		applyShipEquipmentUpdate(ship, current)
		return client.SendMessage(12007, &response)
	}
	if current.EquipID == equipID {
		return client.SendMessage(12007, &response)
	}
	owned := client.Commander.GetOwnedEquipment(equipID)
	if owned == nil || owned.Count == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12007, &response)
	}
	cache := make(map[uint32]*orm.Equipment)
	equipConfig, err := resolveEquipmentConfig(cache, equipID)
	if err != nil {
		return 0, 12006, err
	}
	if !containsUint32(slotTypes, equipConfig.Type) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12007, &response)
	}
	if isForbiddenShipType(equipConfig.ShipTypeForbidden, ship.Ship.Type) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12007, &response)
	}
	if equipConfig.EquipLimit != 0 {
		for _, entry := range entries {
			if entry.Pos == pos || entry.EquipID == 0 {
				continue
			}
			otherConfig, err := resolveEquipmentConfig(cache, entry.EquipID)
			if err != nil {
				return 0, 12006, err
			}
			if otherConfig.EquipLimit == equipConfig.EquipLimit {
				response.Result = proto.Uint32(1)
				return client.SendMessage(12007, &response)
			}
		}
	}
	tx := orm.GormDB.Begin()
	if current.EquipID != 0 {
		if err := client.Commander.AddOwnedEquipmentTx(tx, current.EquipID, 1); err != nil {
			tx.Rollback()
			return 0, 12006, err
		}
	}
	if err := client.Commander.RemoveOwnedEquipmentTx(tx, equipID, 1); err != nil {
		tx.Rollback()
		return 0, 12006, err
	}
	current.EquipID = equipID
	current.SkinID = 0
	if err := orm.UpsertOwnedShipEquipmentTx(tx, current); err != nil {
		tx.Rollback()
		return 0, 12006, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 12006, err
	}
	applyShipEquipmentUpdate(ship, current)
	return client.SendMessage(12007, &response)
}

func findShipEquipment(entries []orm.OwnedShipEquipment, pos uint32) *orm.OwnedShipEquipment {
	for i := range entries {
		if entries[i].Pos == pos {
			return &entries[i]
		}
	}
	return nil
}

func applyShipEquipmentUpdate(ship *orm.OwnedShip, update *orm.OwnedShipEquipment) {
	for i := range ship.Equipments {
		if ship.Equipments[i].Pos == update.Pos {
			ship.Equipments[i] = *update
			return
		}
	}
	ship.Equipments = append(ship.Equipments, *update)
}

func resolveEquipmentConfig(cache map[uint32]*orm.Equipment, equipmentID uint32) (*orm.Equipment, error) {
	if cached, ok := cache[equipmentID]; ok {
		return cached, nil
	}
	var entry orm.Equipment
	if err := orm.GormDB.First(&entry, equipmentID).Error; err != nil {
		return nil, err
	}
	if entry.Base != nil {
		var base orm.Equipment
		if err := orm.GormDB.First(&base, *entry.Base).Error; err != nil {
			return nil, err
		}
		entry = base
	}
	cache[equipmentID] = &entry
	return &entry, nil
}

func isForbiddenShipType(raw json.RawMessage, shipType uint32) bool {
	if len(raw) == 0 {
		return false
	}
	var forbidden []uint32
	if err := json.Unmarshal(raw, &forbidden); err != nil {
		return false
	}
	return containsUint32(forbidden, shipType)
}
