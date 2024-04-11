package answer

import (
	"fmt"

	"github.com/bettercallmolly/belfast/connection"
	"github.com/bettercallmolly/belfast/logger"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC25027 protobuf.SC_25027

// A get with a type?
func UNK_25027(buffer *[]byte, client *connection.Client) (int, int, error) {
	var packet protobuf.CS_25026
	err := proto.Unmarshal(*buffer, &packet)
	if err != nil {
		return 0, 25027, err
	}
	logger.LogEvent("Client", "CS_25026", fmt.Sprintf("client asked for type=%d", packet.GetType()), logger.LOG_LEVEL_DEBUG)
	// Answer with default valid SC_25027
	return client.SendMessage(25027, &validSC25027)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC25027)
}
