package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/medalshop"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetMedalShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_16106
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 16107, err
	}
	config, err := medalshop.LoadConfig()
	if err != nil {
		return 0, 16107, err
	}
	state, goods, err := medalshop.RefreshIfNeeded(client.Commander.CommanderID, time.Now(), config)
	if err != nil {
		return 0, 16107, err
	}
	response := protobuf.SC_16107{
		Result:        proto.Uint32(0),
		ItemFlashTime: proto.Uint32(state.NextRefreshTime),
		GoodList:      buildMedalShopGoods(goods),
	}
	return client.SendMessage(16107, &response)
}

func buildMedalShopGoods(goods []orm.MedalShopGood) []*protobuf.GOODS_INFO_P16 {
	list := make([]*protobuf.GOODS_INFO_P16, 0, len(goods))
	for _, good := range goods {
		list = append(list, &protobuf.GOODS_INFO_P16{
			Id:    proto.Uint32(good.GoodsID),
			Count: proto.Uint32(good.Count),
		})
	}
	return list
}
