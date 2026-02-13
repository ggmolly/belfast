package answer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

const itemDataStatisticsCategory = "sharecfgdata/item_data_statistics.json"

type itemComposeConfig struct {
	ID            uint32 `json:"id"`
	ComposeNumber uint32 `json:"compose_number"`
	TargetID      uint32 `json:"target_id"`
}

func loadItemComposeConfig(itemID uint32) (*itemComposeConfig, error) {
	entry, err := orm.GetConfigEntry(itemDataStatisticsCategory, fmt.Sprintf("%d", itemID))
	if err != nil {
		if db.IsNotFound(err) {
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

func loadCommanderItemCountsTx(ctx context.Context, tx pgx.Tx, commanderID uint32, itemID uint32) (itemsCount uint32, miscCount uint32, err error) {
	var itemCount int64
	err = tx.QueryRow(ctx, `
SELECT count
FROM commander_items
WHERE commander_id = $1 AND item_id = $2
`, int64(commanderID), int64(itemID)).Scan(&itemCount)
	err = db.MapNotFound(err)
	if err != nil {
		if !db.IsNotFound(err) {
			return 0, 0, err
		}
	} else {
		itemsCount = uint32(itemCount)
	}
	var miscItemCount int64
	err = tx.QueryRow(ctx, `
SELECT data
FROM commander_misc_items
WHERE commander_id = $1 AND item_id = $2
`, int64(commanderID), int64(itemID)).Scan(&miscItemCount)
	err = db.MapNotFound(err)
	if err != nil {
		if !db.IsNotFound(err) {
			return 0, 0, err
		}
	} else {
		miscCount = uint32(miscItemCount)
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

	ctx := context.Background()
	consumeItems := uint32(0)
	consumeMisc := uint32(0)
	err = db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		itemsCount, miscCount, err := loadCommanderItemCountsTx(ctx, tx, client.Commander.CommanderID, itemID)
		if err != nil {
			return err
		}
		if uint64(itemsCount)+uint64(miscCount) < uint64(required) {
			return db.ErrNotFound
		}
		if itemsCount > 0 {
			consumeItems = uint32(math.Min(float64(itemsCount), float64(required)))
			if consumeItems > 0 {
				result, err := tx.Exec(ctx, `
UPDATE commander_items
SET count = count - $3
WHERE commander_id = $1 AND item_id = $2 AND count >= $3
`, int64(client.Commander.CommanderID), int64(itemID), int64(consumeItems))
				if err != nil {
					return err
				}
				if result.RowsAffected() == 0 {
					return db.ErrNotFound
				}
			}
		}
		remaining := required - consumeItems
		if remaining > 0 {
			consumeMisc = remaining
			result, err := tx.Exec(ctx, `
UPDATE commander_misc_items
SET data = data - $3
WHERE commander_id = $1 AND item_id = $2 AND data >= $3
`, int64(client.Commander.CommanderID), int64(itemID), int64(consumeMisc))
			if err != nil {
				return err
			}
			if result.RowsAffected() == 0 {
				return db.ErrNotFound
			}
		}
		_, err = tx.Exec(ctx, `
INSERT INTO commander_items (commander_id, item_id, count)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, item_id)
DO UPDATE SET count = commander_items.count + EXCLUDED.count
`, int64(client.Commander.CommanderID), int64(config.TargetID), int64(num))
		return err
	})
	if err != nil {
		if db.IsNotFound(err) {
			return client.SendMessage(15007, &response)
		}
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
