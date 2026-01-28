package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
)

func TestPlayerItemQuantityEndpoint(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9300)
	itemID := uint32(5001)
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.CommanderItem{}).Error; err != nil {
		t.Fatalf("clear commander items: %v", err)
	}
	if err := orm.GormDB.Unscoped().Where("commander_id = ?", commanderID).Delete(&orm.Commander{}).Error; err != nil {
		t.Fatalf("clear commander: %v", err)
	}
	if err := orm.GormDB.Where("id = ?", itemID).Delete(&orm.Item{}).Error; err != nil {
		t.Fatalf("clear item: %v", err)
	}
	item := orm.Item{
		ID:          itemID,
		Name:        "Test Item",
		Rarity:      1,
		ShopID:      -2,
		Type:        1,
		VirtualType: 0,
	}
	if err := orm.GormDB.Create(&item).Error; err != nil {
		t.Fatalf("create item: %v", err)
	}
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "Item Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/9300/items/5001", strings.NewReader(`{"quantity":123}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var owned orm.CommanderItem
	if err := orm.GormDB.First(&owned, "commander_id = ? AND item_id = ?", commanderID, itemID).Error; err != nil {
		t.Fatalf("load commander item: %v", err)
	}
	if owned.Count != 123 {
		t.Fatalf("expected quantity 123, got %d", owned.Count)
	}
}
