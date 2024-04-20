package orm

import "time"

type Like struct {
	GroupID   uint32 `gorm:"primary_key;index:idx_likes_group_id_liker_id"`
	LikerID   uint32 `gorm:"not_null;primary_key;index:idx_likes_group_id_liker_id"`
	CreatedAt time.Time

	Liker Commander `gorm:"foreignkey:LikerID;references:CommanderID"`
}
