package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestGetChapterDropShipListEmpty(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.ChapterDrop{})

	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "101", `{"id":101}`)

	request := protobuf.CS_13109{Id: proto.Uint32(101)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := GetChapterDropShipList(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_13110
	decodeResponse(t, client, &response)
	if len(response.GetDropShipList()) != 0 {
		t.Fatalf("expected empty drop list")
	}
}

func TestGetChapterDropShipListReturnsDrops(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.ChapterDrop{})

	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "101", `{"id":101}`)

	if err := orm.AddChapterDrop(orm.GormDB, &orm.ChapterDrop{CommanderID: client.Commander.CommanderID, ChapterID: 101, ShipID: 2001}); err != nil {
		t.Fatalf("seed drop: %v", err)
	}
	if err := orm.AddChapterDrop(orm.GormDB, &orm.ChapterDrop{CommanderID: client.Commander.CommanderID, ChapterID: 101, ShipID: 2002}); err != nil {
		t.Fatalf("seed drop: %v", err)
	}
	// Duplicate insert should be ignored and not affect response.
	if err := orm.AddChapterDrop(orm.GormDB, &orm.ChapterDrop{CommanderID: client.Commander.CommanderID, ChapterID: 101, ShipID: 2001}); err != nil {
		t.Fatalf("seed duplicate drop: %v", err)
	}

	request := protobuf.CS_13109{Id: proto.Uint32(101)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := GetChapterDropShipList(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_13110
	decodeResponse(t, client, &response)
	ships := response.GetDropShipList()
	if len(ships) != 2 || ships[0] != 2001 || ships[1] != 2002 {
		t.Fatalf("expected [2001 2002], got %v", ships)
	}
}

func TestGetChapterDropShipListInvalidChapter(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.ChapterDrop{})

	request := protobuf.CS_13109{Id: proto.Uint32(999)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := GetChapterDropShipList(&buffer, client); err == nil {
		t.Fatalf("expected error for invalid chapter")
	}
}
