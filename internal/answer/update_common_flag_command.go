package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateCommonFlagCommand(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11019
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11020, err
	}
	response := protobuf.SC_11020{
		Result: proto.Uint32(0),
	}
	if err := orm.SetCommanderCommonFlag(orm.GormDB, client.Commander.CommanderID, payload.GetFlagId()); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11020, &response)
}
