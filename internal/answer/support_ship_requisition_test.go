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

	item := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(supportRequisitionItemID))
	if item != 12 {
		t.Fatalf("expected medals 12, got %d", item)
	}

	commander, err := orm.GetCommanderCoreByID(client.Commander.CommanderID)
	if err != nil {
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

	execAnswerTestSQLT(t, "UPDATE commanders SET support_requisition_month = $1, support_requisition_count = $2 WHERE commander_id = $3", int64(orm.SupportRequisitionMonth(time.Now())), int64(1), int64(client.Commander.CommanderID))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
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

	item := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(supportRequisitionItemID))
	if item != 12 {
		t.Fatalf("expected medals 12, got %d", item)
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

	count := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM owned_ships WHERE owner_id = $1", int64(client.Commander.CommanderID))
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
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.Name, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	execAnswerTestSQLT(t, "INSERT INTO requisition_ships (ship_id) VALUES ($1)", int64(shipID))
}
