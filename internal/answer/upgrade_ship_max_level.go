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

type shipMaxLevelRequirement struct {
	DropType uint32
	ID       uint32
	Count    uint32
}

func UpgradeShipMaxLevel(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12038
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12038, err
	}

	response := protobuf.SC_12039{Result: proto.Uint32(1)}
	shipID := payload.GetShipId()

	owned, ok := client.Commander.OwnedShipsMap[shipID]
	if !ok {
		return client.SendMessage(12039, &response)
	}
	if owned.Level != owned.MaxLevel {
		return client.SendMessage(12039, &response)
	}
	if owned.MaxLevel < 100 {
		return client.SendMessage(12039, &response)
	}

	nextMaxLevel, err := findNextShipMaxLevel(owned.MaxLevel)
	if err != nil {
		return 0, 12038, err
	}
	if nextMaxLevel == 0 {
		return client.SendMessage(12039, &response)
	}

	reqs, err := shipMaxLevelUpgradeRequirements(owned.MaxLevel, owned.Ship.RarityID)
	if err != nil {
		return 0, 12038, err
	}
	if len(reqs) == 0 {
		return client.SendMessage(12039, &response)
	}
	if !hasMaxLevelUpgradeRequirements(client.Commander, reqs) {
		return client.SendMessage(12039, &response)
	}

	newLevel, newExp, newSurplus, err := convertSurplusExpAfterMaxLevelIncrease(owned, nextMaxLevel)
	if err != nil {
		return 0, 12038, err
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(12039, &response)
	}

	if err := consumeMaxLevelUpgradeRequirementsTx(tx, client.Commander, reqs); err != nil {
		tx.Rollback()
		return 0, 12038, err
	}

	updates := map[string]any{
		"max_level":   nextMaxLevel,
		"level":       newLevel,
		"exp":         newExp,
		"surplus_exp": newSurplus,
	}
	if err := tx.Model(&orm.OwnedShip{}).
		Where("owner_id = ? AND id = ?", client.Commander.CommanderID, owned.ID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return 0, 12038, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 12038, err
	}

	owned.MaxLevel = nextMaxLevel
	owned.Level = newLevel
	owned.Exp = newExp
	owned.SurplusExp = newSurplus
	response.Result = proto.Uint32(0)
	return client.SendMessage(12039, &response)
}

func hasMaxLevelUpgradeRequirements(commander *orm.Commander, reqs []shipMaxLevelRequirement) bool {
	for _, req := range reqs {
		if req.Count == 0 {
			continue
		}
		switch req.DropType {
		case 1:
			if !commander.HasEnoughResource(req.ID, req.Count) {
				return false
			}
		case 2:
			if !commander.HasEnoughItem(req.ID, req.Count) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func consumeMaxLevelUpgradeRequirementsTx(tx *gorm.DB, commander *orm.Commander, reqs []shipMaxLevelRequirement) error {
	for _, req := range reqs {
		if req.Count == 0 {
			continue
		}
		switch req.DropType {
		case 1:
			if err := commander.ConsumeResourceTx(tx, req.ID, req.Count); err != nil {
				return err
			}
		case 2:
			if err := commander.ConsumeItemTx(tx, req.ID, req.Count); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported requirement type %d", req.DropType)
		}
	}
	return nil
}

func findNextShipMaxLevel(currentMaxLevel uint32) (uint32, error) {
	for level := currentMaxLevel + 1; level <= 200; level++ {
		data, err := loadShipLevelConfigRaw(level)
		if err != nil {
			return 0, err
		}
		if data == nil {
			return 0, nil
		}
		var entry struct {
			LevelLimit uint32 `json:"level_limit"`
		}
		if err := json.Unmarshal(data, &entry); err != nil {
			return 0, err
		}
		if entry.LevelLimit == 1 {
			return level, nil
		}
	}
	return 0, nil
}

func shipMaxLevelUpgradeRequirements(currentMaxLevel uint32, rarity uint32) ([]shipMaxLevelRequirement, error) {
	data, err := loadShipLevelConfigRaw(currentMaxLevel)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	key := fmt.Sprintf("need_item_rarity%d", rarity)
	raw, ok := root[key]
	if !ok {
		return nil, nil
	}
	var tuples [][]uint32
	if err := json.Unmarshal(raw, &tuples); err != nil {
		return nil, err
	}

	reqs := make([]shipMaxLevelRequirement, 0, len(tuples))
	for _, tuple := range tuples {
		if len(tuple) < 3 {
			return nil, errors.New("invalid ship_level need_item_rarity tuple")
		}
		req := shipMaxLevelRequirement{DropType: tuple[0], ID: tuple[1], Count: tuple[2]}
		reqs = append(reqs, req)
	}
	return reqs, nil
}

func loadShipLevelConfigRaw(level uint32) (json.RawMessage, error) {
	if level == 0 {
		return nil, nil
	}
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/ship_level.json", fmt.Sprintf("%d", level))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return entry.Data, nil
}

func convertSurplusExpAfterMaxLevelIncrease(owned *orm.OwnedShip, nextMaxLevel uint32) (uint32, uint32, uint32, error) {
	level := owned.Level
	exp := owned.Exp + owned.SurplusExp
	surplus := uint32(0)

	for level < nextMaxLevel {
		config, err := loadShipLevelConfig(level)
		if err != nil {
			return 0, 0, 0, err
		}
		if config == nil {
			break
		}
		required := config.Exp
		if owned.Ship.RarityID == 6 {
			required = config.ExpUR
		}
		if required == 0 || exp < required {
			break
		}
		exp -= required
		level++
	}

	if level >= nextMaxLevel && nextMaxLevel >= 100 && exp > 0 {
		surplus = addSurplusExp(surplus, exp)
		exp = 0
	}

	return level, exp, surplus, nil
}
