package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func GameMailbox(buffer *[]byte, client *connection.Client) (int, int, error) {
	var unread uint32
	var total uint32
	for _, mail := range client.Commander.Mails {
		if !mail.Read {
			unread++
		}
		total++
	}
	answer := protobuf.SC_30001{
		UnreadNumber: proto.Uint32(unread),
		TotalNumber:  proto.Uint32(total),
	}
	return client.SendMessage(30001, &answer)
}
