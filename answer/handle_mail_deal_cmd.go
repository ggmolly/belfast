package answer

import (
	"fmt"
	"log"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/consts"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

type MailDealCmdHandler func(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) error

func handleMailDealCmdRead(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) error {
	return mail.SetRead(true)
}

func handleMailDealCmdImportant(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) error {
	return mail.SetImportant(true)
}

func handleMailDealCmdUnimportant(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) error {
	return mail.SetImportant(false)
}

func handleMailDealCmdDelete(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) error {
	if err := orm.GormDB.Delete(&orm.Mail{}, "read = ?", true).Error; err != nil {
		return err
	}

	// Reload mails
	if err := orm.GormDB.Preload("Attachments").Find(&client.Commander.Mails).Error; err != nil {
		log.Println("!, found mails:", len(client.Commander.Mails))
		return err
	}

	// load MailsMap
	mail.Commander.MailsMap = make(map[uint32]*orm.Mail)
	for i, mail := range client.Commander.Mails {
		mail.Commander.MailsMap[mail.ID] = &client.Commander.Mails[i]
	}
	return nil
}

func handleMailDealCmdAttachment(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) error {
	panic("not implemented")
}

func handleMailDealCmdOverflow(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) error {
	panic("not implemented")
}

func handleMailDealCmdMove(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) error {
	return mail.SetArchived(true)
}

var cmdHandlers = map[uint32]MailDealCmdHandler{
	consts.MAIL_DEAL_CMDS_READ:        handleMailDealCmdRead,
	consts.MAIL_DEAL_CMDS_IMPORTANT:   handleMailDealCmdImportant,
	consts.MAIL_DEAL_CMDS_UNIMPORTANT: handleMailDealCmdUnimportant,
	consts.MAIL_DEAL_CMDS_DELETE:      handleMailDealCmdDelete,
	consts.MAIL_DEAL_CMDS_ATTACHMENT:  handleMailDealCmdAttachment,
	consts.MAIL_DEAL_CMDS_OVERFLOW:    handleMailDealCmdOverflow,
	consts.MAIL_DEAL_CMDS_MOVE:        handleMailDealCmdMove,
}

func HandleMailDealCmd(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_30006
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 30006, err
	}

	logger.LogEvent("Server", "MailCmd", payload.String(), logger.LOG_LEVEL_DEBUG)

	response := protobuf.SC_30007{
		Result: proto.Uint32(0),
	}
	_, ok := cmdHandlers[payload.GetCmd()]
	if !ok {
		return 0, 30006, fmt.Errorf("unknown mail deal cmd: %d", payload.GetCmd())
	}
	var mailIndex uint32
	if len(payload.GetMatchList()) > 0 {
		mailIndex = payload.GetMatchList()[0].GetType() - 1
	}
	if mailIndex >= uint32(len(client.Commander.Mails)) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(30007, &response)
	}
	fn := cmdHandlers[payload.GetCmd()]
	mail := client.Commander.Mails[mailIndex]
	if err := fn(client, &payload, &response, &mail); err != nil {
		return 0, 30006, err
	}

	var unreadCount uint32
	for _, mail := range client.Commander.Mails {
		if !mail.Read {
			unreadCount++
		}
	}
	response.UnreadNumber = proto.Uint32(unreadCount)

	// Copy mail ids
	var mailIds []uint32
	for _, mail := range client.Commander.Mails {
		if !mail.IsArchived {
			mailIds = append(mailIds, mail.ID)
		}
	}
	response.MailIdList = mailIds
	log.Println(response.String())
	return client.SendMessage(30007, &response)
}
