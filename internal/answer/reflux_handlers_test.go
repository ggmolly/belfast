package answer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedRefluxConfig(t *testing.T) {
	t.Helper()
	clearTable(t, &orm.ConfigEntry{})
	seedConfigEntry(t, returnSignTemplateCategory, "1", `{"id":1,"level":[[1,100]],"award_display":[[[1,1,100]]]}`)
	seedConfigEntry(t, returnPtTemplateCategory, "1", `{"id":1,"level":[[1,100]],"pt_require":50,"virtual_item":59616,"award_display":[[1,1,200]]}`)
	seedConfigEntry(t, activityTemplateCategory, "30094", `{"id":30094,"type":42,"config_id":0,"time":"stop","config_client":"","config_data":[10,30,90]}`)
}

func TestRefluxRequestDataActivates(t *testing.T) {
	client := setupHandlerCommander(t)
	seedRefluxConfig(t)
	clearTable(t, &orm.RefluxState{})
	client.Commander.Level = 20
	if err := orm.GormDB.Save(client.Commander).Error; err != nil {
		t.Fatalf("save commander: %v", err)
	}
	client.PreviousLoginAt = time.Now().UTC().Add(-40 * 24 * time.Hour)
	buffer, err := proto.Marshal(&protobuf.CS_11751{Type: proto.Uint32(0)})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := RefluxRequestData(&buffer, client); err != nil {
		t.Fatalf("reflux request data failed: %v", err)
	}
	var response protobuf.SC_11752
	decodeResponse(t, client, &response)
	if response.GetActive() != 1 {
		t.Fatalf("expected active 1")
	}
	if response.GetReturnLv() != 20 {
		t.Fatalf("expected return lv 20")
	}
	if response.GetLastOfflineTime() != uint32(client.PreviousLoginAt.Unix()) {
		t.Fatalf("unexpected last offline time")
	}
}

func TestRefluxRequestDataIneligible(t *testing.T) {
	client := setupHandlerCommander(t)
	seedRefluxConfig(t)
	clearTable(t, &orm.RefluxState{})
	client.Commander.Level = 20
	if err := orm.GormDB.Save(client.Commander).Error; err != nil {
		t.Fatalf("save commander: %v", err)
	}
	client.PreviousLoginAt = time.Now().UTC().Add(-24 * time.Hour)
	buffer, err := proto.Marshal(&protobuf.CS_11751{Type: proto.Uint32(0)})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := RefluxRequestData(&buffer, client); err != nil {
		t.Fatalf("reflux request data failed: %v", err)
	}
	var response protobuf.SC_11752
	decodeResponse(t, client, &response)
	if response.GetActive() != 0 {
		t.Fatalf("expected active 0")
	}
}

func TestRefluxRequestDataMissingTemplates(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	buffer, err := proto.Marshal(&protobuf.CS_11751{Type: proto.Uint32(0)})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := RefluxRequestData(&buffer, client); err != nil {
		t.Fatalf("reflux request data failed: %v", err)
	}
	var response protobuf.SC_11752
	decodeResponse(t, client, &response)
	if response.GetActive() != 0 {
		t.Fatalf("expected active 0")
	}
}

func TestRefluxRequestDataExpires(t *testing.T) {
	client := setupHandlerCommander(t)
	seedRefluxConfig(t)
	clearTable(t, &orm.RefluxState{})
	expired := orm.RefluxState{
		CommanderID: client.Commander.CommanderID,
		Active:      1,
		ReturnTime:  uint32(time.Now().UTC().Add(-48 * time.Hour).Unix()),
	}
	if err := orm.GormDB.Create(&expired).Error; err != nil {
		t.Fatalf("seed reflux state: %v", err)
	}
	buffer, err := proto.Marshal(&protobuf.CS_11751{Type: proto.Uint32(0)})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := RefluxRequestData(&buffer, client); err != nil {
		t.Fatalf("reflux request data failed: %v", err)
	}
	var response protobuf.SC_11752
	decodeResponse(t, client, &response)
	if response.GetActive() != 0 {
		t.Fatalf("expected active 0")
	}
	state, err := orm.GetOrCreateRefluxState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load reflux state: %v", err)
	}
	if state.Active != 0 {
		t.Fatalf("expected persisted active 0")
	}
}

