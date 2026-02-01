package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChangeManifesto(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11009
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11010, err
	}
	response := protobuf.SC_11010{Result: proto.Uint32(0)}
	manifesto := payload.GetAdv()
	if err := orm.GormDB.Model(client.Commander).Update("manifesto", manifesto).Error; err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11010, &response)
	}
	client.Commander.Manifesto = manifesto
	return client.SendMessage(11010, &response)
}
