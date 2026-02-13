package orm

import "testing"

func TestAddChapterDropIdempotent(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ChapterDrop{})

	drop := &ChapterDrop{CommanderID: 1, ChapterID: 101, ShipID: 2001}
	if err := AddChapterDrop(drop); err != nil {
		t.Fatalf("add drop: %v", err)
	}
	if err := AddChapterDrop(drop); err != nil {
		t.Fatalf("add drop again: %v", err)
	}
	drops, err := GetChapterDrops(1, 101)
	if err != nil {
		t.Fatalf("get drops: %v", err)
	}
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
}

func TestGetChapterDropsFiltersByChapter(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ChapterDrop{})

	_ = AddChapterDrop(&ChapterDrop{CommanderID: 1, ChapterID: 101, ShipID: 2001})
	_ = AddChapterDrop(&ChapterDrop{CommanderID: 1, ChapterID: 102, ShipID: 2002})
	_ = AddChapterDrop(&ChapterDrop{CommanderID: 2, ChapterID: 101, ShipID: 2003})

	drops, err := GetChapterDrops(1, 101)
	if err != nil {
		t.Fatalf("get drops: %v", err)
	}
	if len(drops) != 1 || drops[0].ShipID != 2001 {
		t.Fatalf("unexpected drops: %+v", drops)
	}
}
