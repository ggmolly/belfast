package answer

import (
	"reflect"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChapterBattleResultRequestReturnsState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	seedChapterTrackingConfig(t)

	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start tracking: %v", err)
	}

	payload := protobuf.CS_13106{Arg: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChapterBattleResultRequest(&buffer, client); err != nil {
		t.Fatalf("chapter battle result request failed: %v", err)
	}
	var response protobuf.SC_13105
	decodeResponse(t, client, &response)

	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if !reflect.DeepEqual(response.GetMapUpdate(), current.GetCellList()) {
		t.Fatalf("expected map update to match current state")
	}
	if !reflect.DeepEqual(response.GetAiList(), current.GetAiList()) {
		t.Fatalf("expected ai list to match current state")
	}
	if !reflect.DeepEqual(response.GetBuffList(), current.GetBuffList()) {
		t.Fatalf("expected buff list to match current state")
	}
	if !reflect.DeepEqual(response.GetCellFlagList(), current.GetCellFlagList()) {
		t.Fatalf("expected cell flag list to match current state")
	}
}

func TestChapterBattleResultRequestWithoutState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ChapterState{})

	payload := protobuf.CS_13106{Arg: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChapterBattleResultRequest(&buffer, client); err != nil {
		t.Fatalf("chapter battle result request failed: %v", err)
	}
	var response protobuf.SC_13105
	decodeResponse(t, client, &response)
	if len(response.GetMapUpdate()) != 0 {
		t.Fatalf("expected empty map update")
	}
	if len(response.GetAiList()) != 0 {
		t.Fatalf("expected empty ai list")
	}
	if len(response.GetBuffList()) != 0 {
		t.Fatalf("expected empty buff list")
	}
	if len(response.GetCellFlagList()) != 0 {
		t.Fatalf("expected empty cell flag list")
	}
	if len(response.GetAddFlagList()) != 0 {
		t.Fatalf("expected empty add flag list")
	}
	if len(response.GetDelFlagList()) != 0 {
		t.Fatalf("expected empty del flag list")
	}
}
