package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC18002 protobuf.SC_18002

func ExerciseEnemies(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(18002, &validSC18002)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC18002)
}
