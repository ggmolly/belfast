package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/orm"
)

func newGameDataTestApp(t *testing.T) *iris.Application {
	initPlayerHandlerTestDB(t)
	app := iris.New()
	handler := NewGameDataHandler()
	RegisterGameDataRoutes(app.Party("/api/v1"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func seedShip(t *testing.T, templateID uint32, name string, rarityID uint32, shipType uint32, nationality uint32) {
	t.Helper()
	ship := orm.Ship{
		TemplateID:  templateID,
		Name:        name,
		EnglishName: name,
		RarityID:    rarityID,
		Star:        1,
		Type:        shipType,
		Nationality: nationality,
		BuildTime:   3600,
	}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
}

func clearShips(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM ships").Error; err != nil {
		t.Fatalf("clear ships: %v", err)
	}
}

func seedItem(t *testing.T, id uint32, name string, rarity int, itemType int) {
	t.Helper()
	item := orm.Item{
		ID:          id,
		Name:        name,
		Rarity:      rarity,
		ShopID:      1,
		Type:        itemType,
		VirtualType: 0,
	}
	if err := orm.GormDB.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
}

func clearItems(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM items").Error; err != nil {
		t.Fatalf("clear items: %v", err)
	}
}

func seedResource(t *testing.T, id uint32, itemID uint32, name string) {
	t.Helper()
	resource := orm.Resource{
		ID:     id,
		ItemID: itemID,
		Name:   name,
	}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}
}

func clearResources(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM resources").Error; err != nil {
		t.Fatalf("clear resources: %v", err)
	}
}

func seedSkin(t *testing.T, id uint32, name string, shipGroup int) {
	t.Helper()
	skin := orm.Skin{
		ID:        id,
		Name:      name,
		ShipGroup: shipGroup,
	}
	if err := orm.GormDB.Create(&skin).Error; err != nil {
		t.Fatalf("seed skin: %v", err)
	}
}

func clearSkins(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM skins").Error; err != nil {
		t.Fatalf("clear skins: %v", err)
	}
}

func seedConfigEntry(t *testing.T, category string, key string, data string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: category, Key: key, Data: json.RawMessage(data)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry: %v", err)
	}
}

func clearConfigEntriesByCategory(t *testing.T, category string) {
	t.Helper()
	if err := orm.GormDB.Where("category = ?", category).Delete(&orm.ConfigEntry{}).Error; err != nil {
		t.Fatalf("clear config entries: %v", err)
	}
}

func TestListShipsReturnsEmpty(t *testing.T) {
	app := newGameDataTestApp(t)
	clearShips(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Ships []struct {
				ID          uint32  `json:"id"`
				Name        string  `json:"name"`
				RarityID    uint32  `json:"rarity_id"`
				Star        uint32  `json:"star"`
				Type        uint32  `json:"type"`
				Nationality uint32  `json:"nationality"`
				BuildTime   uint32  `json:"build_time"`
				PoolID      *uint32 `json:"pool_id"`
			} `json:"ships"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Ships) != 0 {
		t.Fatalf("expected empty ships list, got %d", len(responseStruct.Data.Ships))
	}
	if responseStruct.Data.Meta.Total != 0 {
		t.Fatalf("expected total 0, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestListShipsReturnsData(t *testing.T) {
	app := newGameDataTestApp(t)
	clearShips(t)
	seedShip(t, 10001, "Test Ship 1", 5, 1, 1)
	seedShip(t, 10002, "Test Ship 2", 4, 2, 2)
	t.Cleanup(func() {
		clearShips(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Ships []struct {
				ID          uint32  `json:"id"`
				Name        string  `json:"name"`
				RarityID    uint32  `json:"rarity_id"`
				Star        uint32  `json:"star"`
				Type        uint32  `json:"type"`
				Nationality uint32  `json:"nationality"`
				BuildTime   uint32  `json:"build_time"`
				PoolID      *uint32 `json:"pool_id"`
			} `json:"ships"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Ships) != 2 {
		t.Fatalf("expected 2 ships, got %d", len(responseStruct.Data.Ships))
	}
	if responseStruct.Data.Meta.Total != 2 {
		t.Fatalf("expected total 2, got %d", responseStruct.Data.Meta.Total)
	}
	if responseStruct.Data.Ships[0].ID != 10001 {
		t.Fatalf("expected first ship id 10001, got %d", responseStruct.Data.Ships[0].ID)
	}
}

