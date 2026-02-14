package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func mailboxCounts(mails []orm.Mail) (uint32, uint32) {
	var unread uint32
	var total uint32
	for _, mail := range mails {
		if mail.IsArchived {
			continue
		}
		if !mail.Read {
			unread++
		}
		total++
	}
	return unread, total
}

func syncCommanderMailState(client *connection.Client) error {
	totalCount, unreadCount, err := orm.GetMailboxCounts(client.Commander.CommanderID)
	if err != nil {
		return err
	}

	unread, total := mailboxCounts(client.Commander.Mails)
	if unread == unreadCount && total == totalCount {
		return nil
	}

	return client.Commander.Load()
}

func sendMailboxUpdate(client *connection.Client) (int, int, error) {
	unread, total := mailboxCounts(client.Commander.Mails)
	answer := protobuf.SC_30001{
		UnreadNumber: proto.Uint32(unread),
		TotalNumber:  proto.Uint32(total),
	}
	return client.SendMessage(30001, &answer)
}

func GameMailbox(buffer *[]byte, client *connection.Client) (int, int, error) {
	return sendMailboxUpdate(client)
}
