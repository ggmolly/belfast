package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChatRoomChange(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_11401
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 11402, err
	}
	client.Server.ChangeRoom(client.Commander.RoomID, data.GetRoomId(), client)
	client.Commander.UpdateRoom(data.GetRoomId())

	response := protobuf.SC_11402{
		Result: proto.Uint32(0),
		RoomId: data.RoomId,
	}
	return client.SendMessage(11402, &response)
}
