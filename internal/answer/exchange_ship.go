package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ExchangeShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12047
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12048, err
	}
	response := protobuf.SC_12048{
		Result: proto.Uint32(1),
	}
	if client.Commander.ExchangeCount < 400 {
		return client.SendMessage(12048, &response)
	}

	if data.GetShipTid() != 105171 && data.GetShipTid() != 307081 {
		response.Result = proto.Uint32(2)
		return client.SendMessage(12048, &response)
	}

	client.Commander.ExchangeCount -= 400
	if _, err := client.Commander.AddShip(data.GetShipTid()); err != nil {
		response.Result = proto.Uint32(3)
		return client.SendMessage(12048, &response)
	}
	response.Result = proto.Uint32(0)
	response.DropList = []*protobuf.DROPINFO{
		{
			Type:   proto.Uint32(consts.DROP_TYPE_SHIP),
			Id:     data.ShipTid,
			Number: proto.Uint32(1),
		},
	}
	if err := orm.GormDB.Save(&client.Commander).Error; err != nil {
		response.Result = proto.Uint32(3)
		return client.SendMessage(12048, &response)
	}
	return client.SendMessage(12048, &response)
}
