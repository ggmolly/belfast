package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestRemasterAwardReceiveGrantsDrop(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	clearTable(t, &orm.RemasterProgress{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderItem{})
	seedConfigEntry(t, "ShareCfg/re_map_template.json", "1", `{"id":1,"drop_gain":[[1001,2,2001,3]]}`)
	progress := orm.RemasterProgress{
		CommanderID: client.Commander.CommanderID,
		ChapterID:   1001,
		Pos:         1,
		Count:       3,
		Received:    false,
	}
	if err := orm.GormDB.Create(&progress).Error; err != nil {
		t.Fatalf("seed remaster progress: %v", err)
	}

	payload := protobuf.CS_13507{ChapterId: proto.Uint32(1001), Pos: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemasterAwardReceive(&buffer, client); err != nil {
		t.Fatalf("remaster award receive failed: %v", err)
	}

	var response protobuf.SC_13508
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result")
	}
	if len(response.GetDropList()) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(response.GetDropList()))
	}
	drop := response.GetDropList()[0]
	if drop.GetType() != 2 || drop.GetId() != 2001 || drop.GetNumber() != 1 {
		t.Fatalf("unexpected drop: %+v", drop)
	}
	var saved orm.RemasterProgress
	if err := orm.GormDB.First(&saved, "commander_id = ? AND chapter_id = ? AND pos = ?", client.Commander.CommanderID, 1001, 1).Error; err != nil {
		t.Fatalf("load remaster progress: %v", err)
	}
	if !saved.Received {
		t.Fatalf("expected reward to be marked received")
	}
	var item orm.CommanderItem
	if err := orm.GormDB.First(&item, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 2001).Error; err != nil {
		t.Fatalf("load reward item: %v", err)
	}
	if item.Count != 1 {
		t.Fatalf("expected item count 1, got %d", item.Count)
	}
}
