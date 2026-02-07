package answer

import (
	"encoding/json"
	"testing"

	"github.com/ggmolly/belfast/internal/misc"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedConfigEntryEscort(t *testing.T, category string, key string, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: category, Key: key, Data: json.RawMessage(payload)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}

func decodeSC13302(t *testing.T, clientBuffer []byte) *protobuf.SC_13302 {
	t.Helper()
	if len(clientBuffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &clientBuffer)
	if packetID != 13302 {
		t.Fatalf("expected packet 13302, got %d", packetID)
	}
	packetSize := packets.GetPacketSize(0, &clientBuffer) + 2
	if len(clientBuffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(clientBuffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	var response protobuf.SC_13302
	if err := proto.Unmarshal(clientBuffer[payloadStart:payloadEnd], &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return &response
}

func TestEscortQuery_Type0_ReturnsAllActiveEscorts(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.EscortState{})

	state1 := orm.EscortState{
		AccountID:      client.Commander.AccountID,
		LineID:         20001,
		AwardTimestamp: 11,
		FlashTimestamp: 22,
		MapPositions:   json.RawMessage(`[{"map_id":70000,"chapter_id":101}]`),
	}
	state2 := orm.EscortState{
		AccountID:      client.Commander.AccountID,
		LineID:         20002,
		AwardTimestamp: 33,
		FlashTimestamp: 44,
		MapPositions:   json.RawMessage(`[{"map_id":70001,"chapter_id":202},{"map_id":70002,"chapter_id":203}]`),
	}
	if err := orm.GormDB.Create(&state1).Error; err != nil {
		t.Fatalf("create escort state: %v", err)
	}
	if err := orm.GormDB.Create(&state2).Error; err != nil {
		t.Fatalf("create escort state: %v", err)
	}

	payload := protobuf.CS_13301{Type: proto.Uint32(0)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := EscortQuery(&data, client); err != nil {
		t.Fatalf("EscortQuery failed: %v", err)
	}

	response := decodeSC13302(t, client.Buffer.Bytes())
	if len(response.GetEscortInfo()) != 2 {
		t.Fatalf("expected 2 escort entries, got %d", len(response.GetEscortInfo()))
	}
	if response.GetEscortInfo()[0].GetLineId() != 20001 {
		t.Fatalf("expected line_id 20001, got %d", response.GetEscortInfo()[0].GetLineId())
	}
	if response.GetEscortInfo()[0].GetAwardTimestamp() != 11 {
		t.Fatalf("expected award_timestamp 11, got %d", response.GetEscortInfo()[0].GetAwardTimestamp())
	}
	if response.GetEscortInfo()[0].GetFlashTimestamp() != 22 {
		t.Fatalf("expected flash_timestamp 22, got %d", response.GetEscortInfo()[0].GetFlashTimestamp())
	}
	if len(response.GetEscortInfo()[0].GetMap()) != 1 {
		t.Fatalf("expected 1 map entry, got %d", len(response.GetEscortInfo()[0].GetMap()))
	}
	if response.GetEscortInfo()[0].GetMap()[0].GetMapId() != 70000 {
		t.Fatalf("expected map_id 70000, got %d", response.GetEscortInfo()[0].GetMap()[0].GetMapId())
	}
	if response.GetEscortInfo()[0].GetMap()[0].GetChapterId() != 101 {
		t.Fatalf("expected chapter_id 101, got %d", response.GetEscortInfo()[0].GetMap()[0].GetChapterId())
	}

	if response.GetEscortInfo()[1].GetLineId() != 20002 {
		t.Fatalf("expected line_id 20002, got %d", response.GetEscortInfo()[1].GetLineId())
	}
	if len(response.GetEscortInfo()[1].GetMap()) != 2 {
		t.Fatalf("expected 2 map entries, got %d", len(response.GetEscortInfo()[1].GetMap()))
	}
}

func TestEscortQuery_NoState_ReturnsEmptyList(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.EscortState{})

	payload := protobuf.CS_13301{Type: proto.Uint32(0)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := EscortQuery(&data, client); err != nil {
		t.Fatalf("EscortQuery failed: %v", err)
	}

	response := decodeSC13302(t, client.Buffer.Bytes())
	if len(response.GetEscortInfo()) != 0 {
		t.Fatalf("expected empty escort_info, got %d", len(response.GetEscortInfo()))
	}
	if len(response.GetDropList()) != 0 {
		t.Fatalf("expected empty drop_list, got %d", len(response.GetDropList()))
	}
}

func TestEscortConfigLoad_LoadsFromConfigEntries(t *testing.T) {
	_ = setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntryEscort(t, "ShareCfg/escort_template.json", "20001", `{"id":20001,"gardroad_reward":[]}`)
	seedConfigEntryEscort(t, "ShareCfg/escort_map_template.json", "70000", `{"id":70000,"refresh_time":21600,"drop_by_warn":[1,1,1,1,1],"escort_id_list":[1,2,3]}`)

	config, err := misc.GetEscortConfig()
	if err != nil {
		t.Fatalf("GetEscortConfig failed: %v", err)
	}
	if len(config.Templates) != 1 {
		t.Fatalf("expected 1 escort template, got %d", len(config.Templates))
	}
	if _, ok := config.Templates[20001]; !ok {
		t.Fatalf("expected escort template id 20001")
	}
	if len(config.Maps) != 1 {
		t.Fatalf("expected 1 escort map template, got %d", len(config.Maps))
	}
	if config.Maps[0].RefreshTime != 21600 {
		t.Fatalf("expected refresh_time 21600, got %d", config.Maps[0].RefreshTime)
	}
}

func TestEscortStatePersistence_SaveAndLoad(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.EscortState{})

	state := orm.EscortState{
		AccountID:      client.Commander.AccountID,
		LineID:         20001,
		AwardTimestamp: 123,
		FlashTimestamp: 456,
		MapPositions:   json.RawMessage(`[{"map_id":70000,"chapter_id":101}]`),
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("create escort state: %v", err)
	}
	infos, err := misc.LoadEscortState(client.Commander.AccountID)
	if err != nil {
		t.Fatalf("LoadEscortState failed: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 escort info, got %d", len(infos))
	}
	if infos[0].GetLineId() != 20001 {
		t.Fatalf("expected line_id 20001, got %d", infos[0].GetLineId())
	}
	if infos[0].GetAwardTimestamp() != 123 {
		t.Fatalf("expected award_timestamp 123, got %d", infos[0].GetAwardTimestamp())
	}
	if infos[0].GetFlashTimestamp() != 456 {
		t.Fatalf("expected flash_timestamp 456, got %d", infos[0].GetFlashTimestamp())
	}
}
