package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const itemDataStatisticsCategory = "sharecfgdata/item_data_statistics.json"

type itemComposeConfig struct {
	ID            uint32 `json:"id"`
	ComposeNumber uint32 `json:"compose_number"`
	TargetID      uint32 `json:"target_id"`
}

func loadItemComposeConfig(itemID uint32) (*itemComposeConfig, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, itemDataStatisticsCategory, fmt.Sprintf("%d", itemID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var parsed itemComposeConfig
	if err := json.Unmarshal(entry.Data, &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

func loadCommanderItemCountsTx(tx *gorm.DB, commanderID uint32, itemID uint32) (itemsCount uint32, miscCount uint32, err error) {
	var item orm.CommanderItem
	itemResult := tx.Where("commander_id = ? AND item_id = ?", commanderID, itemID).First(&item)
	if itemResult.Error != nil {
		if !errors.Is(itemResult.Error, gorm.ErrRecordNotFound) {
			return 0, 0, itemResult.Error
		}
	} else {
		itemsCount = item.Count
	}

	var misc orm.CommanderMiscItem
	miscResult := tx.Where("commander_id = ? AND item_id = ?", commanderID, itemID).First(&misc)
	if miscResult.Error != nil {
		if !errors.Is(miscResult.Error, gorm.ErrRecordNotFound) {
			return 0, 0, miscResult.Error
		}
	} else {
		miscCount = misc.Data
	}

	return itemsCount, miscCount, nil
}

func ComposeItem(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_15006
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 15007, err
	}

	response := protobuf.SC_15007{Result: proto.Uint32(1)}
	itemID := payload.GetId()
	num := payload.GetNum()
	if itemID == 0 || num == 0 {
		return client.SendMessage(15007, &response)
	}

	if client.Commander.CommanderItemsMap == nil && client.Commander.MiscItemsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return 0, 15007, err
		}
	}

	config, err := loadItemComposeConfig(itemID)
	if err != nil {
		return 0, 15007, err
	}
	if config == nil || config.ComposeNumber == 0 || config.TargetID == 0 {
		return client.SendMessage(15007, &response)
	}

	required64 := uint64(num) * uint64(config.ComposeNumber)
	if required64 == 0 || required64 > math.MaxUint32 {
		return client.SendMessage(15007, &response)
	}
	required := uint32(required64)

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(15007, &response)
	}

	itemsCount, miscCount, err := loadCommanderItemCountsTx(tx, client.Commander.CommanderID, itemID)
	if err != nil {
		tx.Rollback()
		return 0, 15007, err
	}
	if uint64(itemsCount)+uint64(miscCount) < uint64(required) {
		tx.Rollback()
		return client.SendMessage(15007, &response)
	}

	consumeItems := uint32(0)
	if itemsCount > 0 {
		consumeItems = uint32(math.Min(float64(itemsCount), float64(required)))
		if consumeItems > 0 {
			result := tx.Model(&orm.CommanderItem{}).
				Where("commander_id = ? AND item_id = ? AND count >= ?", client.Commander.CommanderID, itemID, consumeItems).
				Update("count", gorm.Expr("count - ?", consumeItems))
			if result.Error != nil {
				tx.Rollback()
				return 0, 15007, result.Error
			}
			if result.RowsAffected == 0 {
				tx.Rollback()
				return client.SendMessage(15007, &response)
			}
		}
	}

	remaining := required - consumeItems
	consumeMisc := uint32(0)
	if remaining > 0 {
		consumeMisc = remaining
		result := tx.Model(&orm.CommanderMiscItem{}).
			Where("commander_id = ? AND item_id = ? AND data >= ?", client.Commander.CommanderID, itemID, consumeMisc).
			Update("data", gorm.Expr("data - ?", consumeMisc))
		if result.Error != nil {
			tx.Rollback()
			return 0, 15007, result.Error
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return client.SendMessage(15007, &response)
		}
	}

	grant := orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: config.TargetID, Count: num}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "item_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"count": gorm.Expr("count + ?", num),
		}),
	}).Create(&grant).Error; err != nil {
		tx.Rollback()
		return 0, 15007, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, 15007, err
	}

	if consumeItems > 0 {
		if entry, ok := client.Commander.CommanderItemsMap[itemID]; ok {
			if entry.Count >= consumeItems {
				entry.Count -= consumeItems
			} else {
				entry.Count = 0
			}
		}
	}
	if consumeMisc > 0 {
		if entry, ok := client.Commander.MiscItemsMap[itemID]; ok {
			if entry.Data >= consumeMisc {
				entry.Data -= consumeMisc
			} else {
				entry.Data = 0
			}
		}
	}
	if entry, ok := client.Commander.CommanderItemsMap[config.TargetID]; ok {
		entry.Count += num
	} else {
		stored := orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: config.TargetID, Count: num}
		client.Commander.Items = append(client.Commander.Items, stored)
		client.Commander.CommanderItemsMap[config.TargetID] = &client.Commander.Items[len(client.Commander.Items)-1]
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(15007, &response)
}
