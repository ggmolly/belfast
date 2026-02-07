package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func MonthShopFlag(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_16203
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 16204, err
	}
	response := protobuf.SC_16204{Ret: proto.Uint32(0)}
	if err := orm.SetCommanderCommonFlag(orm.GormDB, client.Commander.CommanderID, payload.GetFlag()); err != nil {
		response.Ret = proto.Uint32(1)
	}
	return client.SendMessage(16204, &response)
}
