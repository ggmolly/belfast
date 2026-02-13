package answer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
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

	ctx := context.Background()
	var code orm.ExchangeCode
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, code, platform, quota, rewards
FROM exchange_codes
WHERE upper(code) = $1
`, normalizedKey).Scan(&code.ID, &code.Code, &code.Platform, &code.Quota, &code.Rewards)
	err = db.MapNotFound(err)
	if err != nil {
		if db.IsNotFound(err) {
			return client.SendMessage(11509, &response)
		}
		return sendExchangeFailure(client, &response, err)
	}
	if code.Platform != "" && !strings.EqualFold(code.Platform, platform) {
		return client.SendMessage(11509, &response)
	}

	var redeemedExists bool
	err = db.DefaultStore.Pool.QueryRow(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM exchange_code_redeems
  WHERE exchange_code_id = $1
    AND commander_id = $2
)
`, int64(code.ID), int64(client.Commander.CommanderID)).Scan(&redeemedExists)
	if err != nil {
		return sendExchangeFailure(client, &response, err)
	}
	if redeemedExists {
		return client.SendMessage(11509, &response)
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

	errQuotaDepleted := errors.New("exchange code quota depleted")
	err = orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
INSERT INTO exchange_code_redeems (exchange_code_id, commander_id, redeemed_at)
VALUES ($1, $2, $3)
`, int64(code.ID), int64(client.Commander.CommanderID), time.Now().UTC()); err != nil {
			return err
		}
		if !quotaLimited {
			return nil
		}
		res, err := tx.Exec(ctx, `
UPDATE exchange_codes
SET quota = quota - 1
WHERE id = $1
  AND quota > 0
`, int64(code.ID))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return errQuotaDepleted
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errQuotaDepleted) {
			return client.SendMessage(11509, &response)
		}
		return sendExchangeFailure(client, &response, err)
	}
	if quotaLimited {
		code.Quota--
	}

	mail := orm.Mail{
		Title: "Exchange Code Rewards",
		Body:  "Your exchange code rewards are attached.",
	}
	mail.Attachments = attachments
	if err := client.Commander.SendMail(&mail); err != nil {
		rollbackErr := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
			if quotaLimited {
				if _, err := tx.Exec(ctx, `
UPDATE exchange_codes
SET quota = quota + 1
WHERE id = $1
`, int64(code.ID)); err != nil {
					return err
				}
			}
			if _, err := tx.Exec(ctx, `
DELETE FROM exchange_code_redeems
WHERE exchange_code_id = $1
  AND commander_id = $2
`, int64(code.ID), int64(client.Commander.CommanderID)); err != nil {
				return err
			}
			return nil
		})
		if rollbackErr != nil {
			return sendExchangeFailure(client, &response, rollbackErr)
		}
		return sendExchangeFailure(client, &response, err)
	}
	if _, _, err := sendMailboxUpdate(client); err != nil {
		return sendExchangeFailure(client, &response, err)
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(11509, &response)
}
