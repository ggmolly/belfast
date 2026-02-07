package answer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedMedalTemplate(t *testing.T, id uint32, next uint32, targetNum uint32, countInherit uint32) {
	t.Helper()
	payload := fmt.Sprintf(`{"id":%d,"next":%d,"target_num":%d,"count_inherit":%d}`, id, next, targetNum, countInherit)
	seedConfigEntry(t, medalTemplateCategory, strconv.FormatUint(uint64(id), 10), payload)
}

func TestTrophyClaim17301ClaimsAndPersistsTimestamp(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderTrophyProgress{})

	seedMedalTemplate(t, 1001, 0, 10, 0)
	row := orm.CommanderTrophyProgress{CommanderID: client.Commander.CommanderID, TrophyID: 1001, Progress: 10, Timestamp: 0}
	if err := orm.GormDB.Create(&row).Error; err != nil {
		t.Fatalf("seed trophy progress: %v", err)
	}

	payload := protobuf.CS_17301{Id: proto.Uint32(1001)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := TrophyClaim17301(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17302
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetTimestamp() == 0 {
		t.Fatalf("expected timestamp to be set")
	}

	stored, err := orm.GetCommanderTrophyProgress(orm.GormDB, client.Commander.CommanderID, 1001)
	if err != nil {
		t.Fatalf("load stored trophy: %v", err)
	}
	if stored.Timestamp != response.GetTimestamp() {
		t.Fatalf("expected stored timestamp %d, got %d", response.GetTimestamp(), stored.Timestamp)
	}
}

func TestTrophyClaim17301CreatesProgressRowWhenMissing(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderTrophyProgress{})

	seedMedalTemplate(t, 6001, 0, 10, 0)

	payload := protobuf.CS_17301{Id: proto.Uint32(6001)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := TrophyClaim17301(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17302
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetTimestamp() == 0 {
		t.Fatalf("expected timestamp to be set")
	}

	stored, err := orm.GetCommanderTrophyProgress(orm.GormDB, client.Commander.CommanderID, 6001)
	if err != nil {
		t.Fatalf("load stored trophy: %v", err)
	}
	if stored.Progress != 10 {
		t.Fatalf("expected stored progress 10, got %d", stored.Progress)
	}
	if stored.Timestamp != response.GetTimestamp() {
		t.Fatalf("expected stored timestamp %d, got %d", response.GetTimestamp(), stored.Timestamp)
	}
}

func TestTrophyClaim17301UnlocksNextWhenMissing(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderTrophyProgress{})

	seedMedalTemplate(t, 2001, 2002, 1, 0)
	seedMedalTemplate(t, 2002, 0, 999, 0)
	row := orm.CommanderTrophyProgress{CommanderID: client.Commander.CommanderID, TrophyID: 2001, Progress: 1, Timestamp: 0}
	if err := orm.GormDB.Create(&row).Error; err != nil {
		t.Fatalf("seed trophy progress: %v", err)
	}

	payload := protobuf.CS_17301{Id: proto.Uint32(2001)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := TrophyClaim17301(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17302
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetNext()) != 1 {
		t.Fatalf("expected 1 next trophy, got %d", len(response.GetNext()))
	}
	next := response.GetNext()[0]
	if next.GetId() != 2002 {
		t.Fatalf("expected next id 2002, got %d", next.GetId())
	}
	if next.GetTimestamp() != 0 {
		t.Fatalf("expected next timestamp 0")
	}
	if next.GetProgress() != 0 {
		t.Fatalf("expected next progress 0")
	}

	stored, err := orm.GetCommanderTrophyProgress(orm.GormDB, client.Commander.CommanderID, 2002)
	if err != nil {
		t.Fatalf("load stored next trophy: %v", err)
	}
	if stored.Timestamp != 0 {
		t.Fatalf("expected stored next timestamp 0")
	}
}

func TestTrophyClaim17301NextInheritsProgressWhenConfigured(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderTrophyProgress{})

	seedMedalTemplate(t, 3001, 3002, 1, 3002)
	seedMedalTemplate(t, 3002, 0, 999, 0)
	row := orm.CommanderTrophyProgress{CommanderID: client.Commander.CommanderID, TrophyID: 3001, Progress: 77, Timestamp: 0}
	if err := orm.GormDB.Create(&row).Error; err != nil {
		t.Fatalf("seed trophy progress: %v", err)
	}

	payload := protobuf.CS_17301{Id: proto.Uint32(3001)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := TrophyClaim17301(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17302
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetNext()) != 1 {
		t.Fatalf("expected 1 next trophy")
	}
	if response.GetNext()[0].GetProgress() != 77 {
		t.Fatalf("expected inherited progress 77, got %d", response.GetNext()[0].GetProgress())
	}

	stored, err := orm.GetCommanderTrophyProgress(orm.GormDB, client.Commander.CommanderID, 3002)
	if err != nil {
		t.Fatalf("load stored next trophy: %v", err)
	}
	if stored.Progress != 77 {
		t.Fatalf("expected stored inherited progress 77, got %d", stored.Progress)
	}
}

func TestTrophyClaim17301RejectsAlreadyClaimedWithoutMutation(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderTrophyProgress{})

	seedMedalTemplate(t, 4001, 0, 1, 0)
	row := orm.CommanderTrophyProgress{CommanderID: client.Commander.CommanderID, TrophyID: 4001, Progress: 1, Timestamp: 123}
	if err := orm.GormDB.Create(&row).Error; err != nil {
		t.Fatalf("seed trophy progress: %v", err)
	}

	payload := protobuf.CS_17301{Id: proto.Uint32(4001)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := TrophyClaim17301(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17302
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
	stored, err := orm.GetCommanderTrophyProgress(orm.GormDB, client.Commander.CommanderID, 4001)
	if err != nil {
		t.Fatalf("load stored trophy: %v", err)
	}
	if stored.Timestamp != 123 {
		t.Fatalf("expected timestamp unchanged")
	}
}

func TestTrophyClaim17301RejectsInsufficientProgressWithoutMutation(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderTrophyProgress{})

	seedMedalTemplate(t, 5001, 0, 10, 0)
	row := orm.CommanderTrophyProgress{CommanderID: client.Commander.CommanderID, TrophyID: 5001, Progress: 9, Timestamp: 0}
	if err := orm.GormDB.Create(&row).Error; err != nil {
		t.Fatalf("seed trophy progress: %v", err)
	}

	payload := protobuf.CS_17301{Id: proto.Uint32(5001)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := TrophyClaim17301(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17302
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
	stored, err := orm.GetCommanderTrophyProgress(orm.GormDB, client.Commander.CommanderID, 5001)
	if err != nil {
		t.Fatalf("load stored trophy: %v", err)
	}
	if stored.Timestamp != 0 {
		t.Fatalf("expected timestamp unchanged")
	}
}
