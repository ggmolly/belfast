package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type breakoutItem struct {
	ID    uint32
	Count uint32
}

func UpgradeStar(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12027
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12027, err
	}
	response := protobuf.SC_12028{Result: proto.Uint32(1)}
	ship, ok := client.Commander.OwnedShipsMap[data.GetShipId()]
	if !ok {
		return client.SendMessage(12028, &response)
	}
	breakout, err := orm.GetShipBreakoutConfig(ship.ShipID)
	if err != nil {
		return 0, 12027, err
	}
	if breakout.BreakoutID == 0 {
		return client.SendMessage(12028, &response)
	}
	if ship.Level < breakout.Level {
		return client.SendMessage(12028, &response)
	}
	materials := data.GetMaterialIdList()
	if breakout.UseCharNum == 0 {
		if len(materials) != 0 {
			return client.SendMessage(12028, &response)
		}
	} else if len(materials) != int(breakout.UseCharNum) {
		return client.SendMessage(12028, &response)
	}
	if len(materials) > 0 {
		if ok, err := validateBreakoutMaterials(client.Commander, ship.ID, materials, breakout.UseChar); err != nil {
			return 0, 12027, err
		} else if !ok {
			return client.SendMessage(12028, &response)
		}
	}
	items, err := breakoutItems(breakout)
	if err != nil {
		return 0, 12027, err
	}
	if breakout.UseGold > 0 && !client.Commander.HasEnoughGold(breakout.UseGold) {
		return client.SendMessage(12028, &response)
	}
	if !hasEnoughBreakoutItems(client.Commander, items) {
		return client.SendMessage(12028, &response)
	}
	updatedTemplate, err := orm.GetShipTemplateConfig(breakout.BreakoutID)
	if err != nil {
		return 0, 12027, err
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(12028, &response)
	}
	if breakout.UseGold > 0 {
		if err := client.Commander.ConsumeResourceTx(tx, 1, breakout.UseGold); err != nil {
			tx.Rollback()
			return 0, 12027, err
		}
	}
	for _, item := range items {
		if item.Count == 0 {
			continue
		}
		if err := client.Commander.ConsumeItemTx(tx, item.ID, item.Count); err != nil {
			tx.Rollback()
			return 0, 12027, err
		}
	}
	if len(materials) > 0 {
		if err := consumeMaterialShips(tx, client.Commander, materials); err != nil {
			tx.Rollback()
			return 0, 12027, err
		}
	}
	if err := tx.Model(&orm.OwnedShip{}).
		Where("owner_id = ? AND id = ?", client.Commander.CommanderID, ship.ID).
		Updates(map[string]any{"ship_id": breakout.BreakoutID, "max_level": updatedTemplate.MaxLevel}).Error; err != nil {
		tx.Rollback()
		return 0, 12027, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 12027, err
	}

	ship.ShipID = breakout.BreakoutID
	ship.MaxLevel = updatedTemplate.MaxLevel
	if len(materials) > 0 {
		removeOwnedShips(client.Commander, materials)
	}
	response.Result = proto.Uint32(0)
	return client.SendMessage(12028, &response)
}

func breakoutItems(config *orm.ShipBreakoutConfig) ([]breakoutItem, error) {
	if len(config.UseItem) == 0 {
		return nil, nil
	}
	items := make([]breakoutItem, 0, len(config.UseItem))
	for _, entry := range config.UseItem {
		if len(entry) < 2 {
			return nil, errors.New("invalid breakout item entry")
		}
		items = append(items, breakoutItem{ID: entry[0], Count: entry[1]})
	}
	return items, nil
}

func hasEnoughBreakoutItems(commander *orm.Commander, items []breakoutItem) bool {
	for _, item := range items {
		if item.Count == 0 {
			continue
		}
		if !commander.HasEnoughItem(item.ID, item.Count) {
			return false
		}
	}
	return true
}

func validateBreakoutMaterials(commander *orm.Commander, shipID uint32, materialIDs []uint32, groupType uint32) (bool, error) {
	seen := make(map[uint32]struct{}, len(materialIDs))
	for _, materialID := range materialIDs {
		if materialID == shipID {
			return false, nil
		}
		if _, ok := seen[materialID]; ok {
			return false, nil
		}
		seen[materialID] = struct{}{}
		material, ok := commander.OwnedShipsMap[materialID]
		if !ok {
			return false, nil
		}
		template, err := orm.GetShipTemplateConfig(material.ShipID)
		if err != nil {
			return false, err
		}
		if template.GroupType != groupType {
			return false, nil
		}
	}
	return true, nil
}
