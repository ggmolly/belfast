package answer

import (
	"sort"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/shopstreet"
	"google.golang.org/protobuf/proto"
)

func GetShopStreet(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_22101
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 22102, err
	}
	state, goods, err := shopstreet.RefreshIfNeeded(client.Commander.CommanderID, time.Now())
	if err != nil {
		return 0, 22102, err
	}
	response := protobuf.SC_22102{
		Street: buildShoppingStreetProto(state, goods),
	}
	return client.SendMessage(22102, &response)
}

func buildShoppingStreetProto(state *orm.ShoppingStreetState, goods []orm.ShoppingStreetGood) *protobuf.SHOPPINGSTREET {
	sort.Slice(goods, func(i, j int) bool {
		return goods[i].GoodsID < goods[j].GoodsID
	})
	goodsList := make([]*protobuf.STREETGOODS, 0, len(goods))
	for _, good := range goods {
		goodsList = append(goodsList, &protobuf.STREETGOODS{
			GoodsId:  proto.Uint32(good.GoodsID),
			Discount: proto.Uint32(good.Discount),
			BuyCount: proto.Uint32(good.BuyCount),
		})
	}
	return &protobuf.SHOPPINGSTREET{
		Lv:            proto.Uint32(state.Level),
		NextFlashTime: proto.Uint32(state.NextFlashTime),
		LvUpTime:      proto.Uint32(state.LevelUpTime),
		GoodsList:     goodsList,
		FlashCount:    proto.Uint32(state.FlashCount),
	}
}
