package orm

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
	"github.com/ggmolly/belfast/internal/logger"
)

type Mail struct {
	ID                   uint32    `gorm:"primary_key"`
	ReceiverID           uint32    `gorm:"not_null"`
	Read                 bool      `gorm:"not_null;default:false"`
	Date                 time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	Title                string    `gorm:"type:varchar(40);not_null"`
	Body                 string    `gorm:"type:varchar(2000);not_null"`
	AttachmentsCollected bool      `gorm:"not_null;default:false"`
	IsImportant          bool      `gorm:"not_null;default:false"`
	CustomSender         *string   `gorm:"type:varchar(30)"`
	IsArchived           bool      `gorm:"not_null;default:false"`
	CreatedAt            time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`

	Attachments []MailAttachment `gorm:"foreignkey:MailID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Commander   Commander        `gorm:"foreignkey:ReceiverID;references:CommanderID"`
}

// TODO: Unused yet
type MailAttachment struct {
	ID       uint32 `gorm:"primary_key"`
	MailID   uint32 `gorm:"not_null"`
	Type     uint32 `gorm:"not_null"`
	ItemID   uint32 `gorm:"not_null"`
	Quantity uint32 `gorm:"not_null"`
}

func (m *Mail) Create() error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		row, err := q.CreateMail(ctx, gen.CreateMailParams{
			ReceiverID:           int64(m.ReceiverID),
			Read:                 m.Read,
			Title:                m.Title,
			Body:                 m.Body,
			AttachmentsCollected: m.AttachmentsCollected,
			IsImportant:          m.IsImportant,
			CustomSender:         pgTextFromPtr(m.CustomSender),
			IsArchived:           m.IsArchived,
		})
		if err != nil {
			return err
		}
		m.ID = uint32(row.ID)
		m.Date = row.Date.Time
		m.CreatedAt = row.CreatedAt.Time
		for i := range m.Attachments {
			id, err := q.CreateMailAttachment(ctx, gen.CreateMailAttachmentParams{
				MailID:   int64(m.ID),
				Type:     int64(m.Attachments[i].Type),
				ItemID:   int64(m.Attachments[i].ItemID),
				Quantity: int64(m.Attachments[i].Quantity),
			})
			if err != nil {
				return err
			}
			m.Attachments[i].ID = uint32(id)
			m.Attachments[i].MailID = m.ID
		}
		return nil
	})
}

func (m *Mail) Delete() error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Queries.DeleteMail(ctx, gen.DeleteMailParams{ReceiverID: int64(m.ReceiverID), ID: int64(m.ID)})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func GetMailByReceiverAndID(receiverID uint32, mailID uint32) (*Mail, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, receiver_id, read, date, title, body, attachments_collected, is_important, custom_sender, is_archived, created_at
FROM mails
WHERE receiver_id = $1
  AND id = $2
`, int64(receiverID), int64(mailID))

	mail := Mail{}
	var customSender pgtype.Text
	err := row.Scan(
		&mail.ID,
		&mail.ReceiverID,
		&mail.Read,
		&mail.Date,
		&mail.Title,
		&mail.Body,
		&mail.AttachmentsCollected,
		&mail.IsImportant,
		&customSender,
		&mail.IsArchived,
		&mail.CreatedAt,
	)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	mail.CustomSender = pgTextPtr(customSender)

	attachments, err := db.DefaultStore.Queries.ListMailAttachmentsByMailIDs(ctx, []int64{int64(mailID)})
	if err != nil {
		return nil, err
	}
	mail.Attachments = make([]MailAttachment, 0, len(attachments))
	for _, attachment := range attachments {
		mail.Attachments = append(mail.Attachments, MailAttachment{
			ID:       uint32(attachment.ID),
			MailID:   uint32(attachment.MailID),
			Type:     uint32(attachment.Type),
			ItemID:   uint32(attachment.ItemID),
			Quantity: uint32(attachment.Quantity),
		})
	}

	return &mail, nil
}

func (m *Mail) Update() error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Queries.UpdateMail(ctx, gen.UpdateMailParams{
		ReceiverID:           int64(m.ReceiverID),
		ID:                   int64(m.ID),
		Read:                 m.Read,
		Title:                m.Title,
		Body:                 m.Body,
		AttachmentsCollected: m.AttachmentsCollected,
		IsImportant:          m.IsImportant,
		CustomSender:         pgTextFromPtr(m.CustomSender),
		IsArchived:           m.IsArchived,
	})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func (m *Mail) SetRead(read bool) error {
	m.Read = read
	return m.Update()
}

func (m *Mail) SetImportant(important bool) error {
	m.IsImportant = important
	return m.Update()
}

// TODO: Check whether the Commander has enough space to archive the mail
func (m *Mail) SetArchived(archived bool) error {
	m.IsArchived = archived
	return m.Update()
}

// CollectAttachments returns the attachments and marks the mail as collected.
func (m *Mail) CollectAttachments(commander *Commander) ([]MailAttachment, error) {
	attachments := make([]MailAttachment, len(m.Attachments))
	for i, attachment := range m.Attachments {
		attachments[i] = attachment
		switch attachment.Type {
		case consts.DROP_TYPE_RESOURCE:
			commander.AddResource(attachment.ItemID, attachment.Quantity)
		case consts.DROP_TYPE_ITEM:
			commander.AddItem(attachment.ItemID, attachment.Quantity)
		case consts.DROP_TYPE_SHIP:
			for count := uint32(0); count < attachment.Quantity; count++ {
				if _, err := commander.AddShip(attachment.ItemID); err != nil {
					return nil, err
				}
			}
		case consts.DROP_TYPE_SKIN:
			for count := uint32(0); count < attachment.Quantity; count++ {
				if err := commander.GiveSkin(attachment.ItemID); err != nil {
					return nil, err
				}
			}
		default:
			logger.LogEvent("Mail", "CollectAttachments", fmt.Sprintf("Unknown attachment type %d", attachment.Type), logger.LOG_LEVEL_ERROR)
		}
	}
	m.AttachmentsCollected = true
	return attachments, m.Update()
}
