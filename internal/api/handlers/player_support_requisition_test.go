package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
)

func TestPlayerSupportRequisitionEndpoints(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9400)
	if err := orm.GormDB.Unscoped().Where("commander_id = ?", commanderID).Delete(&orm.Commander{}).Error; err != nil {
		t.Fatalf("clear commander: %v", err)
	}
	t.Cleanup(func() {
		orm.GormDB.Unscoped().Where("commander_id = ?", commanderID).Delete(&orm.Commander{})
		orm.GormDB.Where("category = ? AND key = ?", "ShareCfg/gameset.json", "supports_config").Delete(&orm.ConfigEntry{})
	})
	if err := orm.GormDB.Where("category = ? AND key = ?", "ShareCfg/gameset.json", "supports_config").Delete(&orm.ConfigEntry{}).Error; err != nil {
		t.Fatalf("clear supports_config: %v", err)
	}
	entry := orm.ConfigEntry{
		Category: "ShareCfg/gameset.json",
		Key:      "supports_config",
		Data:     json.RawMessage(`{"key_value":0,"description":[6,[[2,5400],[3,3200],[4,1000],[5,400]],999]}`),
	}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("create supports_config: %v", err)
	}
	oldMonth := orm.SupportRequisitionMonth(time.Now().AddDate(0, -1, 0))
	commander := orm.Commander{
		CommanderID:             commanderID,
		AccountID:               1,
		Level:                   1,
		Exp:                     0,
		Name:                    "Support Counter Tester",
		LastLogin:               time.Now().UTC(),
		SupportRequisitionMonth: oldMonth,
		SupportRequisitionCount: 5,
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/9400/support-requisition", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var getResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Month uint32 `json:"month"`
			Count uint32 `json:"count"`
			Cap   uint32 `json:"cap"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&getResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !getResponse.OK {
		t.Fatalf("expected ok response")
	}
	currentMonth := orm.SupportRequisitionMonth(time.Now())
	if getResponse.Data.Month != currentMonth {
		t.Fatalf("expected month %d, got %d", currentMonth, getResponse.Data.Month)
	}
	if getResponse.Data.Count != 0 {
		t.Fatalf("expected count 0, got %d", getResponse.Data.Count)
	}
	if getResponse.Data.Cap != 999 {
		t.Fatalf("expected cap 999, got %d", getResponse.Data.Cap)
	}

	var updated orm.Commander
	if err := orm.GormDB.First(&updated, commanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if updated.SupportRequisitionCount != 0 || updated.SupportRequisitionMonth != currentMonth {
		t.Fatalf("expected counters reset")
	}

	updated.SupportRequisitionCount = 7
	if err := orm.GormDB.Save(&updated).Error; err != nil {
		t.Fatalf("save commander: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/9400/support-requisition/reset", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var resetResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Month uint32 `json:"month"`
			Count uint32 `json:"count"`
			Cap   uint32 `json:"cap"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&resetResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resetResponse.Data.Count != 0 {
		t.Fatalf("expected count 0, got %d", resetResponse.Data.Count)
	}
}
