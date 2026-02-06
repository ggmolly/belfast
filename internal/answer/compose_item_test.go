package answer

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var composeCommanderID uint32 = 12000

func setupComposeItemTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.CommanderMiscItem{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})

	commanderID := atomic.AddUint32(&composeCommanderID, 1)
	commander := orm.Commander{CommanderID: commanderID, AccountID: 1, Name: fmt.Sprintf("Compose Tester %d", commanderID)}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	client := &connection.Client{Commander: &commander}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return client
}

func TestComposeItemSuccess(t *testing.T) {
	client := setupComposeItemTest(t)
	seedConfigEntry(t, itemDataStatisticsCategory, "1000", `{"id":1000,"compose_number":3,"target_id":2000}`)
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 1000, Count: 10}).Error; err != nil {
		t.Fatalf("seed source item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	request := protobuf.CS_15006{Id: proto.Uint32(1000), Num: proto.Uint32(2)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := ComposeItem(&buffer, client); err != nil {
		t.Fatalf("compose item failed: %v", err)
	}

	var response protobuf.SC_15007
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var source orm.CommanderItem
	if err := orm.GormDB.First(&source, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 1000).Error; err != nil {
		t.Fatalf("load source item: %v", err)
	}
	if source.Count != 4 {
		t.Fatalf("expected source count 4, got %d", source.Count)
	}
	var target orm.CommanderItem
	if err := orm.GormDB.First(&target, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 2000).Error; err != nil {
		t.Fatalf("load target item: %v", err)
	}
	if target.Count != 2 {
		t.Fatalf("expected target count 2, got %d", target.Count)
	}
}

func TestComposeItemInsufficientItems(t *testing.T) {
	client := setupComposeItemTest(t)
	seedConfigEntry(t, itemDataStatisticsCategory, "1000", `{"id":1000,"compose_number":3,"target_id":2000}`)
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 1000, Count: 5}).Error; err != nil {
		t.Fatalf("seed source item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	request := protobuf.CS_15006{Id: proto.Uint32(1000), Num: proto.Uint32(2)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := ComposeItem(&buffer, client); err != nil {
		t.Fatalf("compose item failed: %v", err)
	}

	var response protobuf.SC_15007
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var source orm.CommanderItem
	if err := orm.GormDB.First(&source, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 1000).Error; err != nil {
		t.Fatalf("load source item: %v", err)
	}
	if source.Count != 5 {
		t.Fatalf("expected source count to remain 5, got %d", source.Count)
	}
	var target orm.CommanderItem
	if err := orm.GormDB.First(&target, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 2000).Error; err == nil {
		t.Fatalf("expected no target item to be granted")
	}
}

func TestComposeItemUsesCombinedItemSources(t *testing.T) {
	client := setupComposeItemTest(t)
	seedConfigEntry(t, itemDataStatisticsCategory, "1000", `{"id":1000,"compose_number":3,"target_id":2000}`)
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 1000, Count: 2}).Error; err != nil {
		t.Fatalf("seed source item: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderMiscItem{CommanderID: client.Commander.CommanderID, ItemID: 1000, Data: 10}).Error; err != nil {
		t.Fatalf("seed misc item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	request := protobuf.CS_15006{Id: proto.Uint32(1000), Num: proto.Uint32(4)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := ComposeItem(&buffer, client); err != nil {
		t.Fatalf("compose item failed: %v", err)
	}

	var response protobuf.SC_15007
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var source orm.CommanderItem
	if err := orm.GormDB.First(&source, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 1000).Error; err != nil {
		t.Fatalf("load source item: %v", err)
	}
	if source.Count != 0 {
		t.Fatalf("expected source count 0, got %d", source.Count)
	}
	var misc orm.CommanderMiscItem
	if err := orm.GormDB.First(&misc, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 1000).Error; err != nil {
		t.Fatalf("load misc item: %v", err)
	}
	if misc.Data != 0 {
		t.Fatalf("expected misc count 0, got %d", misc.Data)
	}
	var target orm.CommanderItem
	if err := orm.GormDB.First(&target, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 2000).Error; err != nil {
		t.Fatalf("load target item: %v", err)
	}
	if target.Count != 4 {
		t.Fatalf("expected target count 4, got %d", target.Count)
	}
}

func TestComposeItemInvalidConfig(t *testing.T) {
	client := setupComposeItemTest(t)
	seedConfigEntry(t, itemDataStatisticsCategory, "1000", `{"id":1000,"compose_number":0,"target_id":2000}`)
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 1000, Count: 10}).Error; err != nil {
		t.Fatalf("seed source item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	request := protobuf.CS_15006{Id: proto.Uint32(1000), Num: proto.Uint32(2)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := ComposeItem(&buffer, client); err != nil {
		t.Fatalf("compose item failed: %v", err)
	}

	var response protobuf.SC_15007
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var source orm.CommanderItem
	if err := orm.GormDB.First(&source, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 1000).Error; err != nil {
		t.Fatalf("load source item: %v", err)
	}
	if source.Count != 10 {
		t.Fatalf("expected source count to remain 10, got %d", source.Count)
	}
}
