package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"google.golang.org/protobuf/proto"
)

func GetMetaProgress(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(63315, &validSC63315)
}

func init() {
	data := []byte{0x08, 0x01, 0x10, 0xc0, 0xac, 0xd0, 0x04}
	proto.Unmarshal(data, &validSC63315)
}
