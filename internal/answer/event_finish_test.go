package answer

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func setupEventFinishTest(t *testing.T) *connection.Client {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.EventCollection{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.Ship{})
	return client
}

func seedEventFinishTemplate(t *testing.T, collectionID uint32, payload string) {
	t.Helper()
	seedConfigEntry(t, "ShareCfg/collection_template.json", fmt.Sprintf("%d", collectionID), payload)
}

func seedEventFinishOwnedShip(t *testing.T, commanderID uint32, templateID uint32, level uint32, maxLevel uint32) orm.OwnedShip {
	t.Helper()
	owned := orm.OwnedShip{OwnerID: commanderID, ShipID: templateID, Level: level, MaxLevel: maxLevel}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	return owned
}

func TestEventFinishSuccessAwardsAndPersists(t *testing.T) {
	client := setupEventFinishTest(t)
	oldIntn := eventFinishIntn
	eventFinishIntn = func(n int) int {
		if n == 100 {
			return 50 // no crit
		}
		return 0
	}
	defer func() { eventFinishIntn = oldIntn }()

	seedEventFinishTemplate(t, 101, `{"id":101,"exp":25,"collect_time":1,"ship_num":2,"ship_lv":0,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"drop_display":[[1,1,"20~40"],[2,30001,2]],"special_drop":[[2,40001,1]],"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 1)
	ship1 := seedEventFinishOwnedShip(t, client.Commander.CommanderID, 1001, 100, 100)
	ship2 := seedEventFinishOwnedShip(t, client.Commander.CommanderID, 1002, 100, 100)

	readyTime := uint32(time.Now().Unix()) - 1
	if err := orm.GormDB.Create(&orm.EventCollection{CommanderID: client.Commander.CommanderID, CollectionID: 101, StartTime: readyTime - 10, FinishTime: readyTime, ShipIDs: orm.ToInt64List([]uint32{ship1.ID, ship2.ID})}).Error; err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	request := protobuf.CS_13005{Id: proto.Uint32(101)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buf := data
	if _, _, err := EventFinish(&buf, client); err != nil {
		t.Fatalf("handler: %v", err)
	}

	var response protobuf.SC_13006
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	if response.GetExp() != 25 {
		t.Fatalf("expected exp 25")
	}
	if response.GetIsCri() != 0 {
		t.Fatalf("expected is_cri=0")
	}

	var stored orm.EventCollection
	if err := orm.GormDB.First(&stored, "commander_id = ? AND collection_id = ?", client.Commander.CommanderID, 101).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected event deleted")
	}

	var updatedCommander orm.Commander
	if err := orm.GormDB.First(&updatedCommander, "commander_id = ?", client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if updatedCommander.CollectAttackCount != 1 {
		t.Fatalf("expected collect_attack_count=1, got %d", updatedCommander.CollectAttackCount)
	}

	var updatedShip orm.OwnedShip
	if err := orm.GormDB.First(&updatedShip, "id = ?", ship1.ID).Error; err != nil {
		t.Fatalf("load ship: %v", err)
	}
	if updatedShip.SurplusExp != 25 {
		t.Fatalf("expected surplus exp 25, got %d", updatedShip.SurplusExp)
	}

	var gold orm.OwnedResource
	if err := orm.GormDB.First(&gold, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).Error; err != nil {
		t.Fatalf("load gold: %v", err)
	}
	if gold.Amount != 20 {
		t.Fatalf("expected gold 20, got %d", gold.Amount)
	}

	var item orm.CommanderItem
	if err := orm.GormDB.First(&item, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 30001).Error; err != nil {
		t.Fatalf("load item: %v", err)
	}
	if item.Count != 2 {
		t.Fatalf("expected item count 2, got %d", item.Count)
	}
}

func TestEventFinishInvalidEventIDReturnsCode1(t *testing.T) {
	client := setupEventFinishTest(t)
	request := protobuf.CS_13005{Id: proto.Uint32(999)}
	data, _ := proto.Marshal(&request)
	buf := data
	if _, _, err := EventFinish(&buf, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13006
	decodeResponse(t, client, &response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1")
	}
}

func TestEventFinishNoActiveEventReturnsCode2(t *testing.T) {
	client := setupEventFinishTest(t)
	seedEventFinishTemplate(t, 101, `{"id":101,"exp":1,"collect_time":1,"ship_num":1,"ship_lv":0,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"drop_display":[],"special_drop":[],"type":1,"max_team":0}`)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13005{Id: proto.Uint32(101)}
	data, _ := proto.Marshal(&request)
	buf := data
	if _, _, err := EventFinish(&buf, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13006
	decodeResponse(t, client, &response)
	if response.GetResult() != 2 {
		t.Fatalf("expected result 2")
	}
}

func TestEventFinishExpiredTemplateReturnsCode3(t *testing.T) {
	client := setupEventFinishTest(t)
	seedEventFinishTemplate(t, 101, `{"id":101,"exp":1,"collect_time":1,"ship_num":1,"ship_lv":0,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":1,"drop_display":[],"special_drop":[],"type":1,"max_team":0}`)
	request := protobuf.CS_13005{Id: proto.Uint32(101)}
	data, _ := proto.Marshal(&request)
	buf := data
	if _, _, err := EventFinish(&buf, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13006
	decodeResponse(t, client, &response)
	if response.GetResult() != 3 {
		t.Fatalf("expected result 3")
	}
}

func TestEventFinishCriticalHitAddsSpecialDrops(t *testing.T) {
	client := setupEventFinishTest(t)
	oldIntn := eventFinishIntn
	eventFinishIntn = func(n int) int { return 0 }
	defer func() { eventFinishIntn = oldIntn }()

	seedEventFinishTemplate(t, 101, `{"id":101,"exp":0,"collect_time":1,"ship_num":1,"ship_lv":0,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"drop_display":[],"special_drop":[[2,40001,1]],"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	ship1 := seedEventFinishOwnedShip(t, client.Commander.CommanderID, 1001, 1, 100)
	readyTime := uint32(time.Now().Unix()) - 1
	if err := orm.GormDB.Create(&orm.EventCollection{CommanderID: client.Commander.CommanderID, CollectionID: 101, StartTime: readyTime - 10, FinishTime: readyTime, ShipIDs: orm.ToInt64List([]uint32{ship1.ID})}).Error; err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	request := protobuf.CS_13005{Id: proto.Uint32(101)}
	data, _ := proto.Marshal(&request)
	buf := data
	if _, _, err := EventFinish(&buf, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13006
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success")
	}
	if response.GetIsCri() != 1 {
		t.Fatalf("expected is_cri=1")
	}
	found := false
	for _, drop := range response.GetDropList() {
		if drop.GetType() == 2 && drop.GetId() == 40001 && drop.GetNumber() == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected special drop to be included")
	}
}
