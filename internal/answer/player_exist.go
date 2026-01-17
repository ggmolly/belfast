package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

// Check if an account exists on the game server
func PlayerExist(buffer *[]byte, client *connection.Client) (int, int, error) {
	var protoData protobuf.CS_10026
	err := proto.Unmarshal((*buffer), &protoData)
	if err != nil {
		return 0, 10027, err
	}

	response := protobuf.SC_10027{}

	err = client.GetCommander(protoData.GetAccountId())
	if err != nil {
		// Player not found?
		response.UserId = proto.Uint32(0)
		response.Level = proto.Uint32(0)
	} else {
		// Even if the player is punished, we still return player information
		response.UserId = proto.Uint32(client.Commander.CommanderID)
		response.Level = proto.Uint32(uint32(client.Commander.Level))
	}

	return client.SendMessage(10027, &response)
}
