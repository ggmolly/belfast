package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/arenashop"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetArenaShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_18100
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 18101, err
	}
	config, err := arenashop.LoadConfig()
	if err != nil {
		return 0, 18101, err
	}
	state, err := arenashop.RefreshIfNeeded(client.Commander.CommanderID, time.Now())
	if err != nil {
		return 0, 18101, err
	}
	shopList := arenashop.BuildShopList(state.FlashCount, config)
	response := protobuf.SC_18101{
		FlashCount:    proto.Uint32(state.FlashCount),
		ArenaShopList: shopList,
		NextFlashTime: proto.Uint32(state.NextFlashTime),
	}
	return client.SendMessage(18101, &response)
}

func RefreshArenaShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_18102
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 18103, err
	}
	config, err := arenashop.LoadConfig()
	if err != nil {
		return 0, 18103, err
	}
	state, err := arenashop.RefreshIfNeeded(client.Commander.CommanderID, time.Now())
	if err != nil {
		return 0, 18103, err
	}
	response := protobuf.SC_18103{
		Result: proto.Uint32(0),
	}
	nextFlashCount := int(state.FlashCount + 1)
	if nextFlashCount > len(config.Template.RefreshPrice) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(18103, &response)
	}
	refreshCost := config.Template.RefreshPrice[nextFlashCount-1]
	if refreshCost > 0 && !client.Commander.HasEnoughResource(4, refreshCost) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(18103, &response)
	}
	_, shopList, cost, err := arenashop.RefreshShop(client.Commander.CommanderID, time.Now(), config)
	if err != nil {
		return 0, 18103, err
	}
	if cost > 0 {
		if err := client.Commander.ConsumeResource(4, cost); err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(18103, &response)
		}
	}
	response.ArenaShopList = shopList
	return client.SendMessage(18103, &response)
}
