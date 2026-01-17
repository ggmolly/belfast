package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func DeleteArchivedMail(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_30008
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 30008, err
	}
	response := protobuf.SC_30009{
		Result: proto.Uint32(0),
	}
	mail, ok := client.Commander.MailsMap[payload.GetMailId()]
	if !ok {
		response.Result = proto.Uint32(1)
	} else if err := mail.Delete(); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(30009, &response)
}
