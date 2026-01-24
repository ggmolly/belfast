package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/minigameshop"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetMiniGameShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_26150
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 26151, err
	}
	config, err := minigameshop.LoadConfig(time.Now())
	if err != nil {
		return 0, 26151, err
	}
	state, goods, err := minigameshop.RefreshIfNeeded(client.Commander.CommanderID, time.Now(), config)
	if err != nil {
		return 0, 26151, err
	}
	response := protobuf.SC_26151{
		Goods:         buildMiniGameShopGoods(goods),
		NextFlashTime: proto.Uint32(state.NextRefreshTime),
	}
	return client.SendMessage(26151, &response)
}

func buildMiniGameShopGoods(goods []orm.MiniGameShopGood) []*protobuf.GOODS_INFO_P26 {
	list := make([]*protobuf.GOODS_INFO_P26, 0, len(goods))
	for _, good := range goods {
		list = append(list, &protobuf.GOODS_INFO_P26{
			Id:    proto.Uint32(good.GoodsID),
			Count: proto.Uint32(good.Count),
		})
	}
	return list
}
