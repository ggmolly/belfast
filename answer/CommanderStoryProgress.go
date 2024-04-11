package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC13001 protobuf.SC_13001

func CommanderStoryProgress(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(13001, &validSC13001)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC13001)
}
