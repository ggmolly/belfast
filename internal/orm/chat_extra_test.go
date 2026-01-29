package orm

import (
	"testing"
	"time"
)

func TestMessageCRUDAndHistory(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Message{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 40, AccountID: 40, Name: "Chatter"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	message := Message{SenderID: commander.CommanderID, RoomID: 1, Content: "hello", SentAt: time.Now()}
	if err := message.Create(); err != nil {
		t.Fatalf("create message: %v", err)
	}
	message.Content = "updated"
	if err := message.Update(); err != nil {
		t.Fatalf("update message: %v", err)
	}
	if err := message.Delete(); err != nil {
		t.Fatalf("delete message: %v", err)
	}

	for i := 0; i < 3; i++ {
		msg := Message{SenderID: commander.CommanderID, RoomID: 2, Content: "msg", SentAt: time.Now().Add(time.Duration(i) * time.Second)}
		if err := msg.Create(); err != nil {
			t.Fatalf("seed message: %v", err)
		}
	}
	history, err := GetRoomHistory(2)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(history))
	}
}

func TestSendMessage(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Message{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 41, AccountID: 41, Name: "Sender"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	message, err := SendMessage(3, "content", &commander)
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if message.RoomID != 3 || message.Content != "content" {
		t.Fatalf("unexpected message: %+v", message)
	}
}
