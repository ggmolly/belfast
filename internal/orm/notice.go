package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type Notice struct {
	ID         int    `gorm:"primary_key"`
	Version    string `gorm:"default:'1';not_null"`
	BtnTitle   string `gorm:"type:varchar(48);not_null"`
	Title      string `gorm:"type:varchar(48);not_null"`
	TitleImage string `gorm:"type:text;not_null"`
	TimeDesc   string `gorm:"type:varchar(10);not_null"`
	Content    string `gorm:"type:text;not_null"`
	TagType    int    `gorm:"not_null;default:1"`
	Icon       int    `gorm:"not_null;default:1"`
	Track      string `gorm:"type:varchar(10);not_null"`
}

// Inserts or updates a notice in the database (based on the primary key)
func (n *Notice) Create() error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertNotice(ctx, gen.UpsertNoticeParams{
		ID:         int64(n.ID),
		Version:    n.Version,
		BtnTitle:   n.BtnTitle,
		Title:      n.Title,
		TitleImage: n.TitleImage,
		TimeDesc:   n.TimeDesc,
		Content:    n.Content,
		TagType:    int64(n.TagType),
		Icon:       int64(n.Icon),
		Track:      n.Track,
	})
}

// Updates a notice in the database
func (n *Notice) Update() error {
	return n.Create()
}

// Gets a notice from the database by its primary key
// If greedy is true, it will also load the relations
func (n *Notice) Retrieve(greedy bool) error {
	// ignore greediness because there are no relations
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetNotice(ctx, int64(n.ID))
	err = db.MapNotFound(err)
	if err != nil {
		return err
	}
	n.ID = int(row.ID)
	n.Version = row.Version
	n.BtnTitle = row.BtnTitle
	n.Title = row.Title
	n.TitleImage = row.TitleImage
	n.TimeDesc = row.TimeDesc
	n.Content = row.Content
	n.TagType = int(row.TagType)
	n.Icon = int(row.Icon)
	n.Track = row.Track
	return nil
}

// Deletes a notice from the database
func (n *Notice) Delete() error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Queries.DeleteNotice(ctx, int64(n.ID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
