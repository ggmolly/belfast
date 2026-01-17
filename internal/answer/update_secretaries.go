package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateSecretaries(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_11011
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 11012, err
	}
	response := protobuf.SC_11012{
		Result: proto.Uint32(0),
	}

	// Check if all ships are owned by the player
	for _, ship := range data.GetCharacter() {
		if _, ok := client.Commander.OwnedShipsMap[ship]; !ok {
			response.Result = proto.Uint32(1)
			break
		}
	}

	if *response.Result == 0 {
		if err := client.Commander.RemoveSecretaries(); err != nil {
			response.Result = proto.Uint32(1)
		} else if err := client.Commander.UpdateSecretaries(data.GetCharacter()); err != nil { // Update secretaries
			response.Result = proto.Uint32(1)
		}
	}

	return client.SendMessage(11012, &response)
}
