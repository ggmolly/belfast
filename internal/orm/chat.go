package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

const (
	MSG_TYPE_BANNED  = 100
	MSG_ACTOBSS_WORD = 1000 // litterally no idea wtf this is
	MSG_TYPE_NORMAL  = 1
)

type Message struct {
	ID       uint32    `gorm:"primary_key"`
	SenderID uint32    `gorm:"not_null"`
	RoomID   uint32    `gorm:"not_null"`
	SentAt   time.Time `gorm:"not_null;default:CURRENT_TIMESTAMP;index:idx_sent_at,sort:desc"`
	Content  string    `gorm:"not_null;type:varchar(512)"`

	Sender Commander `gorm:"foreignkey:SenderID;references:CommanderID"`
}

func (m *Message) Create() error {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.CreateMessage(ctx, gen.CreateMessageParams{SenderID: int64(m.SenderID), RoomID: int64(m.RoomID), Content: m.Content})
	if err != nil {
		return err
	}
	m.ID = uint32(row.ID)
	m.SentAt = row.SentAt.Time
	return nil
}

func (m *Message) Update() error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpdateMessageContent(ctx, gen.UpdateMessageContentParams{ID: int64(m.ID), Content: m.Content})
}

func (m *Message) Delete() error {
	ctx := context.Background()
	_, err := db.DefaultStore.Queries.DeleteMessageByID(ctx, int64(m.ID))
	return err
}

// Returns the last 50 messages from a room
func GetRoomHistory(roomID uint32) ([]Message, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListRoomHistory(ctx, int64(roomID))
	if err != nil {
		return nil, err
	}
	var messages []Message
	for _, r := range rows {
		messages = append(messages, Message{ID: uint32(r.ID), SenderID: uint32(r.SenderID), RoomID: uint32(r.RoomID), SentAt: r.SentAt.Time, Content: r.Content})
	}
	return messages, nil
}

// Inserts a message in the database
func SendMessage(roomID uint32, content string, sender *Commander) (*Message, error) {
	message := Message{
		SenderID: sender.CommanderID,
		RoomID:   roomID,
		Content:  content,
	}
	err := message.Create()
	return &message, err
}
