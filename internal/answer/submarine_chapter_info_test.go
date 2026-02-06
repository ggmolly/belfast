package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func decodeSC13404(t *testing.T, clientBuffer []byte) protobuf.SC_13404 {
	t.Helper()
	buffer := clientBuffer
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != 13404 {
		t.Fatalf("expected packet 13404, got %d", packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	var resp protobuf.SC_13404
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func startOfDayForTest(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func TestSubmarineChapterInfoSuccess(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.RemasterState{})

	refreshTime := uint32(1_000_000)
	escortIDs := []uint32{14001, 14002, 14003}
	seedConfigEntry(t, escortMapTemplateCategory, "1", `{"id":1,"refresh_time":1000000,"escort_id_list":[14001,14002,14003]}`)

	now := time.Now()
	state := orm.RemasterState{
		CommanderID:      client.Commander.CommanderID,
		DailyCount:       7,
		LastDailyResetAt: startOfDayForTest(now.Add(-24 * time.Hour)),
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed remaster state: %v", err)
	}

	payload := protobuf.CS_13403{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SubmarineChapterInfo(&buf, client); err != nil {
		t.Fatalf("SubmarineChapterInfo failed: %v", err)
	}

	resp := decodeSC13404(t, client.Buffer.Bytes())
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}
	if resp.GetChapterId() == nil {
		t.Fatalf("expected chapter_id to be set")
	}

	// With a large refresh_time, the active slot is stable for the test duration.
	slotIndex := uint32((now.Unix() / int64(refreshTime)) % int64(len(escortIDs)))
	expectedChapter := escortIDs[slotIndex]
	expectedActiveAt := uint32(now.Unix() - (now.Unix() % int64(refreshTime)))
	if got := resp.GetChapterId().GetChapterId(); got != expectedChapter {
		t.Fatalf("expected chapter_id %d, got %d", expectedChapter, got)
	}
	if got := resp.GetChapterId().GetIndex(); got != slotIndex+1 {
		t.Fatalf("expected index %d, got %d", slotIndex+1, got)
	}
	if got := resp.GetChapterId().GetActiveTime(); got != expectedActiveAt {
		t.Fatalf("expected active_time %d, got %d", expectedActiveAt, got)
	}

	var updated orm.RemasterState
	if err := orm.GormDB.First(&updated, "commander_id = ?", client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load remaster state: %v", err)
	}
	if updated.DailyCount != 0 {
		t.Fatalf("expected DailyCount reset to 0, got %d", updated.DailyCount)
	}
	if !updated.LastDailyResetAt.Equal(startOfDayForTest(now)) {
		t.Fatalf("expected LastDailyResetAt %v, got %v", startOfDayForTest(now), updated.LastDailyResetAt)
	}
}

func TestSubmarineChapterInfoMissingConfigReturnsErrorResult(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.RemasterState{})

	payload := protobuf.CS_13403{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SubmarineChapterInfo(&buf, client); err != nil {
		t.Fatalf("SubmarineChapterInfo failed: %v", err)
	}

	resp := decodeSC13404(t, client.Buffer.Bytes())
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
	if resp.GetChapterId() != nil {
		t.Fatalf("expected chapter_id to be unset on error")
	}
}

func TestSubmarineChapterInfoTypeNonZeroNoops(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.RemasterState{})
	seedConfigEntry(t, escortMapTemplateCategory, "1", `{"id":1,"refresh_time":1000000,"escort_id_list":[14001]}`)

	payload := protobuf.CS_13403{Type: proto.Uint32(1)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SubmarineChapterInfo(&buf, client); err != nil {
		t.Fatalf("SubmarineChapterInfo failed: %v", err)
	}

	resp := decodeSC13404(t, client.Buffer.Bytes())
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}
