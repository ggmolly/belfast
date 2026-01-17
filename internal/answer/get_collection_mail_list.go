package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func mailToSimpleMailInfo(mail *orm.Mail) *protobuf.MAIL_SIMPLE_INFO {
	attachments := make([]*protobuf.DROPINFO, len(mail.Attachments))
	for i, attachment := range mail.Attachments {
		attachments[i] = &protobuf.DROPINFO{
			Type:   proto.Uint32(attachment.Type),
			Id:     proto.Uint32(attachment.ItemID),
			Number: proto.Uint32(attachment.Quantity),
		}
	}
	return &protobuf.MAIL_SIMPLE_INFO{
		Id:             proto.Uint32(mail.ID),
		Date:           proto.Uint32(uint32(mail.Date.Unix())),
		Title:          proto.String(mail.Title),
		Content:        proto.String(mail.Body),
		AttachmentList: attachments,
	}
}

// Returns archived mails
func GetCollectionMailList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_30004
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 30002, err
	}
	var response protobuf.SC_30005
	commanderMailsCount := uint32(len(client.Commander.Mails))
	if commanderMailsCount == 0 {
		return client.SendMessage(30003, &response)
	}

	// If end range is 0, it means we want to send all the mails
	if payload.GetIndexEnd() == 0 {
		payload.IndexEnd = proto.Uint32(commanderMailsCount + 1)
	}

	// Lua range starts at 1, so we will compensate for that
	payload.IndexBegin = proto.Uint32(payload.GetIndexBegin() - 1)

	for i := payload.GetIndexBegin(); i < commanderMailsCount && i < payload.GetIndexEnd(); i++ {
		if client.Commander.Mails[i].IsArchived {
			response.MailList = append(response.MailList, mailToSimpleMailInfo(&client.Commander.Mails[i]))
		}
	}
	return client.SendMessage(30005, &response)
}
