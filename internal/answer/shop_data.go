package answer

import (
	"encoding/json"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type monthShopTemplate struct {
	CoreShopGoods       []uint32 `json:"core_shop_goods"`
	BlueprintShopGoods  []uint32 `json:"blueprint_shop_goods"`
	BlueprintShopLimit  []uint32 `json:"blueprint_shop_limit_goods"`
	HonorMedalShopGoods []uint32 `json:"honormedal_shop_goods"`
	BlueprintShopLimit2 []uint32 `json:"blueprint_shop_limit_goods_2"`
	BlueprintShopGoods2 []uint32 `json:"blueprint_shop_goods_2"`
	BlueprintShopLimit3 []uint32 `json:"blueprint_shop_limit_goods_3"`
	BlueprintShopGoods3 []uint32 `json:"blueprint_shop_goods_3"`
	BlueprintShopGoods4 []uint32 `json:"blueprint_shop_goods_4"`
	BlueprintShopLimit4 []uint32 `json:"blueprint_shop_limit_goods_4"`
}

func ShopData(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_16200{
		Month: proto.Uint32(uint32(time.Now().Month())),
	}
	monthKey := uint32(time.Now().Year()*100 + int(time.Now().Month()))
	counts, err := orm.ListMonthShopPurchaseCounts(client.Commander.CommanderID, monthKey)
	if err != nil {
		return 0, 16200, err
	}
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/month_shop_template.json")
	if err != nil {
		return 0, 16200, err
	}
	if len(entries) == 0 {
		return client.SendMessage(16200, &response)
	}
	var template monthShopTemplate
	if err := json.Unmarshal(entries[0].Data, &template); err != nil {
		return 0, 16200, err
	}
	blueprints := append([]uint32{}, template.BlueprintShopGoods...)
	blueprints = append(blueprints, template.BlueprintShopLimit...)
	blueprints = append(blueprints, template.BlueprintShopGoods2...)
	blueprints = append(blueprints, template.BlueprintShopLimit2...)
	blueprints = append(blueprints, template.BlueprintShopGoods3...)
	blueprints = append(blueprints, template.BlueprintShopLimit3...)
	blueprints = append(blueprints, template.BlueprintShopGoods4...)
	blueprints = append(blueprints, template.BlueprintShopLimit4...)
	response.CoreShopList = buildShopInfoList(template.CoreShopGoods, counts)
	response.BlueShopList = buildShopInfoList(blueprints, counts)
	response.NormalShopList = buildShopInfoList(template.HonorMedalShopGoods, counts)
	return client.SendMessage(16200, &response)
}

func buildShopInfoList(ids []uint32, counts map[uint32]uint32) []*protobuf.SHOPINFO {
	if len(ids) == 0 {
		return nil
	}
	entries := make([]*protobuf.SHOPINFO, len(ids))
	for i, id := range ids {
		payCount := uint32(0)
		if counts != nil {
			if count, ok := counts[id]; ok {
				payCount = count
			}
		}
		entries[i] = &protobuf.SHOPINFO{
			ShopId:   proto.Uint32(id),
			PayCount: proto.Uint32(payCount),
		}
	}
	return entries
}
