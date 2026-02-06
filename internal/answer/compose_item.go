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

	available, _, err := getCommanderItemCountFromDB(client, itemID)
	if err != nil {
		return 0, 15007, err
	}
	if available < required {
		return client.SendMessage(15007, &response)
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(15007, &response)
	}

	consumedFromItems := false
	itemResult := tx.Model(&orm.CommanderItem{}).
		Where("commander_id = ? AND item_id = ? AND count >= ?", client.Commander.CommanderID, itemID, required).
		Update("count", gorm.Expr("count - ?", required))
	if itemResult.Error != nil {
		tx.Rollback()
		return 0, 15007, itemResult.Error
	}
	if itemResult.RowsAffected > 0 {
		consumedFromItems = true
	} else {
		miscResult := tx.Model(&orm.CommanderMiscItem{}).
			Where("commander_id = ? AND item_id = ? AND data >= ?", client.Commander.CommanderID, itemID, required).
			Update("data", gorm.Expr("data - ?", required))
		if miscResult.Error != nil {
			tx.Rollback()
			return 0, 15007, miscResult.Error
		}
		if miscResult.RowsAffected == 0 {
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

	if consumedFromItems {
		if entry, ok := client.Commander.CommanderItemsMap[itemID]; ok {
			if entry.Count >= required {
				entry.Count -= required
			} else {
				entry.Count = 0
			}
		}
	} else {
		if entry, ok := client.Commander.MiscItemsMap[itemID]; ok {
			if entry.Data >= required {
				entry.Data -= required
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
