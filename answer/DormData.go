package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func DormData(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_19001 // Send an empty DormData
	response.Lv = proto.Uint32(0)
	response.Food = proto.Uint32(0)
	response.FoodMaxIncrease = proto.Uint32(0)
	response.FoodMaxIncreaseCount = proto.Uint32(0)
	response.FloorNum = proto.Uint32(0)
	response.ExpPos = proto.Uint32(0)
	response.NextTimestamp = proto.Uint32(0)
	response.LoadExp = proto.Uint32(0)
	response.LoadFood = proto.Uint32(0)
	response.LoadTime = proto.Uint32(0)
	response.Name = proto.String("")
	return client.SendMessage(19001, &response)
}
