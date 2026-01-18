package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	mailMatchTypeIDs       = 1
	mailMatchTypeResources = 2
	mailMatchTypeItems     = 3
)

// The bool is whether the function has modified the response or not
type MailDealCmdHandler func(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mails []*orm.Mail) (bool, error)

func selectMailTargets(client *connection.Client, matchList []*protobuf.MATCH_EXPRESSION) ([]*orm.Mail, error) {
	if len(matchList) == 0 {
		mails := make([]*orm.Mail, 0, len(client.Commander.Mails))
		for i := range client.Commander.Mails {
			mail := &client.Commander.Mails[i]
			if mail.IsArchived {
				continue
			}
			mails = append(mails, mail)
		}
		return mails, nil
	}

	idSet := map[uint32]struct{}{}
	resourceSet := map[uint32]struct{}{}
	itemSet := map[uint32]struct{}{}

	for _, expr := range matchList {
		switch expr.GetType() {
		case mailMatchTypeIDs:
			for _, id := range expr.GetArgList() {
				if id == 0 {
					continue
				}
				idSet[id] = struct{}{}
			}
		case mailMatchTypeResources:
			for _, id := range expr.GetArgList() {
				if id == 0 {
					continue
				}
				resourceSet[id] = struct{}{}
			}
		case mailMatchTypeItems:
			for _, id := range expr.GetArgList() {
				if id == 0 {
					continue
				}
				itemSet[id] = struct{}{}
			}
		default:
			return nil, fmt.Errorf("unknown match expression type: %d", expr.GetType())
		}
	}

	matchesByDrop := func(mail *orm.Mail) bool {
		if len(resourceSet) == 0 && len(itemSet) == 0 {
			return false
		}
		for _, attachment := range mail.Attachments {
			switch attachment.Type {
			case 1:
				if _, ok := resourceSet[attachment.ItemID]; ok {
					return true
				}
			case 2:
				if _, ok := itemSet[attachment.ItemID]; ok {
					return true
				}
			}
		}
		return false
	}

	seen := make(map[uint32]struct{})
	mails := make([]*orm.Mail, 0, len(client.Commander.Mails))
	for i := range client.Commander.Mails {
		mail := &client.Commander.Mails[i]
		if mail.IsArchived {
			continue
		}
		_, matchedByID := idSet[mail.ID]
		matchedByDrop := !matchedByID && matchesByDrop(mail)
		if !matchedByID && !matchedByDrop {
			continue
		}
		if _, ok := seen[mail.ID]; ok {
			continue
		}
		seen[mail.ID] = struct{}{}
		mails = append(mails, mail)
	}
	return mails, nil
}

func mailAttachmentToDropInfo(attachment orm.MailAttachment) *protobuf.DROPINFO {
	return &protobuf.DROPINFO{
		Type:   proto.Uint32(attachment.Type),
		Id:     proto.Uint32(attachment.ItemID),
		Number: proto.Uint32(attachment.Quantity),
	}
}

type dropKey struct {
	typeID uint32
	itemID uint32
}

