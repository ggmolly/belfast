package orm

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/protobuf"
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
	return GormDB.Create(m).Error
}

func (m *Mail) Delete() error {
	return GormDB.Delete(m).Error
}

func (m *Mail) Update() error {
	return GormDB.Save(m).Error
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

// Returns a list of protobuf.ATTACHMENT containing all attachments of the mail, and marks the mail as "collected"
func (m *Mail) CollectAttachments(commander *Commander) ([]*protobuf.ATTACHMENT, error) {
	attachments := make([]*protobuf.ATTACHMENT, len(m.Attachments))
	for i, attachment := range m.Attachments {
		attachments[i] = &protobuf.ATTACHMENT{
			Type:   &attachment.Type,
			Id:     &attachment.ItemID,
			Number: &attachment.Quantity,
		}
		switch attachment.Type {
		case 1: // resources
			commander.AddResource(attachment.ItemID, attachment.Quantity)
		case 2: // item
			commander.AddItem(attachment.ItemID, attachment.Quantity)
		default:
			logger.LogEvent("Mail", "CollectAttachments", fmt.Sprintf("Unknown attachment type %d", attachment.Type), logger.LOG_LEVEL_ERROR)
		}
	}
	m.AttachmentsCollected = true
	err := m.Update()
	return attachments, err
}
