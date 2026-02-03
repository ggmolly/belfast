package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type playerEquipmentResponse struct {
	OK   bool                          `json:"ok"`
	Data types.PlayerEquipmentResponse `json:"data"`
}

type playerEquipmentEntryResponse struct {
	OK   bool                       `json:"ok"`
	Data types.PlayerEquipmentEntry `json:"data"`
}

type playerShipEquipmentResponse struct {
	OK   bool                              `json:"ok"`
	Data types.PlayerShipEquipmentResponse `json:"data"`
}

func TestPlayerEquipmentEndpoints(t *testing.T) {
	setupTestAPI(t)
	if err := orm.GormDB.Exec("DELETE FROM owned_ship_equipments").Error; err != nil {
		t.Fatalf("clear owned_ship_equipments: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM owned_equipments").Error; err != nil {
		t.Fatalf("clear owned_equipments: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM owned_ships").Error; err != nil {
		t.Fatalf("clear owned_ships: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM ships").Error; err != nil {
		t.Fatalf("clear ships: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM config_entries").Error; err != nil {
		t.Fatalf("clear config_entries: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commanders").Error; err != nil {
		t.Fatalf("clear commanders: %v", err)
	}

	commander := orm.Commander{CommanderID: 9001, AccountID: 9001, Name: "Equip API"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: commander.CommanderID, EquipmentID: 2001, Count: 2}).Error; err != nil {
		t.Fatalf("create owned equipment: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/9001/equipment", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var bagPayload playerEquipmentResponse
	if err := json.NewDecoder(response.Body).Decode(&bagPayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !bagPayload.OK || len(bagPayload.Data.Equipment) != 1 || bagPayload.Data.Equipment[0].Count != 2 {
		t.Fatalf("unexpected equipment response")
	}

	body := []byte(`{"equipment_id":2001,"count":3}`)
	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/9001/equipment", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9001/equipment/2001", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var entryPayload playerEquipmentEntryResponse
	if err := json.NewDecoder(response.Body).Decode(&entryPayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !entryPayload.OK || entryPayload.Data.Count != 3 {
		t.Fatalf("expected updated count 3")
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/players/9001/equipment/2001", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	ship := orm.Ship{TemplateID: 7001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 1}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
	owned := orm.OwnedShip{ID: 8001, OwnerID: commander.CommanderID, ShipID: ship.TemplateID}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("create owned ship: %v", err)
	}
	entry := orm.ConfigEntry{Category: "sharecfgdata/ship_data_template.json", Key: "7001", Data: json.RawMessage(`{"id":7001,"equip_1":[1],"equip_2":[2],"equip_3":[3],"equip_4":[],"equip_5":[],"equip_id_1":0,"equip_id_2":0,"equip_id_3":0}`)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("create config entry: %v", err)
	}

	body = []byte(`{"equipment":[{"pos":1,"equip_id":9001,"skin_id":0}]}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/9001/ships/8001/equipment", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9001/ships/8001/equipment", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var shipPayload playerShipEquipmentResponse
	if err := json.NewDecoder(response.Body).Decode(&shipPayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !shipPayload.OK || len(shipPayload.Data.Equipment) != 1 || shipPayload.Data.Equipment[0].EquipID != 9001 {
		t.Fatalf("unexpected ship equipment response")
	}
}
