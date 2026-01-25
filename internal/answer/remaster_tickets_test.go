package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestRemasterTicketsClaimSuccess(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.RemasterState{})
	seedConfigEntry(t, "ShareCfg/gameset.json", "reactivity_ticket_daily", `{"key_value":4}`)
	seedConfigEntry(t, "ShareCfg/gameset.json", "reactivity_ticket_max", `{"key_value":6}`)

	payload := protobuf.CS_13503{Type: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemasterTickets(&buffer, client); err != nil {
		t.Fatalf("remaster tickets failed: %v", err)
	}

	var response protobuf.SC_13504
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result")
	}
	var state orm.RemasterState
	if err := orm.GormDB.First(&state, "commander_id = ?", client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load remaster state: %v", err)
	}
	if state.TicketCount != 4 || state.DailyCount != 4 {
		t.Fatalf("unexpected state: %+v", state)
	}
}

func TestRemasterTicketsClaimFailsWhenAlreadyUsed(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.RemasterState{})
	seedConfigEntry(t, "ShareCfg/gameset.json", "reactivity_ticket_daily", `{"key_value":4}`)
	seedConfigEntry(t, "ShareCfg/gameset.json", "reactivity_ticket_max", `{"key_value":6}`)
	state := orm.RemasterState{
		CommanderID:      client.Commander.CommanderID,
		TicketCount:      2,
		DailyCount:       4,
		LastDailyResetAt: time.Now(),
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed remaster state: %v", err)
	}

	payload := protobuf.CS_13503{Type: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemasterTickets(&buffer, client); err != nil {
		t.Fatalf("remaster tickets failed: %v", err)
	}

	var response protobuf.SC_13504
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}
	var saved orm.RemasterState
	if err := orm.GormDB.First(&saved, "commander_id = ?", client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load remaster state: %v", err)
	}
	if saved.TicketCount != 2 {
		t.Fatalf("expected ticket count to remain 2, got %d", saved.TicketCount)
	}
}
