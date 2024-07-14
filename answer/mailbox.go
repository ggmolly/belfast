package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func GameMailbox(buffer *[]byte, client *connection.Client) (int, int, error) {
	var unread uint32
	var total uint32
	for _, mail := range client.Commander.Mails {
		if mail.IsArchived {
			continue
		}
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
