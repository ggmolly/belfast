package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func GiveMailAttachments(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_30004
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 30005, err
	}
	var attachments []*protobuf.ATTACHMENT
	for _, mailId := range data.GetId() {
		mail, ok := client.Commander.MailsMap[mailId]
		if !ok {
			return 0, 30005, nil
		}
		mailAttachments, err := mail.CollectAttachments(client.Commander)
		if err != nil {
			return 0, 30005, err
		}
		attachments = append(attachments, mailAttachments...)
	}
	response := protobuf.SC_30005{
		AttachmentList: attachments,
	}
	return client.SendMessage(30005, &response)
}
