package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func OwnedItems(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_15001
	for _, item := range client.Commander.Items {
		response.ItemList = append(response.ItemList, &protobuf.ITEMINFO{
			Id:    proto.Uint32(item.ItemID),
			Count: proto.Uint32(item.Count),
		})
	}
	// for _, item := range ownedLimitItems {
	// 	response.LimitList = append(response.LimitList, &protobuf.ITEMINFO{
	// 		Id:    proto.Uint32(item.Item.ID),
	// 		Count: proto.Uint32(item.Count),
	// 	})
	// }
	for _, item := range client.Commander.MiscItems {
		response.ItemMiscList = append(response.ItemMiscList, &protobuf.ITEMMISC{
			Id:   proto.Uint32(item.ItemID),
			Data: proto.Uint32(item.Data),
		})
	}
	return client.SendMessage(15001, &response)
}
