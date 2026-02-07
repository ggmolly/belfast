package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type activityShopTemplateEntry struct {
	ID               uint32 `json:"id"`
	ResourceCategory uint32 `json:"resource_category"`
	ResourceType     uint32 `json:"resource_type"`
	ResourceNum      uint32 `json:"resource_num"`
	CommodityType    uint32 `json:"commodity_type"`
	CommodityID      uint32 `json:"commodity_id"`
	Num              uint32 `json:"num"`
	NumLimit         uint32 `json:"num_limit"`
}

func MonthShopPurchase(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_16201
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 16202, err
	}

	response := protobuf.SC_16202{Result: proto.Uint32(0)}

	if payload.GetCount() == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(16202, &response)
	}

	monthShop, ok, err := loadMonthShopTemplate()
	if err != nil {
		return 0, 16202, err
	}
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(16202, &response)
	}
	allowedIDs := monthShopIDsByType(monthShop, payload.GetType())
	if len(allowedIDs) == 0 || !containsUint32(allowedIDs, payload.GetId()) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(16202, &response)
	}

	good, ok, err := loadActivityShopEntry(payload.GetId())
	if err != nil {
		return 0, 16202, err
	}
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(16202, &response)
	}

	totalCost64 := uint64(good.ResourceNum) * uint64(payload.GetCount())
	if totalCost64 > math.MaxUint32 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(16202, &response)
	}
	totalCost := uint32(totalCost64)

	rewardAmount64 := uint64(good.Num) * uint64(payload.GetCount())
	if rewardAmount64 > math.MaxUint32 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(16202, &response)
	}
	rewardAmount := uint32(rewardAmount64)

	response.DropList = []*protobuf.DROPINFO{buildDrop(good.CommodityType, good.CommodityID, rewardAmount)}
	if response.DropList[0] == nil {
		response.Result = proto.Uint32(4)
		response.DropList = nil
		return client.SendMessage(16202, &response)
	}

	const (
		errInvalid      = 1
		errInsufficient = 2
		errLimit        = 3
		errUnsupported  = 4
		errDBError      = 5
	)

	sentinelInsufficient := errors.New("insufficient")
	sentinelLimit := errors.New("limit")
	sentinelUnsupported := errors.New("unsupported")

	monthKey := uint32(time.Now().Year()*100 + int(time.Now().Month()))
	commanderID := client.Commander.CommanderID
	err = orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if good.NumLimit > 0 {
			current, err := orm.GetMonthShopPurchaseCountTx(tx, commanderID, payload.GetId(), monthKey)
			if err != nil {
				return err
			}
			if current+payload.GetCount() > good.NumLimit {
				return sentinelLimit
			}
		}

		switch good.ResourceCategory {
		case 1:
			if !client.Commander.HasEnoughResource(good.ResourceType, totalCost) {
				return sentinelInsufficient
			}
			if err := client.Commander.ConsumeResourceTx(tx, good.ResourceType, totalCost); err != nil {
				return err
			}
		case 2:
			if !client.Commander.HasEnoughItem(good.ResourceType, totalCost) {
				return sentinelInsufficient
			}
			if err := client.Commander.ConsumeItemTx(tx, good.ResourceType, totalCost); err != nil {
				return err
			}
		default:
			return sentinelUnsupported
		}

		switch good.CommodityType {
		case 1:
			if err := client.Commander.AddResourceTx(tx, good.CommodityID, rewardAmount); err != nil {
				return err
			}
		case 2:
			if err := client.Commander.AddItemTx(tx, good.CommodityID, rewardAmount); err != nil {
				return err
			}
		case 6:
			if rewardAmount != 1 {
				return sentinelUnsupported
			}
			if err := client.Commander.GiveSkinTx(tx, good.CommodityID); err != nil {
				return err
			}
		default:
			return sentinelUnsupported
		}

		return orm.IncrementMonthShopPurchaseTx(tx, commanderID, payload.GetId(), monthKey, payload.GetCount())
	})
	if err != nil {
		switch {
		case errors.Is(err, sentinelInsufficient):
			response.Result = proto.Uint32(errInsufficient)
		case errors.Is(err, sentinelLimit):
			response.Result = proto.Uint32(errLimit)
		case errors.Is(err, sentinelUnsupported):
			response.Result = proto.Uint32(errUnsupported)
		default:
			response.Result = proto.Uint32(errDBError)
		}
		response.DropList = nil
		return client.SendMessage(16202, &response)
	}

	return client.SendMessage(16202, &response)
}

func buildDrop(typ uint32, id uint32, amount uint32) *protobuf.DROPINFO {
	switch typ {
	case 1, 2, 4, 6:
		return &protobuf.DROPINFO{Type: proto.Uint32(typ), Id: proto.Uint32(id), Number: proto.Uint32(amount)}
	default:
		return nil
	}
}

func loadMonthShopTemplate() (*monthShopTemplate, bool, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/month_shop_template.json")
	if err != nil {
		return nil, false, err
	}
	if len(entries) == 0 {
		return nil, false, nil
	}
	var out monthShopTemplate
	if err := json.Unmarshal(entries[0].Data, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func monthShopIDsByType(template *monthShopTemplate, typ uint32) []uint32 {
	switch typ {
	case 1:
		return template.CoreShopGoods
	case 2:
		ids := append([]uint32{}, template.BlueprintShopGoods...)
		ids = append(ids, template.BlueprintShopLimit...)
		ids = append(ids, template.BlueprintShopGoods2...)
		ids = append(ids, template.BlueprintShopLimit2...)
		ids = append(ids, template.BlueprintShopGoods3...)
		ids = append(ids, template.BlueprintShopLimit3...)
		ids = append(ids, template.BlueprintShopGoods4...)
		ids = append(ids, template.BlueprintShopLimit4...)
		return ids
	case 3:
		return template.HonorMedalShopGoods
	default:
		return nil
	}
}

func loadActivityShopEntry(id uint32) (*activityShopTemplateEntry, bool, error) {
	key := fmt.Sprintf("%d", id)
	if entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/activity_shop_template.json", key); err == nil {
		var out activityShopTemplateEntry
		if err := json.Unmarshal(entry.Data, &out); err != nil {
			return nil, false, err
		}
		return &out, true, nil
	}
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/activity_shop_template.json")
	if err != nil {
		return nil, false, err
	}
	for i := range entries {
		var out activityShopTemplateEntry
		if err := json.Unmarshal(entries[i].Data, &out); err != nil {
			return nil, false, err
		}
		if out.ID == id {
			return &out, true, nil
		}
	}
	return nil, false, nil
}