func mergeDropInfos(drops []*protobuf.DROPINFO) []*protobuf.DROPINFO {
	if len(drops) == 0 {
		return nil
	}
	merged := make(map[dropKey]*protobuf.DROPINFO)
	order := make([]dropKey, 0, len(drops))
	for _, drop := range drops {
		key := dropKey{typeID: drop.GetType(), itemID: drop.GetId()}
		existing, ok := merged[key]
		if ok {
			existing.Number = proto.Uint32(existing.GetNumber() + drop.GetNumber())
			continue
		}
		merged[key] = &protobuf.DROPINFO{
			Type:   proto.Uint32(drop.GetType()),
			Id:     proto.Uint32(drop.GetId()),
			Number: proto.Uint32(drop.GetNumber()),
		}
		order = append(order, key)
	}
	result := make([]*protobuf.DROPINFO, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	return result
}

func handleMailDealCmdRead(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mails []*orm.Mail) (bool, error) {
	mailIds := make([]uint32, 0, len(mails))
	for _, mail := range mails {
		if !mail.Read {
			if err := mail.SetRead(true); err != nil {
				return true, err
			}
		}
		mailIds = append(mailIds, mail.ID)
	}
	response.MailIdList = mailIds
	return true, nil
}

func handleMailDealCmdImportant(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mails []*orm.Mail) (bool, error) {
	mailIds := make([]uint32, 0, len(mails))
	for _, mail := range mails {
		if !mail.IsImportant {
			if err := mail.SetImportant(true); err != nil {
				return true, err
			}
		}
		mailIds = append(mailIds, mail.ID)
	}
	response.MailIdList = mailIds
	return true, nil
}

func handleMailDealCmdUnimportant(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mails []*orm.Mail) (bool, error) {
	mailIds := make([]uint32, 0, len(mails))
	for _, mail := range mails {
		if mail.IsImportant {
			if err := mail.SetImportant(false); err != nil {
				return true, err
			}
		}
		mailIds = append(mailIds, mail.ID)
	}
	response.MailIdList = mailIds
	return true, nil
}

func handleMailDealCmdDelete(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mails []*orm.Mail) (bool, error) {
	mailIds := make([]uint32, 0, len(mails))
	for _, mail := range mails {
		if mail.Read && (mail.AttachmentsCollected || len(mail.Attachments) == 0) {
			mailIds = append(mailIds, mail.ID)
		}
	}

	if len(mailIds) == 0 {
		return true, nil
	}

	if err := orm.GormDB.Where("receiver_id = ?", client.Commander.CommanderID).Where("id IN ?", mailIds).Delete(&orm.Mail{}).Error; err != nil {
		return false, err
	}

	if err := orm.GormDB.Preload("Attachments").Find(&client.Commander.Mails).Error; err != nil {
		return false, err
	}

	client.Commander.MailsMap = make(map[uint32]*orm.Mail)
	for i, mail := range client.Commander.Mails {
		client.Commander.MailsMap[mail.ID] = &client.Commander.Mails[i]
	}

	response.MailIdList = mailIds
	return true, nil
}

func collectAttachmentDrops(client *connection.Client, mails []*orm.Mail, apply bool) ([]uint32, []*protobuf.DROPINFO, error) {
	var mailIds []uint32
	drops := []*protobuf.DROPINFO{}
	for _, mail := range mails {
		if len(mail.Attachments) == 0 {
			continue
		}
		if mail.AttachmentsCollected {
			continue
		}
		if apply {
			mail.Read = true
			attachments, err := mail.CollectAttachments(client.Commander)
			if err != nil {
				return nil, nil, err
			}
			for _, attachment := range attachments {
				drops = append(drops, mailAttachmentToDropInfo(attachment))
			}
		} else {
			for _, attachment := range mail.Attachments {
				drops = append(drops, mailAttachmentToDropInfo(attachment))
			}
		}
		mailIds = append(mailIds, mail.ID)
	}
	return mailIds, mergeDropInfos(drops), nil
}

func handleMailDealCmdAttachment(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mails []*orm.Mail) (bool, error) {
	mailIds, drops, err := collectAttachmentDrops(client, mails, true)
	if err != nil {
		return true, err
	}
	response.MailIdList = mailIds
	response.DropList = drops
	return true, nil
}

func handleMailDealCmdOverflow(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mails []*orm.Mail) (bool, error) {
	mailIds, drops, err := collectAttachmentDrops(client, mails, false)
	if err != nil {
		return true, err
	}
	response.MailIdList = mailIds
	response.DropList = drops
	return true, nil
}

func handleMailDealCmdMove(client *connection.Client, payload *protobuf.CS_30006, response *protobuf.SC_30007, mails []*orm.Mail) (bool, error) {
	mailIds := make([]uint32, 0, len(mails))
	for _, mail := range mails {
		if !mail.IsArchived {
			if err := mail.SetArchived(true); err != nil {
				return true, err
			}
		}
		mailIds = append(mailIds, mail.ID)
	}
	response.MailIdList = mailIds
	return true, nil
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

	response := protobuf.SC_30007{
		Result:       proto.Uint32(0),
		UnreadNumber: proto.Uint32(0),
	}
	fn, ok := cmdHandlers[payload.GetCmd()]
	if !ok {
		return 0, 30006, fmt.Errorf("unknown mail deal cmd: %d", payload.GetCmd())
	}
	matchList := payload.GetMatchList()
	mails, err := selectMailTargets(client, matchList)
	if err != nil {
		return 0, 30006, err
	}
	if len(mails) == 0 {
		var unreadCount uint32
		for _, mail := range client.Commander.Mails {
			if !mail.Read {
				unreadCount++
			}
		}
		response.UnreadNumber = proto.Uint32(unreadCount)
		return client.SendMessage(30007, &response)
	}
	dirty, err := fn(client, &payload, &response, mails)
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
	return client.SendMessage(30007, &response)
}
