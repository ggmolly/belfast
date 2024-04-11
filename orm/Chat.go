package orm

import (
	"time"
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
	return GormDB.Create(m).Error
}

func (m *Message) Update() error {
	return GormDB.Save(m).Error
}

func (m *Message) Delete() error {
	return GormDB.Delete(m).Error
}

// Returns the last 50 messages from a room
func GetRoomHistory(roomID uint32) ([]Message, error) {
	var messages []Message
	err := GormDB.
		Where("room_id = ?", roomID).
		Order("sent_at DESC").
		Limit(50).
		Find(&messages).
		Error
	return messages, err
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
