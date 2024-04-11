package answer

import (
	"log"

	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC17204 protobuf.SC_17204

func UNK_17204(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17203
	err := proto.Unmarshal((*buffer), &payload)
	if err != nil {
		return 0, 17204, err
	}
	switch payload.GetType() {
	case 0: // idk yet

	case 29: // mailbox

	}
	// client asked for type=0 -> On login ?
	// client asked for type=29 -> Mailbox
	log.Println("Client asked for type:", payload.GetType())
	return client.SendMessage(17204, &validSC17204)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC17204)
}
