package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func AskMailBody(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_30008
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 30009, err
	}
	mail, ok := client.Commander.MailsMap[data.GetId()]
	if !ok {
		return 0, 30009, nil
	}
	body := protobuf.SC_30009{
		DetailInfo: &protobuf.MAIL_DETAIL{
			Id:      proto.Uint32(mail.ID),
			Content: proto.String(mail.Body),
		},
	}
	mail.SetRead(true)
	return client.SendMessage(30009, &body)
}
