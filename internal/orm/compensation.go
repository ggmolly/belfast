package orm

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/logger"
)

type Compensation struct {
	ID          uint32    `gorm:"primary_key"`
	CommanderID uint32    `gorm:"not_null"`
	Title       string    `gorm:"type:varchar(100);not_null"`
	Text        string    `gorm:"type:varchar(2000);not_null"`
	SendTime    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	ExpiresAt   time.Time `gorm:"type:timestamp;not_null"`
	AttachFlag  bool      `gorm:"not_null;default:false"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`

	Attachments []CompensationAttachment `gorm:"foreignkey:CompensationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Commander   Commander                `gorm:"foreignkey:CommanderID;references:CommanderID"`
}

type CompensationAttachment struct {
	ID             uint32 `gorm:"primary_key"`
	CompensationID uint32 `gorm:"not_null"`
	Type           uint32 `gorm:"not_null"`
	ItemID         uint32 `gorm:"not_null"`
	Quantity       uint32 `gorm:"not_null"`
}

func (c *Compensation) Create() error {
	return GormDB.Create(c).Error
}

func (c *Compensation) Update() error {
	return GormDB.Save(c).Error
}

func (c *Compensation) Delete() error {
	return GormDB.Delete(c).Error
}

func (c *Compensation) IsExpired(now time.Time) bool {
	if c.ExpiresAt.IsZero() {
		return true
	}
	return !c.ExpiresAt.After(now)
}

func (c *Compensation) CollectAttachments(commander *Commander) ([]CompensationAttachment, error) {
	attachments := make([]CompensationAttachment, len(c.Attachments))
	for i, attachment := range c.Attachments {
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
			logger.LogEvent("Compensation", "CollectAttachments", fmt.Sprintf("Unknown attachment type %d", attachment.Type), logger.LOG_LEVEL_ERROR)
		}
	}
	c.AttachFlag = true
	return attachments, c.Update()
}

func CompensationSummary(compensations []Compensation, now time.Time) (uint32, uint32) {
	var count uint32
	var maxTimestamp uint32
	nowUnix := uint32(now.Unix())
	for _, compensation := range compensations {
		if compensation.IsExpired(now) {
			continue
		}
		expiresAt := uint32(compensation.ExpiresAt.Unix())
		if !compensation.AttachFlag {
			count++
		}
		if expiresAt > maxTimestamp {
			maxTimestamp = expiresAt
		}
	}
	if maxTimestamp <= nowUnix {
		maxTimestamp = 0
	}
	return count, maxTimestamp
}

func LoadCommanderCompensations(commanderID uint32) ([]Compensation, error) {
	var compensations []Compensation
	if err := GormDB.Preload("Attachments").Where("commander_id = ?", commanderID).Find(&compensations).Error; err != nil {
		return nil, err
	}
	return compensations, nil
}
