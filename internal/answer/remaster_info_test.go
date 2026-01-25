package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestRemasterInfoReturnsProgress(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.RemasterProgress{})
	clearTable(t, &orm.ConfigEntry{})
	seedConfigEntry(t, "ShareCfg/re_map_template.json", "1", `{"id":1,"drop_gain":[[1001,2,2001,3]]}`)
	progress := orm.RemasterProgress{
		CommanderID: client.Commander.CommanderID,
		ChapterID:   1001,
		Pos:         1,
		Count:       2,
		Received:    true,
	}
	if err := orm.GormDB.Create(&progress).Error; err != nil {
		t.Fatalf("seed remaster progress: %v", err)
	}

	payload := protobuf.CS_13505{Type: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemasterInfo(&buffer, client); err != nil {
		t.Fatalf("remaster info failed: %v", err)
	}

	var response protobuf.SC_13506
	decodeResponse(t, client, &response)
	list := response.GetRemapCountList()
	if len(list) != 1 {
		t.Fatalf("expected 1 remap entry, got %d", len(list))
	}
	entry := list[0]
	if entry.GetChapterId() != 1001 || entry.GetPos() != 1 {
		t.Fatalf("unexpected entry: %+v", entry)
	}
	if entry.GetCount() != 2 || entry.GetFlag() != 1 {
		t.Fatalf("unexpected count/flag: %+v", entry)
	}
}
