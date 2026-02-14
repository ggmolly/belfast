package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
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
	ctx := context.Background()
	id, err := db.DefaultStore.Queries.CreateGuildChatMessage(ctx, gen.CreateGuildChatMessageParams{
		GuildID:  int64(guildID),
		SenderID: int64(senderID),
		SentAt:   pgTimestamptz(sentAt),
		Content:  content,
	})
	if err != nil {
		return nil, err
	}
	message.ID = uint32(id)
	return &message, nil
}

func ListGuildChatMessages(guildID uint32, limit int) ([]GuildChatMessage, error) {
	ctx := context.Background()
	var messages []GuildChatMessage
	if limit > 0 {
		rows, err := db.DefaultStore.Queries.ListGuildChatMessages(ctx, gen.ListGuildChatMessagesParams{GuildID: int64(guildID), Limit: int32(limit)})
		if err != nil {
			return nil, err
		}
		messages = make([]GuildChatMessage, 0, len(rows))
		for _, r := range rows {
			messages = append(messages, GuildChatMessage{
				ID:       uint32(r.ID),
				GuildID:  uint32(r.GuildID),
				SenderID: uint32(r.SenderID),
				SentAt:   r.SentAt.Time,
				Content:  r.Content,
				Sender: Commander{
					CommanderID:         uint32(r.SenderCommanderID),
					Name:                r.SenderName,
					Level:               int(r.SenderLevel),
					DisplayIconID:       uint32(r.SenderDisplayIconID),
					DisplaySkinID:       uint32(r.SenderDisplaySkinID),
					SelectedIconFrameID: uint32(r.SenderSelectedIconFrameID),
					SelectedChatFrameID: uint32(r.SenderSelectedChatFrameID),
					DisplayIconThemeID:  uint32(r.SenderDisplayIconThemeID),
				},
			})
		}
	} else {
		rows, err := db.DefaultStore.Queries.ListGuildChatMessagesAll(ctx, int64(guildID))
		if err != nil {
			return nil, err
		}
		messages = make([]GuildChatMessage, 0, len(rows))
		for _, r := range rows {
			messages = append(messages, GuildChatMessage{
				ID:       uint32(r.ID),
				GuildID:  uint32(r.GuildID),
				SenderID: uint32(r.SenderID),
				SentAt:   r.SentAt.Time,
				Content:  r.Content,
				Sender: Commander{
					CommanderID:         uint32(r.SenderCommanderID),
					Name:                r.SenderName,
					Level:               int(r.SenderLevel),
					DisplayIconID:       uint32(r.SenderDisplayIconID),
					DisplaySkinID:       uint32(r.SenderDisplaySkinID),
					SelectedIconFrameID: uint32(r.SenderSelectedIconFrameID),
					SelectedChatFrameID: uint32(r.SenderSelectedChatFrameID),
					DisplayIconThemeID:  uint32(r.SenderDisplayIconThemeID),
				},
			})
		}
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}
