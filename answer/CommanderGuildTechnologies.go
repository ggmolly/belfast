package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC62101 protobuf.SC_62101

func CommanderGuildTechnologies(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(62101, &validSC62101)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC62101)
}
