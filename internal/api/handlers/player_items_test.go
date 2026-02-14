package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
)

func TestPlayerItemQuantityEndpoint(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9300)
	itemID := uint32(5001)
	execTestSQL(t, "DELETE FROM commander_items WHERE commander_id = $1", int64(commanderID))
	execTestSQL(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	execTestSQL(t, "DELETE FROM items WHERE id = $1", int64(itemID))
	item := orm.Item{
		ID:          itemID,
		Name:        "Test Item",
		Rarity:      1,
		ShopID:      -2,
		Type:        1,
		VirtualType: 0,
	}
	if err := orm.CreateItemRecord(&item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	seedCommander(t, commanderID, "Item Tester")

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/9300/items/5001", strings.NewReader(`{"quantity":123}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	count := queryUint32TestSQL(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(commanderID), int64(itemID))
	if count != 123 {
		t.Fatalf("expected quantity 123, got %d", count)
	}
}
