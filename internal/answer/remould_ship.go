package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func RemouldShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12011
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12011, err
	}
	response := protobuf.SC_12012{Result: proto.Uint32(1)}
	ship, ok := client.Commander.OwnedShipsMap[data.GetShipId()]
	if !ok {
		return client.SendMessage(12012, &response)
	}
	config, err := orm.GetTransformDataTemplate(data.GetRemouldId())
	if err != nil {
		return 0, 12011, err
	}
	targetTemplateID, ok := findRemouldTarget(config.ShipID, ship.ShipID)
	if !ok {
		return client.SendMessage(12012, &response)
	}
	currentLevel := findTransformLevel(ship.Transforms, config.ID)
	if currentLevel >= config.MaxLevel {
		return client.SendMessage(12012, &response)
	}
	if ship.Level < config.LevelLimit {
		return client.SendMessage(12012, &response)
	}
	if ship.Ship.Star < config.StarLimit {
		return client.SendMessage(12012, &response)
	}
	if !client.Commander.HasEnoughGold(config.UseGold) {
		return client.SendMessage(12012, &response)
	}
	items, err := remouldItemsForNextLevel(config, currentLevel)
	if err != nil {
		return 0, 12011, err
	}
	for _, item := range items {
		if !client.Commander.HasEnoughItem(item.ID, item.Count) {
			return client.SendMessage(12012, &response)
		}
	}
	if !remouldPrerequisitesMet(ship.Transforms, config.ConditionID) {
		return client.SendMessage(12012, &response)
	}
	materialIDs := data.GetMaterialId()
	if config.UseShip == 0 {
		if len(materialIDs) != 0 {
			return client.SendMessage(12012, &response)
		}
	} else {
		if len(materialIDs) != int(config.UseShip) {
			return client.SendMessage(12012, &response)
		}
		for _, materialID := range materialIDs {
			if materialID == ship.ID {
				return client.SendMessage(12012, &response)
			}
			if _, ok := client.Commander.OwnedShipsMap[materialID]; !ok {
				return client.SendMessage(12012, &response)
			}
		}
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(12012, &response)
	}
	if err := client.Commander.ConsumeResourceTx(tx, 1, config.UseGold); err != nil {
		tx.Rollback()
		return 0, 12011, err
	}
	for _, item := range items {
		if err := client.Commander.ConsumeItemTx(tx, item.ID, item.Count); err != nil {
			tx.Rollback()
			return 0, 12011, err
		}
	}
	if err := orm.UpsertOwnedShipTransformTx(tx, &orm.OwnedShipTransform{
		OwnerID:     client.Commander.CommanderID,
		ShipID:      ship.ID,
		TransformID: config.ID,
		Level:       currentLevel + 1,
	}); err != nil {
		tx.Rollback()
		return 0, 12011, err
	}
	if err := orm.DeleteOwnedShipTransformsTx(tx, client.Commander.CommanderID, ship.ID, config.EditTrans); err != nil {
		tx.Rollback()
		return 0, 12011, err
	}
	update := map[string]any{
		"ship_id": targetTemplateID,
	}
	if config.SkinID != 0 {
		update["skin_id"] = config.SkinID
		if err := client.Commander.GiveSkinTx(tx, config.SkinID); err != nil {
			tx.Rollback()
			return 0, 12011, err
		}
	}
	if err := tx.Model(&orm.OwnedShip{}).
		Where("owner_id = ? AND id = ?", client.Commander.CommanderID, ship.ID).
		Updates(update).Error; err != nil {
		tx.Rollback()
		return 0, 12011, err
	}
	if err := consumeMaterialShips(tx, client.Commander, materialIDs); err != nil {
		tx.Rollback()
		return 0, 12011, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 12011, err
	}

	applyTransformUpdate(ship, config.ID, currentLevel+1)
	ship.ShipID = targetTemplateID
	if config.SkinID != 0 {
		ship.SkinID = config.SkinID
	}
	removeTransforms(ship, config.EditTrans)
	if len(materialIDs) > 0 {
		removeOwnedShips(client.Commander, materialIDs)
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(12012, &response)
}

type remouldItem struct {
	ID    uint32
	Count uint32
}

func remouldItemsForNextLevel(config *orm.TransformDataTemplate, currentLevel uint32) ([]remouldItem, error) {
	raw := config.UseItemsForLevel(currentLevel + 1)
	if len(raw) == 0 {
		return nil, nil
	}
	items := make([]remouldItem, 0, len(raw))
	for _, entry := range raw {
		if len(entry) < 2 {
			return nil, errors.New("invalid remould item entry")
		}
		items = append(items, remouldItem{ID: entry[0], Count: entry[1]})
	}
	return items, nil
}

func remouldPrerequisitesMet(transforms []orm.OwnedShipTransform, prereqIDs []uint32) bool {
	if len(prereqIDs) == 0 {
		return true
	}
	for _, prereqID := range prereqIDs {
		level := findTransformLevel(transforms, prereqID)
		prereqConfig, err := orm.GetTransformDataTemplate(prereqID)
		if err != nil {
			return false
		}
		if level != prereqConfig.MaxLevel {
			return false
		}
	}
	return true
}

func findRemouldTarget(options [][]uint32, current uint32) (uint32, bool) {
	for _, entry := range options {
		if len(entry) < 2 {
			continue
		}
		if entry[0] == current {
			return entry[1], true
		}
	}
	return 0, false
}

func findTransformLevel(transforms []orm.OwnedShipTransform, transformID uint32) uint32 {
	for _, entry := range transforms {
		if entry.TransformID == transformID {
			return entry.Level
		}
	}
	return 0
}

func applyTransformUpdate(ship *orm.OwnedShip, transformID uint32, level uint32) {
	for i := range ship.Transforms {
		if ship.Transforms[i].TransformID == transformID {
			ship.Transforms[i].Level = level
			return
		}
	}
	ship.Transforms = append(ship.Transforms, orm.OwnedShipTransform{
		OwnerID:     ship.OwnerID,
		ShipID:      ship.ID,
		TransformID: transformID,
		Level:       level,
	})
}

func removeTransforms(ship *orm.OwnedShip, transformIDs []uint32) {
	if len(transformIDs) == 0 {
		return
	}
	kept := ship.Transforms[:0]
	for _, entry := range ship.Transforms {
		if !containsTransform(transformIDs, entry.TransformID) {
			kept = append(kept, entry)
		}
	}
	ship.Transforms = kept
}

func containsTransform(ids []uint32, target uint32) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func consumeMaterialShips(tx *gorm.DB, commander *orm.Commander, materialIDs []uint32) error {
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

func removeOwnedShips(commander *orm.Commander, shipIDs []uint32) {
	if len(shipIDs) == 0 {
		return
	}
	kept := commander.Ships[:0]
	toRemove := make(map[uint32]struct{}, len(shipIDs))
	for _, id := range shipIDs {
		toRemove[id] = struct{}{}
	}
	for i := range commander.Ships {
		ship := commander.Ships[i]
		if _, ok := toRemove[ship.ID]; ok {
			delete(commander.OwnedShipsMap, ship.ID)
			continue
		}
		kept = append(kept, ship)
	}
	commander.Ships = kept
}
