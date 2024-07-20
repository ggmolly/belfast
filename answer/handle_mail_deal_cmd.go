package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/consts"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

// The bool is whether the function has modified the response or not
type MailDealCmdHandler func(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) (bool, error)

func handleMailDealCmdRead(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) (bool, error) {
	err := mail.SetRead(true)
	// Put all read mails in the mailIdList
	for _, commanderMail := range client.Commander.Mails {
		if commanderMail.Read || mail.ID == commanderMail.ID {
			response.MailIdList = append(response.MailIdList, commanderMail.ID)
		}
	}
	return true, err
}

func handleMailDealCmdImportant(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) (bool, error) {
	// Put all important mails in the mailIdList
	for _, commanderMail := range client.Commander.Mails {
		if commanderMail.IsImportant || mail.ID == commanderMail.ID {
			response.MailIdList = append(response.MailIdList, commanderMail.ID)
		}
	}
	err := mail.SetImportant(true)
	return true, err
}

func handleMailDealCmdUnimportant(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) (bool, error) {
	// Put all unimportant mails in the mailIdList
	for _, commanderMail := range client.Commander.Mails {
		if !commanderMail.IsImportant || mail.ID == commanderMail.ID {
			response.MailIdList = append(response.MailIdList, commanderMail.ID)
		}
	}
	err := mail.SetImportant(false)
	return true, err
}

func handleMailDealCmdDelete(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) (bool, error) {
	for _, mail := range client.Commander.Mails {
		if mail.Read {
			response.MailIdList = append(response.MailIdList, mail.ID)
		}
	}

	if err := orm.GormDB.Delete(&orm.Mail{}, "read = ?", true).Error; err != nil {
		return false, err
	}

	// Reload mails
	if err := orm.GormDB.Preload("Attachments").Find(&client.Commander.Mails).Error; err != nil {
		return false, err
	}

	// load MailsMap
	client.Commander.MailsMap = make(map[uint32]*orm.Mail)
	for i, mail := range client.Commander.Mails {
		client.Commander.MailsMap[mail.ID] = &client.Commander.Mails[i]
	}

	return true, nil
}

func handleMailDealCmdAttachment(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) (bool, error) {
	panic("not implemented")
}

func handleMailDealCmdOverflow(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) (bool, error) {
	panic("not implemented")
}

func handleMailDealCmdMove(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mail *orm.Mail) (bool, error) {
	return false, mail.SetArchived(true)
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

	logger.LogEvent("Mail", "HandleMailDealCmd", payload.String(), logger.LOG_LEVEL_INFO)

	response := protobuf.SC_30007{
		Result:       proto.Uint32(0),
		UnreadNumber: proto.Uint32(0),
	}
	fn, ok := cmdHandlers[payload.GetCmd()]
	if !ok {
		return 0, 30006, fmt.Errorf("unknown mail deal cmd: %d", payload.GetCmd())
	}
	var mailId uint32
	matchList := payload.GetMatchList()
	if len(matchList) == 0 { // the action doesn't specifically target one / many mails
		mailId = 0
	} else if matchList[0].GetType() == 0 {
		return 0, 30006, fmt.Errorf("unhandled case: matchList[0].GetType() != 1, got %d", matchList[0].GetType())
	} else if len(matchList[0].GetArgList()) == 0 {
		return 0, 30006, fmt.Errorf("unhandled case: matchList[0].GetArgList() is empty")
	} else {
		mailId = matchList[0].GetArgList()[0]
	}
	mail, ok := client.Commander.MailsMap[mailId]
	if !ok && mailId != 0 { // 0 represents a specific case where the action doesn't target any mail
		return 0, 30006, fmt.Errorf("mail #%d not found", mailId)
	}
	dirty, err := fn(client, &payload, &response, mail)
	if err != nil {
		return 0, 30006, err
	}

	var unreadCount uint32
	for _, mail := range client.Commander.Mails {
		if !mail.Read {
			unreadCount++
		}
	}
	response.UnreadNumber = proto.Uint32(unreadCount)

	// Copy mail ids if the handler didn't do it
	if !dirty {
		var mailIds []uint32
		for _, mail := range client.Commander.Mails {
			if !mail.IsArchived {
				mailIds = append(mailIds, mail.ID)
			}
		}
		response.MailIdList = mailIds
	}

	logger.LogEvent("Mail", "HandleMailDealCmdResponse", response.String(), logger.LOG_LEVEL_INFO)
	return client.SendMessage(30007, &response)
}
