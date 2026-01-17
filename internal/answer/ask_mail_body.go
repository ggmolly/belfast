package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func AskMailBody(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_30008
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 30009, err
	}
	mail, ok := client.Commander.MailsMap[data.GetMailId()]
	if !ok {
		return 0, 30009, nil
	}
	body := protobuf.SC_30009{
		Result: proto.Uint32(0),
	}
	if err := mail.SetRead(true); err != nil {
		body.Result = proto.Uint32(1)
	}
	return client.SendMessage(30009, &body)
}
