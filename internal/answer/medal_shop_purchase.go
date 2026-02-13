package answer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

const (
	medalShopCurrencyItemID = uint32(15001)

	medalShopPurchaseResultOK           = uint32(0)
	medalShopPurchaseResultInvalid      = uint32(1)
	medalShopPurchaseResultInsufficient = uint32(2)
	medalShopPurchaseResultStock        = uint32(3)
	medalShopPurchaseResultStale        = uint32(4)
	medalShopPurchaseResultUnsupported  = uint32(5)
	medalShopPurchaseResultDBError      = uint32(6)
)

type honorMedalGoodsListEntry struct {
	Group     uint32   `json:"group"`
	Price     uint32   `json:"price"`
	Goods     []uint32 `json:"goods"`
	GoodsType uint32   `json:"goods_type"`
	Num       uint32   `json:"num"`
	IsShip    uint32   `json:"is_ship"`
}

func MedalShopPurchase(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_16108
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 16109, err
	}

	response := protobuf.SC_16109{Result: proto.Uint32(medalShopPurchaseResultInvalid)}
	shopID := payload.GetShopid()
	if shopID == 0 || payload.GetFlashTime() == 0 {
		return client.SendMessage(16109, &response)
	}

	selected := payload.GetSelected()
	if len(selected) == 0 {
		return client.SendMessage(16109, &response)
	}

	entry, ok, err := loadHonorMedalGoodsListEntry(shopID)
	if err != nil {
		return 0, 16109, err
	}
	if !ok {
		return client.SendMessage(16109, &response)
	}

	if entry.Num == 0 || len(entry.Goods) == 0 {
		return client.SendMessage(16109, &response)
	}
	if entry.GoodsType == 1 && len(selected) != 1 {
		return client.SendMessage(16109, &response)
	}

	totalUnits := uint32(0)
	rewards := map[uint32]uint32{}
	for _, pick := range selected {
		id := pick.GetId()
		count := pick.GetCount()
		if id == 0 || count == 0 {
			return client.SendMessage(16109, &response)
		}
		if !containsUint32(entry.Goods, id) {
			return client.SendMessage(16109, &response)
		}
		totalUnits += count
		rewards[id] += count
	}
	if totalUnits == 0 {
		return client.SendMessage(16109, &response)
	}

	totalCost := entry.Price * totalUnits

	var dropType uint32
	if entry.IsShip != 0 {
		dropType = consts.DROP_TYPE_SHIP
	} else {
		dropType = consts.DROP_TYPE_ITEM
	}

	errInvalid := errors.New("invalid")
	errStale := errors.New("stale")
	errInsufficient := errors.New("insufficient")
	errStock := errors.New("stock")
	errUnsupported := errors.New("unsupported")

	commanderID := client.Commander.CommanderID
	ctx := context.Background()
	err = orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		var nextRefreshTime uint32
		if err := tx.QueryRow(ctx, `
SELECT next_refresh_time
FROM medal_shop_states
WHERE commander_id = $1
`, int64(commanderID)).Scan(&nextRefreshTime); err != nil {
			err = db.MapNotFound(err)
			if db.IsNotFound(err) {
				return errInvalid
			}
			return err
		}
		if nextRefreshTime != payload.GetFlashTime() {
			return errStale
		}

		var goodIndex uint32
		var goodCount uint32
		if err := tx.QueryRow(ctx, `
SELECT "index", count
FROM medal_shop_goods
WHERE commander_id = $1 AND goods_id = $2
`, int64(commanderID), int64(shopID)).Scan(&goodIndex, &goodCount); err != nil {
			err = db.MapNotFound(err)
			if db.IsNotFound(err) {
				return errInvalid
			}
			return err
		}
		if goodCount < totalUnits {
			return errStock
		}

		if !client.Commander.HasEnoughItem(medalShopCurrencyItemID, totalCost) {
			return errInsufficient
		}
		if err := client.Commander.ConsumeItemTx(ctx, tx, medalShopCurrencyItemID, totalCost); err != nil {
			return errInsufficient
		}

		res, err := tx.Exec(ctx, `
UPDATE medal_shop_goods
SET count = count - $3
WHERE commander_id = $1 AND "index" = $2 AND count >= $3
`, int64(commanderID), int64(goodIndex), int64(totalUnits))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return errStock
		}

		drops := make([]*protobuf.DROPINFO, 0, len(rewards))
		for id, units := range rewards {
			rewardAmount := entry.Num * units
			switch dropType {
			case consts.DROP_TYPE_ITEM:
				if err := client.Commander.AddItemTx(ctx, tx, id, rewardAmount); err != nil {
					return err
				}
			case consts.DROP_TYPE_SHIP:
				for i := uint32(0); i < rewardAmount; i++ {
					if _, err := client.Commander.AddShipTx(ctx, tx, id); err != nil {
						return err
					}
				}
			default:
				return errUnsupported
			}
			drops = append(drops, &protobuf.DROPINFO{
				Type:   proto.Uint32(dropType),
				Id:     proto.Uint32(id),
				Number: proto.Uint32(rewardAmount),
			})
		}
		response.DropList = drops
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, errInvalid):
			response.Result = proto.Uint32(medalShopPurchaseResultInvalid)
		case errors.Is(err, errStale):
			response.Result = proto.Uint32(medalShopPurchaseResultStale)
		case errors.Is(err, errInsufficient):
			response.Result = proto.Uint32(medalShopPurchaseResultInsufficient)
		case errors.Is(err, errStock):
			response.Result = proto.Uint32(medalShopPurchaseResultStock)
		case errors.Is(err, errUnsupported):
			response.Result = proto.Uint32(medalShopPurchaseResultUnsupported)
		default:
			response.Result = proto.Uint32(medalShopPurchaseResultDBError)
		}
		response.DropList = nil
		return client.SendMessage(16109, &response)
	}

	response.Result = proto.Uint32(medalShopPurchaseResultOK)
	return client.SendMessage(16109, &response)
}

func loadHonorMedalGoodsListEntry(shopGroupID uint32) (*honorMedalGoodsListEntry, bool, error) {
	entries, err := orm.ListConfigEntries("ShareCfg/honormedal_goods_list.json")
	if err != nil {
		return nil, false, err
	}
	for _, entry := range entries {
		var list []honorMedalGoodsListEntry
		if err := json.Unmarshal(entry.Data, &list); err == nil && len(list) > 0 {
			for i := range list {
				if list[i].Group == shopGroupID {
					return &list[i], true, nil
				}
			}
			continue
		}
		var single honorMedalGoodsListEntry
		if err := json.Unmarshal(entry.Data, &single); err == nil {
			if single.Group == shopGroupID {
				return &single, true, nil
			}
		}
	}
	return nil, false, nil
}
