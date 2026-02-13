package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
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
	client.Commander.Manifesto = manifesto
	if err := client.Commander.Commit(); err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11010, &response)
	}
	return client.SendMessage(11010, &response)
}
