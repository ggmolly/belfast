package answer_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupEquipTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearEquipTable(t, &orm.OwnedShipEquipment{})
	clearEquipTable(t, &orm.OwnedShip{})
	clearEquipTable(t, &orm.OwnedEquipment{})
	clearEquipTable(t, &orm.Equipment{})
	clearEquipTable(t, &orm.ConfigEntry{})
	clearEquipTable(t, &orm.Ship{})
	clearEquipTable(t, &orm.Commander{})
	commander := orm.Commander{CommanderID: 101, AccountID: 101, Name: "Equip Tester"}
	if err := orm.CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestEquipToShipEquipAndUnequip(t *testing.T) {
	client := setupEquipTest(t)
	ship := orm.Ship{
		TemplateID:  1001,
		Name:        "Ship",
		EnglishName: "Ship",
		RarityID:    2,
		Star:        1,
		Type:        1,
		Nationality: 1,
		BuildTime:   10,
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[2],"equip_3":[3],"equip_4":[],"equip_5":[],"equip_id_1":0,"equip_id_2":0,"equip_id_3":0}`)
	equipConfig := orm.Equipment{
		ID:                2001,
		DestroyGold:       10,
		EquipLimit:        0,
		Group:             1,
		Important:         1,
		Level:             1,
		RestoreGold:       0,
		TransUseGold:      0,
		Type:              1,
		ShipTypeForbidden: json.RawMessage(`[]`),
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, destroy_gold, equip_limit, important, level, restore_gold, trans_use_gold, type, ship_type_forbidden) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)", int64(equipConfig.ID), int64(equipConfig.DestroyGold), int64(equipConfig.EquipLimit), int64(equipConfig.Important), int64(equipConfig.Level), int64(equipConfig.RestoreGold), int64(equipConfig.TransUseGold), int64(equipConfig.Type), string(equipConfig.ShipTypeForbidden))
	ownedShip, err := client.Commander.AddShip(ship.TemplateID)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2001), int64(1))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	payload := protobuf.CS_12006{
		ShipId:  proto.Uint32(ownedShip.ID),
		EquipId: proto.Uint32(2001),
		Pos:     proto.Uint32(1),
		Type:    proto.Uint32(0),
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.EquipToShip(&buf, client); err != nil {
		t.Fatalf("EquipToShip failed: %v", err)
	}
	response := &protobuf.SC_12007{}
	packetId := decodePacket(t, client, 12007, response)
	if packetId != 12007 || response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}
	equippedID := queryAnswerExternalTestInt64(t, "SELECT equip_id FROM owned_ship_equipments WHERE owner_id = $1 AND ship_id = $2 AND pos = $3", int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1))
	if equippedID != 2001 {
		t.Fatalf("expected equip id 2001, got %d", equippedID)
	}
	if client.Commander.GetOwnedEquipment(2001) != nil {
		t.Fatalf("expected equipment to be removed from bag")
	}
	payload = protobuf.CS_12006{
		ShipId:  proto.Uint32(ownedShip.ID),
		EquipId: proto.Uint32(0),
		Pos:     proto.Uint32(1),
		Type:    proto.Uint32(0),
	}
	buf, err = proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.EquipToShip(&buf, client); err != nil {
		t.Fatalf("EquipToShip unequip failed: %v", err)
	}
	response = &protobuf.SC_12007{}
	decodePacket(t, client, 12007, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result on unequip, got %d", response.GetResult())
	}
	unequippedID := queryAnswerExternalTestInt64(t, "SELECT equip_id FROM owned_ship_equipments WHERE owner_id = $1 AND ship_id = $2 AND pos = $3", int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1))
	if unequippedID != 0 {
		t.Fatalf("expected equip id 0 after unequip, got %d", unequippedID)
	}
	if owned := client.Commander.GetOwnedEquipment(2001); owned == nil || owned.Count != 1 {
		t.Fatalf("expected bag count 1 after unequip")
	}
}

func clearEquipTable(t *testing.T, model any) {
	t.Helper()
	var table string
	switch model.(type) {
	case *orm.OwnedShipEquipment:
		table = "owned_ship_equipments"
	case *orm.OwnedShip:
		table = "owned_ships"
	case *orm.OwnedEquipment:
		table = "owned_equipments"
	case *orm.Equipment:
		table = "equipments"
	case *orm.ConfigEntry:
		table = "config_entries"
	case *orm.Ship:
		table = "ships"
	case *orm.Commander:
		table = "commanders"
	default:
		t.Fatalf("unsupported model type for clearEquipTable: %T", model)
	}
	execAnswerExternalTestSQLT(t, fmt.Sprintf("DELETE FROM %s", table))
}

func seedShipEquipConfig(t *testing.T, shipID uint32, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: "sharecfgdata/ship_data_template.json", Key: fmt.Sprintf("%d", shipID), Data: json.RawMessage(payload)}
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", entry.Category, entry.Key, string(entry.Data))
}