func TestShipDetail(t *testing.T) {
	app := newGameDataTestApp(t)
	clearShips(t)
	seedShip(t, 10003, "Test Ship 3", 5, 1, 1)
	t.Cleanup(func() {
		clearShips(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships/10003", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			ID          uint32  `json:"id"`
			Name        string  `json:"name"`
			RarityID    uint32  `json:"rarity_id"`
			Star        uint32  `json:"star"`
			Type        uint32  `json:"type"`
			Nationality uint32  `json:"nationality"`
			BuildTime   uint32  `json:"build_time"`
			PoolID      *uint32 `json:"pool_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if responseStruct.Data.ID != 10003 {
		t.Fatalf("expected ship id 10003, got %d", responseStruct.Data.ID)
	}
	if responseStruct.Data.Name != "Test Ship 3" {
		t.Fatalf("expected ship name 'Test Ship 3', got %s", responseStruct.Data.Name)
	}
}

func TestShipDetailNotFound(t *testing.T) {
	app := newGameDataTestApp(t)
	clearShips(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships/99999", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for not found")
	}
}

func TestListItemsReturnsEmpty(t *testing.T) {
	app := newGameDataTestApp(t)
	clearItems(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Items []struct {
				ID          uint32 `json:"id"`
				Name        string `json:"name"`
				Rarity      int    `json:"rarity"`
				ShopID      int    `json:"shop_id"`
				Type        int    `json:"type"`
				VirtualType int    `json:"virtual_type"`
			} `json:"items"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Items) != 0 {
		t.Fatalf("expected empty items list, got %d", len(responseStruct.Data.Items))
	}
	if responseStruct.Data.Meta.Total != 0 {
		t.Fatalf("expected total 0, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestListItemsReturnsData(t *testing.T) {
	app := newGameDataTestApp(t)
	clearItems(t)
	seedItem(t, 20001, "Test Item 1", 5, 1)
	seedItem(t, 20002, "Test Item 2", 4, 2)
	t.Cleanup(func() {
		clearItems(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Items []struct {
				ID          uint32 `json:"id"`
				Name        string `json:"name"`
				Rarity      int    `json:"rarity"`
				ShopID      int    `json:"shop_id"`
				Type        int    `json:"type"`
				VirtualType int    `json:"virtual_type"`
			} `json:"items"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(responseStruct.Data.Items))
	}
	if responseStruct.Data.Meta.Total != 2 {
		t.Fatalf("expected total 2, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestItemDetail(t *testing.T) {
	app := newGameDataTestApp(t)
	clearItems(t)
	seedItem(t, 20003, "Test Item 3", 5, 1)
	t.Cleanup(func() {
		clearItems(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/items/20003", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			ID          uint32 `json:"id"`
			Name        string `json:"name"`
			Rarity      int    `json:"rarity"`
			ShopID      int    `json:"shop_id"`
			Type        int    `json:"type"`
			VirtualType int    `json:"virtual_type"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if responseStruct.Data.ID != 20003 {
		t.Fatalf("expected item id 20003, got %d", responseStruct.Data.ID)
	}
	if responseStruct.Data.Name != "Test Item 3" {
		t.Fatalf("expected item name 'Test Item 3', got %s", responseStruct.Data.Name)
	}
}

func TestItemDetailNotFound(t *testing.T) {
	app := newGameDataTestApp(t)
	clearItems(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/items/99999", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for not found")
	}
}

func TestListResourcesReturnsEmpty(t *testing.T) {
	app := newGameDataTestApp(t)
	clearResources(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/resources", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Resources []struct {
				ID     uint32 `json:"id"`
				ItemID uint32 `json:"itemid"`
				Name   string `json:"name"`
			} `json:"resources"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Resources) != 0 {
		t.Fatalf("expected empty resources list, got %d", len(responseStruct.Data.Resources))
	}
	if responseStruct.Data.Meta.Total != 0 {
		t.Fatalf("expected total 0, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestListResourcesReturnsData(t *testing.T) {
	app := newGameDataTestApp(t)
	clearResources(t)
	seedResource(t, 1, 1, "Gold")
	seedResource(t, 2, 2, "Oil")
	t.Cleanup(func() {
		clearResources(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/resources", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Resources []struct {
				ID     uint32 `json:"id"`
				ItemID uint32 `json:"itemid"`
				Name   string `json:"name"`
			} `json:"resources"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(responseStruct.Data.Resources))
	}
	if responseStruct.Data.Meta.Total != 2 {
		t.Fatalf("expected total 2, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestResourceDetail(t *testing.T) {
	app := newGameDataTestApp(t)
	clearResources(t)
	seedResource(t, 3, 3, "Gem")
	t.Cleanup(func() {
		clearResources(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/resources/3", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			ID     uint32 `json:"id"`
			ItemID uint32 `json:"itemid"`
			Name   string `json:"name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if responseStruct.Data.ID != 3 {
		t.Fatalf("expected resource id 3, got %d", responseStruct.Data.ID)
	}
	if responseStruct.Data.Name != "Gem" {
		t.Fatalf("expected resource name 'Gem', got %s", responseStruct.Data.Name)
	}
}

func TestResourceDetailNotFound(t *testing.T) {
	app := newGameDataTestApp(t)
	clearResources(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/resources/99999", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for not found")
	}
}

func TestListSkinsReturnsEmpty(t *testing.T) {
	app := newGameDataTestApp(t)
	clearSkins(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/skins", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Skins []struct {
				ID        uint32 `json:"id"`
				Name      string `json:"name"`
				ShipGroup int    `json:"ship_group"`
			} `json:"skins"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Skins) != 0 {
		t.Fatalf("expected empty skins list, got %d", len(responseStruct.Data.Skins))
	}
	if responseStruct.Data.Meta.Total != 0 {
		t.Fatalf("expected total 0, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestListSkinsReturnsData(t *testing.T) {
	app := newGameDataTestApp(t)
	clearSkins(t)
	seedSkin(t, 30001, "Test Skin 1", 100)
	seedSkin(t, 30002, "Test Skin 2", 200)
	t.Cleanup(func() {
		clearSkins(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/skins", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Skins []struct {
				ID        uint32 `json:"id"`
				Name      string `json:"name"`
				ShipGroup int    `json:"ship_group"`
			} `json:"skins"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Skins) != 2 {
		t.Fatalf("expected 2 skins, got %d", len(responseStruct.Data.Skins))
	}
	if responseStruct.Data.Meta.Total != 2 {
		t.Fatalf("expected total 2, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestSkinDetail(t *testing.T) {
	app := newGameDataTestApp(t)
	clearSkins(t)
	seedSkin(t, 30003, "Test Skin 3", 300)
	t.Cleanup(func() {
		clearSkins(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/skins/30003", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			ID        uint32 `json:"id"`
			Name      string `json:"name"`
			ShipGroup int    `json:"ship_group"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if responseStruct.Data.ID != 30003 {
		t.Fatalf("expected skin id 30003, got %d", responseStruct.Data.ID)
	}
	if responseStruct.Data.Name != "Test Skin 3" {
		t.Fatalf("expected skin name 'Test Skin 3', got %s", responseStruct.Data.Name)
	}
}

func TestSkinDetailNotFound(t *testing.T) {
	app := newGameDataTestApp(t)
	clearSkins(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/skins/99999", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for not found")
	}
}

func TestListConfigEntries(t *testing.T) {
	app := newGameDataTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/livingarea-covers", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Entries []struct {
				Key  string `json:"key"`
				Data struct {
					Value json.RawMessage `json:"value"`
				} `json:"data"`
			} `json:"entries"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
}

func TestShipSkinsReturnsData(t *testing.T) {
	app := newGameDataTestApp(t)
	clearSkins(t)
	seedSkin(t, 40001, "Ship Skin A", 1100)
	seedSkin(t, 40002, "Ship Skin B", 1100)
	seedSkin(t, 40003, "Other Skin", 1200)
	defer clearSkins(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships/1100/skins", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Skins []struct {
				ID uint32 `json:"id"`
			} `json:"skins"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Skins) != 2 {
		t.Fatalf("expected 2 skins, got %d", len(responseStruct.Data.Skins))
	}
}

func TestShipSkinsBadRequest(t *testing.T) {
	app := newGameDataTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships/0/skins", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestListFrameConfigs(t *testing.T) {
	app := newGameDataTestApp(t)
	clearConfigEntriesByCategory(t, "ShareCfg/item_data_frame.json")
	clearConfigEntriesByCategory(t, "ShareCfg/item_data_chat.json")
	clearConfigEntriesByCategory(t, "ShareCfg/item_data_battleui.json")
	seedConfigEntry(t, "ShareCfg/item_data_frame.json", "1", `{"id":1}`)
	seedConfigEntry(t, "ShareCfg/item_data_chat.json", "2", `{"id":2}`)
	seedConfigEntry(t, "ShareCfg/item_data_battleui.json", "3", `{"id":3}`)
	defer func() {
		clearConfigEntriesByCategory(t, "ShareCfg/item_data_frame.json")
		clearConfigEntriesByCategory(t, "ShareCfg/item_data_chat.json")
		clearConfigEntriesByCategory(t, "ShareCfg/item_data_battleui.json")
	}()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/attire/icon-frames", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/attire/chat-frames", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/attire/battle-ui", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}
