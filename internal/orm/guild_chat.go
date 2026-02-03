package orm

import (
	"time"
)

type GuildChatMessage struct {
	ID       uint32    `gorm:"primary_key"`
	GuildID  uint32    `gorm:"not_null;index:idx_guild_chat_time,priority:1"`
	SenderID uint32    `gorm:"not_null"`
	SentAt   time.Time `gorm:"not_null;default:CURRENT_TIMESTAMP;index:idx_guild_chat_time,priority:2,sort:desc"`
	Content  string    `gorm:"not_null;type:varchar(512)"`

	Sender Commander `gorm:"foreignkey:SenderID;references:CommanderID"`
}

func CreateGuildChatMessage(guildID uint32, senderID uint32, content string, sentAt time.Time) (*GuildChatMessage, error) {
	message := GuildChatMessage{
		GuildID:  guildID,
		SenderID: senderID,
		SentAt:   sentAt,
		Content:  content,
	}
	if err := GormDB.Create(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func ListGuildChatMessages(guildID uint32, limit int) ([]GuildChatMessage, error) {
	var messages []GuildChatMessage
	query := GormDB.
		Where("guild_id = ?", guildID).
		Preload("Sender").
		Order("sent_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}
