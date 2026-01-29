package orm

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestNoticeCRUD(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Notice{})

	notice := Notice{ID: 1, Version: "1", BtnTitle: "Btn", Title: "Title", TitleImage: "Img", TimeDesc: "Now", Content: "Body", TagType: 1, Icon: 1, Track: "T"}
	if err := notice.Create(); err != nil {
		t.Fatalf("create notice: %v", err)
	}
	notice.Title = "Updated"
	if err := notice.Update(); err != nil {
		t.Fatalf("update notice: %v", err)
	}
	loaded := Notice{ID: 1}
	if err := loaded.Retrieve(false); err != nil {
		t.Fatalf("retrieve notice: %v", err)
	}
	if loaded.Title != "Updated" {
		t.Fatalf("expected updated title")
	}
	if err := loaded.Delete(); err != nil {
		t.Fatalf("delete notice: %v", err)
	}
	if err := loaded.Retrieve(false); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
