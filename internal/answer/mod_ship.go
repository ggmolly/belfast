package answer

import (
	"errors"
	"math"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var shipModStrengthIDs = []uint32{2, 3, 4, 5, 6}

func ModShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12017
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12017, err
	}
	response := protobuf.SC_12018{Result: proto.Uint32(1)}
	ship, ok := client.Commander.OwnedShipsMap[data.GetShipId()]
	if !ok {
		return client.SendMessage(12018, &response)
	}
	materialIDs := data.GetMaterialIdList()
	if len(materialIDs) == 0 {
		return client.SendMessage(12018, &response)
	}
	materials, ok := collectMaterialShips(client.Commander, ship.ID, materialIDs)
	if !ok {
		return client.SendMessage(12018, &response)
	}
	strengths, err := orm.ListOwnedShipStrengths(orm.GormDB, client.Commander.CommanderID, ship.ID)
	if err != nil {
		return 0, 12017, err
	}
	ship.Strengths = strengths
	shipTemplate, err := orm.GetShipTemplateConfig(ship.ShipID)
	if err != nil {
		return 0, 12017, err
	}
	strengthenConfig, err := orm.GetShipStrengthenConfig(shipTemplate.StrengthenID)
	if err != nil {
		return 0, 12017, err
	}
	additions, err := shipModAdditions(shipTemplate.GroupType, materials)
	if err != nil {
		return 0, 12017, err
	}
	updates, err := shipModStrengthUpdates(ship, strengthenConfig, additions)
	if err != nil {
		return 0, 12017, err
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(12018, &response)
	}
	for strengthID, exp := range updates {
		entry := orm.OwnedShipStrength{
			OwnerID:    client.Commander.CommanderID,
			ShipID:     ship.ID,
			StrengthID: strengthID,
			Exp:        exp,
		}
		if err := orm.UpsertOwnedShipStrengthTx(tx, &entry); err != nil {
			tx.Rollback()
			return 0, 12017, err
		}
	}
	if err := consumeModMaterialShips(tx, client.Commander, materialIDs); err != nil {
		tx.Rollback()
		return 0, 12017, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 12017, err
	}

	applyStrengthUpdates(ship, updates)
	removeOwnedShips(client.Commander, materialIDs)
	response.Result = proto.Uint32(0)
	return client.SendMessage(12018, &response)
}

func collectMaterialShips(commander *orm.Commander, shipID uint32, materialIDs []uint32) ([]*orm.OwnedShip, bool) {
	materials := make([]*orm.OwnedShip, 0, len(materialIDs))
	for _, materialID := range materialIDs {
		if materialID == shipID {
			return nil, false
		}
		material, ok := commander.OwnedShipsMap[materialID]
		if !ok {
			return nil, false
		}
		materials = append(materials, material)
	}
	return materials, true
}

func shipModAdditions(targetGroup uint32, materials []*orm.OwnedShip) (map[uint32]uint32, error) {
	additions := make(map[uint32]uint32, len(shipModStrengthIDs))
	for _, material := range materials {
		template, err := orm.GetShipTemplateConfig(material.ShipID)
		if err != nil {
			return nil, err
		}
		strengthen, err := orm.GetShipStrengthenConfig(template.StrengthenID)
		if err != nil {
			return nil, err
		}
		attrExp, err := shipModAttrExp(strengthen)
		if err != nil {
			return nil, err
		}
		for _, strengthID := range shipModStrengthIDs {
			addition := attrExp[strengthID]
			if template.GroupType == targetGroup {
				addition *= 2
			}
			additions[strengthID] += addition
		}
	}
	return additions, nil
}

func shipModAttrExp(config *orm.ShipStrengthenConfig) (map[uint32]uint32, error) {
	if len(config.AttrExp) < len(shipModStrengthIDs) {
		return nil, errors.New("ship strengthen attr_exp is incomplete")
	}
	attrExp := make(map[uint32]uint32, len(shipModStrengthIDs))
	for _, strengthID := range shipModStrengthIDs {
		index := int(strengthID - 2)
		attrExp[strengthID] = config.AttrExp[index]
	}
	return attrExp, nil
}

func shipModStrengthUpdates(ship *orm.OwnedShip, config *orm.ShipStrengthenConfig, additions map[uint32]uint32) (map[uint32]uint32, error) {
	if len(config.Durability) < len(shipModStrengthIDs) || len(config.LevelExp) < len(shipModStrengthIDs) {
		return nil, errors.New("ship strengthen config is incomplete")
	}
	current := make(map[uint32]uint32, len(ship.Strengths))
	for _, entry := range ship.Strengths {
		current[entry.StrengthID] = entry.Exp
	}
	updates := make(map[uint32]uint32, len(shipModStrengthIDs))
	for _, strengthID := range shipModStrengthIDs {
		addition := additions[strengthID]
		if addition == 0 {
			continue
		}
		index := int(strengthID - 2)
		expRatio := config.LevelExp[index]
		if expRatio == 0 {
			expRatio = 1
		}
		topLimit := shipModTopLimit(ship.Level, config.Durability[index])
		if topLimit == 0 {
			continue
		}
		cap := topLimit * expRatio
		newExp := current[strengthID] + addition
		if newExp > cap {
			newExp = cap
		}
		updates[strengthID] = newExp
	}
	return updates, nil
}

func shipModTopLimit(level uint32, durability uint32) uint32 {
	if durability == 0 {
		return 0
	}
	levelValue := float64(level)
	if levelValue > 100 {
		levelValue = 100
	}
	factor := 3 + 7*levelValue/100
	value := factor * float64(durability) * 0.1
	return uint32(math.Floor(value))
}

func consumeModMaterialShips(tx *gorm.DB, commander *orm.Commander, materialIDs []uint32) error {
	for _, materialID := range materialIDs {
		material, ok := commander.OwnedShipsMap[materialID]
		if !ok {
			return errors.New("material ship not found")
		}
		entries, err := orm.ListOwnedShipEquipment(tx, commander.CommanderID, material.ID)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.EquipID == 0 {
				continue
			}
			if err := commander.AddOwnedEquipmentTx(tx, entry.EquipID, 1); err != nil {
				return err
			}
		}
		if err := tx.Where("owner_id = ? AND ship_id = ?", commander.CommanderID, material.ID).Delete(&orm.OwnedShipEquipment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_id = ? AND id = ?", commander.CommanderID, material.ID).Delete(&orm.OwnedShip{}).Error; err != nil {
			return err
		}
	}
	return nil
}

func applyStrengthUpdates(ship *orm.OwnedShip, updates map[uint32]uint32) {
	if len(updates) == 0 {
		return
	}
	for i := range ship.Strengths {
		if exp, ok := updates[ship.Strengths[i].StrengthID]; ok {
			ship.Strengths[i].Exp = exp
			delete(updates, ship.Strengths[i].StrengthID)
		}
	}
	if len(updates) == 0 {
		return
	}
	for strengthID, exp := range updates {
		ship.Strengths = append(ship.Strengths, orm.OwnedShipStrength{
			OwnerID:    ship.OwnerID,
			ShipID:     ship.ID,
			StrengthID: strengthID,
			Exp:        exp,
		})
	}
}
