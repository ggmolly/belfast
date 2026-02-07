package expedition

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"gorm.io/gorm"
)

func setupSubmarineConfigTest(t *testing.T) {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	if err := orm.GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&orm.ConfigEntry{}).Error; err != nil {
		t.Fatalf("clear config entries: %v", err)
	}
}

func seedConfigEntry(t *testing.T, category string, key string, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: category, Key: key, Data: json.RawMessage(payload)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry: %v", err)
	}
}

func TestLoadSubmarineChaptersFiltersAndAppliesLevelLimits(t *testing.T) {
	setupSubmarineConfigTest(t)
	seedConfigEntry(t, submarineDailyTemplateCategory, "501", `{"id":501,"expedition_and_lv_limit_list":[[1000,35],[1001,45],[1002,55],[1003,65],[1004,75],[1005,95]]}`)
	for id := 1000; id <= 1005; id++ {
		seedConfigEntry(t, submarineDataTemplateCategory, fmt.Sprintf("%d", id), fmt.Sprintf(`{"id":%d,"type":15}`, id))
	}
	seedConfigEntry(t, submarineDataTemplateCategory, "noop", `{"id":9999,"type":15}`)
	seedConfigEntry(t, submarineDataTemplateCategory, "wrong", `{"id":1000,"type":14}`)

	chapters, err := LoadSubmarineChapters()
	if err != nil {
		t.Fatalf("load chapters: %v", err)
	}
	if len(chapters) != 6 {
		t.Fatalf("expected 6 chapters, got %d", len(chapters))
	}
	for i, chapter := range chapters {
		expectedID := uint32(1000 + i)
		if chapter.ChapterID != expectedID {
			t.Fatalf("expected chapter id %d, got %d", expectedID, chapter.ChapterID)
		}
		if chapter.Index != uint32(i) {
			t.Fatalf("expected index %d, got %d", i, chapter.Index)
		}
		if chapter.MinLevel == 0 {
			t.Fatalf("expected min level to be populated")
		}
	}
}
