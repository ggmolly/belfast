package answer

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type equipSkinTemplateConfig struct {
	EquipType []uint32 `json:"equip_type"`
}

func UpdateShipEquipmentSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12036
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12036, err
	}

	response := protobuf.SC_12037{Result: proto.Uint32(0)}

	ship, ok := client.Commander.OwnedShipsMap[data.GetShipId()]
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12037, &response)
	}

	config, err := orm.GetShipEquipConfig(ship.ShipID)
	if err != nil {
		return 0, 12036, err
	}
	pos := data.GetPos()
	slotCount := config.SlotCount()
	slotTypes := config.SlotTypes(pos)
	if pos == 0 || pos > slotCount || len(slotTypes) == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12037, &response)
	}

	current, err := orm.GetOwnedShipEquipment(orm.GormDB, client.Commander.CommanderID, ship.ID, pos)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			current = buildShipEquipmentFromMemory(client.Commander.CommanderID, ship, pos)
		} else {
			return 0, 12036, err
		}
	}

	skinID := data.GetEquipSkinId()
	if skinID != 0 {
		if current.EquipID == 0 {
			response.Result = proto.Uint32(1)
			return client.SendMessage(12037, &response)
		}
		cache := make(map[uint32]*orm.Equipment)
		equipConfig, err := resolveEquipmentConfig(cache, current.EquipID)
		if err != nil {
			return 0, 12036, err
		}

		entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/equip_skin_template.json", fmt.Sprintf("%d", skinID))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Result = proto.Uint32(1)
				return client.SendMessage(12037, &response)
			}
			return 0, 12036, err
		}
		var template equipSkinTemplateConfig
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return 0, 12036, err
		}
		if len(template.EquipType) == 0 || !containsUint32(template.EquipType, equipConfig.Type) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(12037, &response)
		}
	}

	if current.SkinID == skinID {
		return client.SendMessage(12037, &response)
	}
	current.SkinID = skinID
	if err := orm.GormDB.Save(current).Error; err != nil {
		return 0, 12036, err
	}
	applyShipEquipmentUpdate(ship, current)
	return client.SendMessage(12037, &response)
}

func intersectsUint32(left []uint32, right []uint32) bool {
	for _, entry := range left {
		if containsUint32(right, entry) {
			return true
		}
	}
	return false
}

func buildShipEquipmentFromMemory(ownerID uint32, ship *orm.OwnedShip, pos uint32) *orm.OwnedShipEquipment {
	entry := &orm.OwnedShipEquipment{
		OwnerID: ownerID,
		ShipID:  ship.ID,
		Pos:     pos,
		EquipID: 0,
		SkinID:  0,
	}
	for i := range ship.Equipments {
		if ship.Equipments[i].Pos == pos {
			entry.EquipID = ship.Equipments[i].EquipID
			entry.SkinID = ship.Equipments[i].SkinID
			break
		}
	}
	return entry
}
