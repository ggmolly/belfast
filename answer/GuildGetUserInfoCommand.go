package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC60103 protobuf.SC_60103

func GuildGetUserInfoCommand(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(60103, &validSC60103)
}

func init() {
	data := []byte{0x0a, 0x1a, 0x08, 0x00, 0x10, 0x01, 0x10, 0x02, 0x10, 0x14, 0x18, 0xf0, 0xd0, 0x98, 0x97, 0x06, 0x20, 0x89, 0x27, 0x20, 0xb9, 0x17, 0x28, 0x00, 0x30, 0x00, 0x38, 0x00}
	proto.Unmarshal(data, &validSC60103)
}
