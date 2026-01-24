package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CancelCommonFlagCommand(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11021
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11022, err
	}
	response := protobuf.SC_11022{Result: proto.Uint32(0)}
	if err := orm.ClearCommanderCommonFlag(orm.GormDB, client.Commander.CommanderID, payload.GetFlagId()); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11022, &response)
}