func TestRefluxSignSuccess(t *testing.T) {
	client := setupHandlerCommander(t)
	seedRefluxConfig(t)
	clearTable(t, &orm.RefluxState{})
	state := orm.RefluxState{
		CommanderID: client.Commander.CommanderID,
		Active:      1,
		ReturnLv:    20,
		ReturnTime:  uint32(time.Now().UTC().Unix()),
		SignCnt:     0,
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed reflux state: %v", err)
	}
	buffer, err := proto.Marshal(&protobuf.CS_11753{Type: proto.Uint32(0)})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := RefluxSign(&buffer, client); err != nil {
		t.Fatalf("reflux sign failed: %v", err)
	}
	var response protobuf.SC_11754
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	if len(response.GetAwardList()) != 1 {
		t.Fatalf("expected award list")
	}
	updated, err := orm.GetOrCreateRefluxState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load reflux state: %v", err)
	}
	if updated.SignCnt != 1 || updated.SignLastTime == 0 {
		t.Fatalf("expected sign progress update")
	}
	if resource := client.Commander.OwnedResourcesMap[1]; resource == nil || resource.Amount < 100 {
		t.Fatalf("expected resource drop applied")
	}
}

func TestRefluxSignAlreadySigned(t *testing.T) {
	client := setupHandlerCommander(t)
	seedRefluxConfig(t)
	clearTable(t, &orm.RefluxState{})
	now := uint32(time.Now().UTC().Unix())
	state := orm.RefluxState{
		CommanderID:  client.Commander.CommanderID,
		Active:       1,
		ReturnLv:     20,
		ReturnTime:   now,
		SignCnt:      0,
		SignLastTime: now,
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed reflux state: %v", err)
	}
	buffer, err := proto.Marshal(&protobuf.CS_11753{Type: proto.Uint32(0)})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := RefluxSign(&buffer, client); err != nil {
		t.Fatalf("reflux sign failed: %v", err)
	}
	var response protobuf.SC_11754
	decodeResponse(t, client, &response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1")
	}
	if len(response.GetAwardList()) != 0 {
		t.Fatalf("expected empty award list")
	}
}

func TestRefluxGetPTAwardSuccess(t *testing.T) {
	client := setupHandlerCommander(t)
	seedRefluxConfig(t)
	clearTable(t, &orm.RefluxState{})
	seedHandlerCommanderItem(t, client, 59616, 100)
	state := orm.RefluxState{
		CommanderID: client.Commander.CommanderID,
		Active:      1,
		ReturnLv:    20,
		ReturnTime:  uint32(time.Now().UTC().Unix()),
		PtStage:     0,
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed reflux state: %v", err)
	}
	buffer, err := proto.Marshal(&protobuf.CS_11755{Type: proto.Uint32(0)})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := RefluxGetPTAward(&buffer, client); err != nil {
		t.Fatalf("reflux pt award failed: %v", err)
	}
	var response protobuf.SC_11756
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	if len(response.GetAwardList()) != 1 {
		t.Fatalf("expected award list")
	}
	updated, err := orm.GetOrCreateRefluxState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load reflux state: %v", err)
	}
	if updated.PtStage != 1 {
		t.Fatalf("expected pt stage increment")
	}
}

func TestRefluxGetPTAwardInsufficientPt(t *testing.T) {
	client := setupHandlerCommander(t)
	seedRefluxConfig(t)
	clearTable(t, &orm.RefluxState{})
	seedHandlerCommanderItem(t, client, 59616, 10)
	state := orm.RefluxState{
		CommanderID: client.Commander.CommanderID,
		Active:      1,
		ReturnLv:    20,
		ReturnTime:  uint32(time.Now().UTC().Unix()),
		PtStage:     0,
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed reflux state: %v", err)
	}
	buffer, err := proto.Marshal(&protobuf.CS_11755{Type: proto.Uint32(0)})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := RefluxGetPTAward(&buffer, client); err != nil {
		t.Fatalf("reflux pt award failed: %v", err)
	}
	var response protobuf.SC_11756
	decodeResponse(t, client, &response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1")
	}
}

func TestRefluxConfigParsing(t *testing.T) {
	seedRefluxConfig(t)
	entries, err := orm.ListConfigEntries(orm.GormDB, activityTemplateCategory)
	if err != nil {
		t.Fatalf("list config entries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected activity template entries")
	}
	var cfg refluxEligibilityConfig
	for _, entry := range entries {
		var template activityTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			t.Fatalf("unmarshal activity template: %v", err)
		}
		if template.Type != activityTypeReflux {
			continue
		}
		cfg, err = parseRefluxEligibilityConfig(template.ConfigData)
		if err != nil {
			t.Fatalf("parse reflux config: %v", err)
		}
		break
	}
	if cfg.MinLevel != 10 || cfg.MinOfflineDays != 30 || cfg.MaxOfflineDays != 90 {
		t.Fatalf("unexpected reflux config values")
	}
}
