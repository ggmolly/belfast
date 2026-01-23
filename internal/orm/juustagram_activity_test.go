package orm

import (
	"sync"
	"testing"
	"time"
)

var juustagramOrmTestOnce sync.Once

func initJuustagramOrmTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	juustagramOrmTestOnce.Do(func() {
		InitDatabase()
	})
}

func TestJuustagramTemplateQueries(t *testing.T) {
	initJuustagramOrmTestDB(t)
	if err := GormDB.Exec("DELETE FROM juustagram_templates").Error; err != nil {
		t.Fatalf("clear templates: %v", err)
	}
	if err := GormDB.Exec("DELETE FROM juustagram_message_states").Error; err != nil {
		t.Fatalf("clear message states: %v", err)
	}
	message := JuustagramTemplate{
		ID:             10,
		GroupID:        1,
		ShipGroup:      999,
		Name:           "Test",
		Sculpture:      "test",
		PicturePersist: "https://example.com",
		MessagePersist: "ins_10",
		IsActive:       0,
		TimePersist:    JuustagramTimeConfig{{2024, 1, 1}, {0, 0, 0}},
	}
	if err := GormDB.Create(&message).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	templates, total, err := ListJuustagramTemplates(0, 10)
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(templates) != 1 || templates[0].ID != 10 {
		t.Fatalf("unexpected template list")
	}
	state, err := GetOrCreateJuustagramMessageState(42, 10, uint32(time.Now().Unix()))
	if err != nil {
		t.Fatalf("create state: %v", err)
	}
	if state.CommanderID != 42 || state.MessageID != 10 {
		t.Fatalf("unexpected state values")
	}
}
