package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/guildshop"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	guildShopGetShop       = 0
	guildShopAutoRefresh   = 1
	guildShopManualRefresh = 2
)

func GetGuildShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_60033
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 60034, err
	}
	config, err := guildshop.LoadConfig()
	if err != nil {
		return 0, 60034, err
	}
	state, goods, err := guildshop.RefreshIfNeeded(client.Commander.CommanderID, time.Now(), config)
	if err != nil {
		return 0, 60034, err
	}
	result := uint32(0)
	requestType := payload.GetType()
	if requestType == guildShopManualRefresh {
		if state.RefreshCount > 0 {
			result = 1
		} else if config.ResetCost > 0 && !client.Commander.HasEnoughResource(8, config.ResetCost) {
			result = 1
		} else {
			goods, err = guildshop.RefreshGoods(client.Commander.CommanderID, time.Now(), config, guildshop.RefreshOptions{
				RefreshCount:    1,
				NextRefreshTime: state.NextRefreshTime,
			})
			if err != nil {
				return 0, 60034, err
			}
			if config.ResetCost > 0 {
				if err := client.Commander.ConsumeResource(8, config.ResetCost); err != nil {
					result = 1
				}
			}
			state, goods, err = guildshop.RefreshIfNeeded(client.Commander.CommanderID, time.Now(), config)
			if err != nil {
				return 0, 60034, err
			}
		}
	} else if requestType != guildShopGetShop && requestType != guildShopAutoRefresh {
		result = 1
	}
	response := protobuf.SC_60034{
		Result: proto.Uint32(result),
		Info: &protobuf.SHOP_INFO{
			RefreshCount:    proto.Uint32(state.RefreshCount),
			NextRefreshTime: proto.Uint32(state.NextRefreshTime),
			GoodList:        buildGuildShopGoods(goods),
		},
	}
	return client.SendMessage(60034, &response)
}

func buildGuildShopGoods(goods []orm.GuildShopGood) []*protobuf.GOODS_INFO_P60 {
	list := make([]*protobuf.GOODS_INFO_P60, 0, len(goods))
	for _, good := range goods {
		list = append(list, &protobuf.GOODS_INFO_P60{
			Id:    proto.Uint32(good.GoodsID),
			Count: proto.Uint32(good.Count),
			Index: proto.Uint32(good.Index),
		})
	}
	return list
}
