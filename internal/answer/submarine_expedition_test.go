package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	expeditionDailyTemplateCategory = "ShareCfg/expedition_daily_template.json"
	expeditionDataTemplateCategory  = "ShareCfg/expedition_data_template.json"
)

func decodeSC13402(t *testing.T, clientBuffer []byte) protobuf.SC_13402 {
	t.Helper()
	buffer := clientBuffer
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != 13402 {
		t.Fatalf("expected packet 13402, got %d", packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	var resp protobuf.SC_13402
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func seedSubmarineExpeditionConfig(t *testing.T) {
	t.Helper()
	seedConfigEntry(t, expeditionDailyTemplateCategory, "501", `{"id":501,"expedition_and_lv_limit_list":[[1000,35],[1001,45],[1002,55],[1003,65],[1004,75],[1005,95]]}`)
	seedConfigEntry(t, expeditionDataTemplateCategory, "1000", `{"id":1000,"type":15}`)
	seedConfigEntry(t, expeditionDataTemplateCategory, "1001", `{"id":1001,"type":15}`)
	seedConfigEntry(t, expeditionDataTemplateCategory, "1002", `{"id":1002,"type":15}`)
	seedConfigEntry(t, expeditionDataTemplateCategory, "1003", `{"id":1003,"type":15}`)
	seedConfigEntry(t, expeditionDataTemplateCategory, "1004", `{"id":1004,"type":15}`)
	seedConfigEntry(t, expeditionDataTemplateCategory, "1005", `{"id":1005,"type":15}`)
}

func TestGetSubmarineExpeditionInfoBuildsResponseAndFiltersByLevel(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.SubmarineExpeditionState{})
	seedSubmarineExpeditionConfig(t)
	client.Commander.Level = 35

	now := time.Now()
	weekStart := weekStartMondayUTC(now)
	state := orm.SubmarineExpeditionState{
		CommanderID:        client.Commander.CommanderID,
		LastRefreshTime:    uint32(weekStart.Unix()),
		WeeklyRefreshCount: 1,
		OverallProgress:    7,
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed state: %v", err)
	}

	payload := protobuf.CS_13401{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := GetSubmarineExpeditionInfo(&buf, client); err != nil {
		t.Fatalf("GetSubmarineExpeditionInfo failed: %v", err)
	}

	resp := decodeSC13402(t, client.Buffer.Bytes())
	expectedNext := uint32(nextWeeklyResetUTC(weekStart).Unix())
	if resp.GetNextRefreshTime() != expectedNext {
		t.Fatalf("expected next refresh time %d, got %d", expectedNext, resp.GetNextRefreshTime())
	}
	if resp.GetRefreshCount() != 3 {
		t.Fatalf("expected refresh count 3, got %d", resp.GetRefreshCount())
	}
	if resp.GetProgress() != 7 {
		t.Fatalf("expected progress 7, got %d", resp.GetProgress())
	}
	if len(resp.GetChapterList()) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(resp.GetChapterList()))
	}
	if got := resp.GetChapterList()[0].GetChapterId(); got != 1000 {
		t.Fatalf("expected chapter 1000, got %d", got)
	}
	if got := resp.GetChapterList()[0].GetIndex(); got != 0 {
		t.Fatalf("expected index 0, got %d", got)
	}
}

func TestGetSubmarineExpeditionInfoRefreshCountClampsAtZero(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.SubmarineExpeditionState{})
	seedSubmarineExpeditionConfig(t)
	client.Commander.Level = 120

	weekStart := weekStartMondayUTC(time.Now())
	state := orm.SubmarineExpeditionState{CommanderID: client.Commander.CommanderID, LastRefreshTime: uint32(weekStart.Unix()), WeeklyRefreshCount: 9}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed state: %v", err)
	}
	payload := protobuf.CS_13401{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := GetSubmarineExpeditionInfo(&buf, client); err != nil {
		t.Fatalf("GetSubmarineExpeditionInfo failed: %v", err)
	}
	resp := decodeSC13402(t, client.Buffer.Bytes())
	if resp.GetRefreshCount() != 0 {
		t.Fatalf("expected refresh count 0, got %d", resp.GetRefreshCount())
	}
}

func TestGetSubmarineExpeditionInfoResetsWeeklyCountOnNewWeek(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.SubmarineExpeditionState{})
	seedSubmarineExpeditionConfig(t)
	client.Commander.Level = 120

	now := time.Now()
	weekStart := weekStartMondayUTC(now)
	state := orm.SubmarineExpeditionState{
		CommanderID:        client.Commander.CommanderID,
		LastRefreshTime:    uint32(weekStart.Add(-24 * time.Hour).Unix()),
		WeeklyRefreshCount: 2,
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed state: %v", err)
	}
	payload := protobuf.CS_13401{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := GetSubmarineExpeditionInfo(&buf, client); err != nil {
		t.Fatalf("GetSubmarineExpeditionInfo failed: %v", err)
	}
	resp := decodeSC13402(t, client.Buffer.Bytes())
	if resp.GetRefreshCount() != 4 {
		t.Fatalf("expected refresh count 4 after reset, got %d", resp.GetRefreshCount())
	}
	stored, err := orm.GetSubmarineState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if stored.WeeklyRefreshCount != 0 {
		t.Fatalf("expected weekly refresh count reset, got %d", stored.WeeklyRefreshCount)
	}
	if stored.LastRefreshTime != uint32(weekStart.Unix()) {
		t.Fatalf("expected last refresh time to update")
	}
}
