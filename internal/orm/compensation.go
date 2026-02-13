package orm

import (
	"context"
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
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
}

type CompensationAttachment struct {
	ID             uint32 `gorm:"primary_key"`
	CompensationID uint32 `gorm:"not_null"`
	Type           uint32 `gorm:"not_null"`
	ItemID         uint32 `gorm:"not_null"`
	Quantity       uint32 `gorm:"not_null"`
}

func (c *Compensation) Create() error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		row, err := q.CreateCompensation(ctx, gen.CreateCompensationParams{
			CommanderID: int64(c.CommanderID),
			Title:       c.Title,
			Text:        c.Text,
			SendTime:    pgTimestamptz(c.SendTime),
			ExpiresAt:   pgTimestamptz(c.ExpiresAt),
			AttachFlag:  c.AttachFlag,
		})
		if err != nil {
			return err
		}
		c.ID = uint32(row.ID)
		c.SendTime = row.SendTime.Time
		c.CreatedAt = row.CreatedAt.Time
		for _, attachment := range c.Attachments {
			if err := q.CreateCompensationAttachment(ctx, gen.CreateCompensationAttachmentParams{
				CompensationID: int64(c.ID),
				Type:           int64(attachment.Type),
				ItemID:         int64(attachment.ItemID),
				Quantity:       int64(attachment.Quantity),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *Compensation) Update() error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.UpdateCompensation(ctx, gen.UpdateCompensationParams{
			ID:         int64(c.ID),
			Title:      c.Title,
			Text:       c.Text,
			SendTime:   pgTimestamptz(c.SendTime),
			ExpiresAt:  pgTimestamptz(c.ExpiresAt),
			AttachFlag: c.AttachFlag,
		}); err != nil {
			return err
		}
		if err := q.DeleteCompensationAttachmentsByCompensationID(ctx, int64(c.ID)); err != nil {
			return err
		}
		for _, attachment := range c.Attachments {
			if err := q.CreateCompensationAttachment(ctx, gen.CreateCompensationAttachmentParams{
				CompensationID: int64(c.ID),
				Type:           int64(attachment.Type),
				ItemID:         int64(attachment.ItemID),
				Quantity:       int64(attachment.Quantity),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *Compensation) Delete() error {
	ctx := context.Background()
	return db.DefaultStore.Queries.DeleteCompensation(ctx, int64(c.ID))
}

func GetCompensationByCommanderAndID(commanderID uint32, compensationID uint32) (*Compensation, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetCompensation(ctx, int64(compensationID))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	if uint32(row.CommanderID) != commanderID {
		return nil, db.ErrNotFound
	}

	compensation := Compensation{
		ID:          uint32(row.ID),
		CommanderID: uint32(row.CommanderID),
		Title:       row.Title,
		Text:        row.Text,
		SendTime:    row.SendTime.Time,
		ExpiresAt:   row.ExpiresAt.Time,
		AttachFlag:  row.AttachFlag,
		CreatedAt:   row.CreatedAt.Time,
	}

	attachments, err := db.DefaultStore.Queries.ListCompensationAttachmentsByCompensationIDs(ctx, []int64{int64(compensationID)})
	if err != nil {
		return nil, err
	}
	compensation.Attachments = make([]CompensationAttachment, 0, len(attachments))
	for _, attachment := range attachments {
		compensation.Attachments = append(compensation.Attachments, CompensationAttachment{
			ID:             uint32(attachment.ID),
			CompensationID: uint32(attachment.CompensationID),
			Type:           uint32(attachment.Type),
			ItemID:         uint32(attachment.ItemID),
			Quantity:       uint32(attachment.Quantity),
		})
	}

	return &compensation, nil
}

func DeleteCompensationByCommanderAndID(commanderID uint32, compensationID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM compensations
WHERE commander_id = $1
  AND id = $2
`, int64(commanderID), int64(compensationID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
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
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCompensationsByCommander(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	compensations := make([]Compensation, 0, len(rows))
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		compensations = append(compensations, Compensation{
			ID:          uint32(r.ID),
			CommanderID: uint32(r.CommanderID),
			Title:       r.Title,
			Text:        r.Text,
			SendTime:    r.SendTime.Time,
			ExpiresAt:   r.ExpiresAt.Time,
			AttachFlag:  r.AttachFlag,
			CreatedAt:   r.CreatedAt.Time,
			Attachments: []CompensationAttachment{},
		})
		ids = append(ids, r.ID)
	}
	if len(ids) == 0 {
		return compensations, nil
	}
	attRows, err := db.DefaultStore.Queries.ListCompensationAttachmentsByCompensationIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byCompID := make(map[uint32]int, len(compensations))
	for i := range compensations {
		byCompID[compensations[i].ID] = i
	}
	for _, ar := range attRows {
		idx, ok := byCompID[uint32(ar.CompensationID)]
		if !ok {
			continue
		}
		compensations[idx].Attachments = append(compensations[idx].Attachments, CompensationAttachment{
			ID:             uint32(ar.ID),
			CompensationID: uint32(ar.CompensationID),
			Type:           uint32(ar.Type),
			ItemID:         uint32(ar.ItemID),
			Quantity:       uint32(ar.Quantity),
		})
	}
	return compensations, nil
}
