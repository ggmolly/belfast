package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestSupportShipRequisitionSuccess(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	clearTable(t, &orm.RequisitionShip{})
	clearTable(t, &orm.Ship{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntry(t, "ShareCfg/gameset.json", "supports_config", `{"key_value":0,"description":[6,[[2,5400],[3,3200],[4,1000],[5,400]],999]}`)
	seedRequisitionShip(t, 1001, 2)
	seedRequisitionShip(t, 1002, 3)
	seedRequisitionShip(t, 1003, 4)
	seedRequisitionShip(t, 1004, 5)
	seedCommanderItem(t, client, supportRequisitionItemID, 24)

	payload := protobuf.CS_16100{Cnt: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SupportShipRequisition(&buffer, client); err != nil {
		t.Fatalf("support requisition failed: %v", err)
	}

	var response protobuf.SC_16101
	decodeResponse(t, client, &response)
	if response.GetResult() != supportRequisitionResultOK {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetShipList()) != 2 {
		t.Fatalf("expected 2 ships, got %d", len(response.GetShipList()))
	}

	requisitionIDs, err := orm.ListRequisitionShipIDs()
	if err != nil {
		t.Fatalf("list requisition ships: %v", err)
	}
	lookup := make(map[uint32]struct{}, len(requisitionIDs))
	for _, id := range requisitionIDs {
		lookup[id] = struct{}{}
	}
	for _, ship := range response.GetShipList() {
		if _, ok := lookup[ship.GetTemplateId()]; !ok {
			t.Fatalf("unexpected ship template id %d", ship.GetTemplateId())
		}
	}

	var item orm.CommanderItem
	if err := orm.GormDB.First(&item, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, supportRequisitionItemID).Error; err != nil {
		t.Fatalf("load medals: %v", err)
	}
	if item.Count != 12 {
		t.Fatalf("expected medals 12, got %d", item.Count)
	}

	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.SupportRequisitionCount != 2 {
		t.Fatalf("expected support requisition count 2, got %d", commander.SupportRequisitionCount)
	}
	if commander.SupportRequisitionMonth != orm.SupportRequisitionMonth(time.Now()) {
		t.Fatalf("unexpected support requisition month %d", commander.SupportRequisitionMonth)
	}
}

func TestSupportShipRequisitionLimit(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	clearTable(t, &orm.RequisitionShip{})
	clearTable(t, &orm.Ship{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntry(t, "ShareCfg/gameset.json", "supports_config", `{"key_value":0,"description":[6,[[2,100]],1]}`)
	seedRequisitionShip(t, 2001, 2)
	seedCommanderItem(t, client, supportRequisitionItemID, 12)

	client.Commander.SupportRequisitionMonth = orm.SupportRequisitionMonth(time.Now())
	client.Commander.SupportRequisitionCount = 1
	if err := orm.GormDB.Save(client.Commander).Error; err != nil {
		t.Fatalf("save commander: %v", err)
	}

	payload := protobuf.CS_16100{Cnt: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SupportShipRequisition(&buffer, client); err != nil {
		t.Fatalf("support requisition failed: %v", err)
	}

	var response protobuf.SC_16101
	decodeResponse(t, client, &response)
	if response.GetResult() != supportRequisitionResultLimitReached {
		t.Fatalf("expected result 30, got %d", response.GetResult())
	}

	var item orm.CommanderItem
	if err := orm.GormDB.First(&item, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, supportRequisitionItemID).Error; err != nil {
		t.Fatalf("load medals: %v", err)
	}
	if item.Count != 12 {
		t.Fatalf("expected medals 12, got %d", item.Count)
	}
}

func TestSupportShipRequisitionNotEnoughMedals(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	clearTable(t, &orm.RequisitionShip{})
	clearTable(t, &orm.Ship{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntry(t, "ShareCfg/gameset.json", "supports_config", `{"key_value":0,"description":[6,[[2,100]],999]}`)
	seedRequisitionShip(t, 3001, 2)
	seedCommanderItem(t, client, supportRequisitionItemID, 5)

	payload := protobuf.CS_16100{Cnt: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SupportShipRequisition(&buffer, client); err != nil {
		t.Fatalf("support requisition failed: %v", err)
	}

	var response protobuf.SC_16101
	decodeResponse(t, client, &response)
	if response.GetResult() != supportRequisitionResultNotEnoughMedals {
		t.Fatalf("expected result 2, got %d", response.GetResult())
	}

	var count int64
	if err := orm.GormDB.Model(&orm.OwnedShip{}).Where("owner_id = ?", client.Commander.CommanderID).Count(&count).Error; err != nil {
		t.Fatalf("count owned ships: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 ships, got %d", count)
	}
}

func seedRequisitionShip(t *testing.T, shipID uint32, rarity uint32) {
	t.Helper()
	ship := orm.Ship{
		TemplateID:  shipID,
		Name:        "Support Test",
		RarityID:    rarity,
		Star:        1,
		Type:        1,
		Nationality: 1,
		BuildTime:   1,
	}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
	entry := orm.RequisitionShip{ShipID: shipID}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("create requisition ship: %v", err)
	}
}
