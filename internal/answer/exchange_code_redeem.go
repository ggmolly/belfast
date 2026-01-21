package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type exchangeReward struct {
	Type  uint32 `json:"type"`
	ID    uint32 `json:"id"`
	Count uint32 `json:"count"`
}

func buildExchangeAttachments(rewards []exchangeReward) ([]orm.MailAttachment, error) {
	attachments := make([]orm.MailAttachment, 0, len(rewards))
	for _, reward := range rewards {
		switch reward.Type {
		case consts.DROP_TYPE_RESOURCE, consts.DROP_TYPE_ITEM, consts.DROP_TYPE_SHIP, consts.DROP_TYPE_SKIN:
			attachments = append(attachments, orm.MailAttachment{
				Type:     reward.Type,
				ItemID:   reward.ID,
				Quantity: reward.Count,
			})
		default:
			return nil, fmt.Errorf("unsupported exchange reward type: %d", reward.Type)
		}
	}
	return attachments, nil
}

func sendExchangeFailure(client *connection.Client, response *protobuf.SC_11509, err error) (int, int, error) {
	if _, _, sendErr := client.SendMessage(11509, response); sendErr != nil {
		return 0, 11509, sendErr
	}
	return 0, 11509, err
}

func ExchangeCodeRedeem(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11508
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11509, err
	}
	response := protobuf.SC_11509{Result: proto.Uint32(1)}

	key := strings.TrimSpace(payload.GetKey())
	if key == "" {
		return client.SendMessage(11509, &response)
	}
	platform := strings.TrimSpace(payload.GetPlatform())
	normalizedKey := strings.ToUpper(key)

	var code orm.ExchangeCode
	if err := orm.GormDB.Where("upper(code) = ?", normalizedKey).First(&code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return client.SendMessage(11509, &response)
		}
		return sendExchangeFailure(client, &response, err)
	}
	if code.Platform != "" && !strings.EqualFold(code.Platform, platform) {
		return client.SendMessage(11509, &response)
	}

	var redeemed orm.ExchangeCodeRedeem
	if err := orm.GormDB.Where("exchange_code_id = ? AND commander_id = ?", code.ID, client.Commander.CommanderID).First(&redeemed).Error; err == nil {
		return client.SendMessage(11509, &response)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return sendExchangeFailure(client, &response, err)
	}

	quotaLimited := code.Quota >= 0
	if quotaLimited && code.Quota == 0 {
		return client.SendMessage(11509, &response)
	}

	var rewards []exchangeReward
	if err := json.Unmarshal(code.Rewards, &rewards); err != nil {
		return sendExchangeFailure(client, &response, err)
	}
	attachments, err := buildExchangeAttachments(rewards)
	if err != nil {
		return sendExchangeFailure(client, &response, err)
	}

	transaction := orm.GormDB.Begin()
	redeem := orm.ExchangeCodeRedeem{
		ExchangeCodeID: code.ID,
		CommanderID:    client.Commander.CommanderID,
		RedeemedAt:     time.Now(),
	}
	if err := transaction.Create(&redeem).Error; err != nil {
		transaction.Rollback()
		return sendExchangeFailure(client, &response, err)
	}
	if quotaLimited {
		code.Quota--
		if err := transaction.Save(&code).Error; err != nil {
			transaction.Rollback()
			return sendExchangeFailure(client, &response, err)
		}
	}
	mail := orm.Mail{
		Title: "Exchange Code Rewards",
		Body:  "Your exchange code rewards are attached.",
	}
	mail.Attachments = attachments
	if err := transaction.Commit().Error; err != nil {
		return sendExchangeFailure(client, &response, err)
	}
	if err := client.Commander.SendMail(&mail); err != nil {
		rollback := orm.GormDB.Begin()
		if quotaLimited {
			code.Quota++
			if err := rollback.Save(&code).Error; err != nil {
				rollback.Rollback()
				return sendExchangeFailure(client, &response, err)
			}
		}
		if err := rollback.Where("exchange_code_id = ? AND commander_id = ?", code.ID, client.Commander.CommanderID).
			Delete(&orm.ExchangeCodeRedeem{}).Error; err != nil {
			rollback.Rollback()
			return sendExchangeFailure(client, &response, err)
		}
		if err := rollback.Commit().Error; err != nil {
			return sendExchangeFailure(client, &response, err)
		}
		return sendExchangeFailure(client, &response, err)
	}
	if _, _, err := sendMailboxUpdate(client); err != nil {
		return sendExchangeFailure(client, &response, err)
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(11509, &response)
}
