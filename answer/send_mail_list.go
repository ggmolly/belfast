package answer

import (
	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/orm"
	"google.golang.org/protobuf/proto"

	"github.com/ggmolly/belfast/protobuf"
)

func mailToMailInfo(mail *orm.Mail) *protobuf.MAIL_INFO {
	attachments := make([]*protobuf.DROPINFO, len(mail.Attachments))
	for i, attachment := range mail.Attachments {
		attachments[i] = &protobuf.DROPINFO{
			Type:   proto.Uint32(attachment.Type),
			Id:     proto.Uint32(attachment.ItemID),
			Number: proto.Uint32(attachment.Quantity),
		}
	}
	fullTitle := mail.Title
	if mail.CustomSender != nil {
		fullTitle += "||" + *mail.CustomSender
	}
	return &protobuf.MAIL_INFO{
		Id:             proto.Uint32(mail.ID),
		Date:           proto.Uint32(uint32(mail.Date.Unix())),
		Title:          proto.String(fullTitle),
		Content:        proto.String(mail.Body),
		AttachmentList: attachments,
		ImpFlag:        proto.Uint32(boolToUint32(mail.IsImportant)),
		ReadFlag:       proto.Uint32(boolToUint32(mail.Read)),
		AttachFlag:     proto.Uint32(boolToUint32(mail.AttachmentsCollected)),
	}
}

func SendMailList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_30002
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 30002, err
	}
	var response protobuf.SC_30003
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
		if !client.Commander.Mails[i].IsArchived {
			response.MailList = append(response.MailList, mailToMailInfo(&client.Commander.Mails[i]))
		}
	}
	return client.SendMessage(30003, &response)
}
