package answer

import (
	"errors"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func TestExerciseEnemies_UsesFleet1Fallback_WhenNoPersistedFleet(t *testing.T) {
	client := setupExerciseTest(t)
	buf := []byte{}
	if _, _, err := ExerciseEnemies(&buf, client); err != nil {
		t.Fatalf("ExerciseEnemies failed: %v", err)
	}

	var resp protobuf.SC_18002
	decodePacketMessage(t, client, 18002, &resp)
	if got := resp.GetVanguardShipIdList(); len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("expected vanguard [1 2 3], got %v", got)
	}
	if got := resp.GetMainShipIdList(); len(got) != 3 || got[0] != 4 || got[1] != 5 || got[2] != 6 {
		t.Fatalf("expected main [4 5 6], got %v", got)
	}
}

func TestUpdateExerciseFleet_PersistsAndReflectsInSeasonInfo(t *testing.T) {
	client := setupExerciseTest(t)

	request := protobuf.CS_18008{
		VanguardShipIdList: []uint32{1, 2, 3},
		MainShipIdList:     []uint32{4, 5, 6},
	}
	buf, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := UpdateExerciseFleet(&buf, client); err != nil {
		t.Fatalf("UpdateExerciseFleet failed: %v", err)
	}

	var updateResp protobuf.SC_18009
	decodePacketMessage(t, client, 18009, &updateResp)
	if updateResp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", updateResp.GetResult())
	}

	stored, err := orm.GetExerciseFleet(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("get persisted exercise fleet: %v", err)
	}
	if got := orm.ToUint32List(stored.VanguardShipIDs); len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("expected stored vanguard [1 2 3], got %v", got)
	}
	if got := orm.ToUint32List(stored.MainShipIDs); len(got) != 3 || got[0] != 4 || got[1] != 5 || got[2] != 6 {
		t.Fatalf("expected stored main [4 5 6], got %v", got)
	}

	client.Buffer.Reset()
	empty := []byte{}
	if _, _, err := ExerciseEnemies(&empty, client); err != nil {
		t.Fatalf("ExerciseEnemies failed: %v", err)
	}
	var seasonResp protobuf.SC_18002
	decodePacketMessage(t, client, 18002, &seasonResp)
	if got := seasonResp.GetVanguardShipIdList(); len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("expected vanguard [1 2 3], got %v", got)
	}
	if got := seasonResp.GetMainShipIdList(); len(got) != 3 || got[0] != 4 || got[1] != 5 || got[2] != 6 {
		t.Fatalf("expected main [4 5 6], got %v", got)
	}
}

func TestUpdateExerciseFleet_RejectsTooManyShips(t *testing.T) {
	client := setupExerciseTest(t)

	request := protobuf.CS_18008{
		VanguardShipIdList: []uint32{1, 2, 3, 4},
		MainShipIdList:     []uint32{5},
	}
	buf, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := UpdateExerciseFleet(&buf, client); err != nil {
		t.Fatalf("UpdateExerciseFleet failed: %v", err)
	}

	var resp protobuf.SC_18009
	decodePacketMessage(t, client, 18009, &resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
	if _, err := orm.GetExerciseFleet(orm.GormDB, client.Commander.CommanderID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no persisted fleet")
	}
}

func TestUpdateExerciseFleet_RejectsShipNotOwned(t *testing.T) {
	client := setupExerciseTest(t)

	request := protobuf.CS_18008{
		VanguardShipIdList: []uint32{1},
		MainShipIdList:     []uint32{999},
	}
	buf, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := UpdateExerciseFleet(&buf, client); err != nil {
		t.Fatalf("UpdateExerciseFleet failed: %v", err)
	}

	var resp protobuf.SC_18009
	decodePacketMessage(t, client, 18009, &resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
	if _, err := orm.GetExerciseFleet(orm.GormDB, client.Commander.CommanderID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no persisted fleet")
	}
}
