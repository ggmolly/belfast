package answer

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupModShipTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.OwnedShipStrength{})
	clearTable(t, &orm.OwnedShipEquipment{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.OwnedEquipment{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})
	commander := orm.Commander{CommanderID: 601, AccountID: 601, Name: "Mod Ship Tester"}
	if err := orm.CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func decodePacket(t *testing.T, client *connection.Client, expectedID int, message proto.Message) {
	t.Helper()
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
}

func seedModShipTemplate(t *testing.T, templateID uint32, strengthenID uint32, groupType uint32) {
	t.Helper()
	payload := fmt.Sprintf(`{"id":%d,"strengthen_id":%d,"group_type":%d}`, templateID, strengthenID, groupType)
	seedConfigEntry(t, "sharecfgdata/ship_data_template.json", fmt.Sprintf("%d", templateID), payload)
}

func seedModStrengthenConfig(t *testing.T, strengthenID uint32, attrExp []uint32, durability []uint32, levelExp []uint32) {
	t.Helper()
	payload, err := json.Marshal(struct {
		ID         uint32   `json:"id"`
		AttrExp    []uint32 `json:"attr_exp"`
		Durability []uint32 `json:"durability"`
		LevelExp   []uint32 `json:"level_exp"`
	}{
		ID:         strengthenID,
		AttrExp:    attrExp,
		Durability: durability,
		LevelExp:   levelExp,
	})
	if err != nil {
		t.Fatalf("marshal strengthen payload: %v", err)
	}
	seedConfigEntry(t, "ShareCfg/ship_data_strengthen.json", fmt.Sprintf("%d", strengthenID), string(payload))
}

func TestModShipSuccess(t *testing.T) {
	client := setupModShipTest(t)
	seedModShipTemplate(t, 1001, 4001, 10)
	seedModShipTemplate(t, 2001, 4002, 10)
	seedModShipTemplate(t, 2002, 4003, 20)
	seedModStrengthenConfig(t, 4001, []uint32{0, 0, 0, 0, 0}, []uint32{10, 10, 10, 10, 10}, []uint32{2, 2, 2, 2, 2})
	seedModStrengthenConfig(t, 4002, []uint32{1, 2, 3, 4, 5}, []uint32{10, 10, 10, 10, 10}, []uint32{2, 2, 2, 2, 2})
	seedModStrengthenConfig(t, 4003, []uint32{2, 2, 2, 2, 2}, []uint32{10, 10, 10, 10, 10}, []uint32{2, 2, 2, 2, 2})
	mainShip := orm.OwnedShip{ID: 61001, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 50}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, NOW(), NOW())", int64(mainShip.ID), int64(mainShip.OwnerID), int64(mainShip.ShipID), int64(mainShip.Level))
	materialShip := orm.OwnedShip{ID: 61002, OwnerID: client.Commander.CommanderID, ShipID: 2001, Level: 1}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, NOW(), NOW())", int64(materialShip.ID), int64(materialShip.OwnerID), int64(materialShip.ShipID), int64(materialShip.Level))
	materialShip2 := orm.OwnedShip{ID: 61003, OwnerID: client.Commander.CommanderID, ShipID: 2002, Level: 1}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, NOW(), NOW())", int64(materialShip2.ID), int64(materialShip2.OwnerID), int64(materialShip2.ShipID), int64(materialShip2.Level))
	equipEntry := orm.OwnedShipEquipment{OwnerID: client.Commander.CommanderID, ShipID: materialShip.ID, Pos: 1, EquipID: 3001, SkinID: 0}
	execAnswerTestSQLT(t, "INSERT INTO owned_ship_equipments (owner_id, ship_id, pos, equip_id, skin_id) VALUES ($1, $2, $3, $4, $5)", int64(equipEntry.OwnerID), int64(equipEntry.ShipID), int64(equipEntry.Pos), int64(equipEntry.EquipID), int64(equipEntry.SkinID))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12017{
		ShipId:         proto.Uint32(mainShip.ID),
		MaterialIdList: []uint32{materialShip.ID, materialShip2.ID},
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ModShip(&buf, client); err != nil {
		t.Fatalf("ModShip failed: %v", err)
	}
	response := &protobuf.SC_12018{}
	decodePacket(t, client, 12018, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}

	strengths, err := orm.ListOwnedShipStrengths(client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load strengths: %v", err)
	}
	if len(strengths) != 5 {
		t.Fatalf("expected 5 strength entries, got %d", len(strengths))
	}
	strengthMap := make(map[uint32]uint32)
	for _, entry := range strengths {
		strengthMap[entry.StrengthID] = entry.Exp
	}
	expected := map[uint32]uint32{2: 4, 3: 6, 4: 8, 5: 10, 6: 12}
	for strengthID, exp := range expected {
		if strengthMap[strengthID] != exp {
			t.Fatalf("expected strength %d exp %d, got %d", strengthID, exp, strengthMap[strengthID])
		}
	}
	materialCount := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM owned_ships WHERE owner_id = $1 AND id = $2", int64(client.Commander.CommanderID), int64(materialShip.ID))
	if materialCount != 0 {
		t.Fatalf("expected material ship to be deleted")
	}
	materialCount2 := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM owned_ships WHERE owner_id = $1 AND id = $2", int64(client.Commander.CommanderID), int64(materialShip2.ID))
	if materialCount2 != 0 {
		t.Fatalf("expected material ship 2 to be deleted")
	}
	bagCount := queryAnswerTestInt64(t, "SELECT count FROM owned_equipments WHERE commander_id = $1 AND equipment_id = $2", int64(client.Commander.CommanderID), int64(3001))
	if bagCount != 1 {
		t.Fatalf("expected equipment count 1, got %d", bagCount)
	}
}

func TestModShipEmptyMaterials(t *testing.T) {
	client := setupModShipTest(t)
	seedModShipTemplate(t, 1001, 4001, 10)
	seedModStrengthenConfig(t, 4001, []uint32{0, 0, 0, 0, 0}, []uint32{10, 10, 10, 10, 10}, []uint32{2, 2, 2, 2, 2})
	mainShip := orm.OwnedShip{ID: 62001, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 50}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, NOW(), NOW())", int64(mainShip.ID), int64(mainShip.OwnerID), int64(mainShip.ShipID), int64(mainShip.Level))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12017{ShipId: proto.Uint32(mainShip.ID)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ModShip(&buf, client); err != nil {
		t.Fatalf("ModShip failed: %v", err)
	}
	response := &protobuf.SC_12018{}
	decodePacket(t, client, 12018, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected failure result, got %d", response.GetResult())
	}
}

func TestModShipMissingMaterial(t *testing.T) {
	client := setupModShipTest(t)
	seedModShipTemplate(t, 1001, 4001, 10)
	seedModStrengthenConfig(t, 4001, []uint32{0, 0, 0, 0, 0}, []uint32{10, 10, 10, 10, 10}, []uint32{2, 2, 2, 2, 2})
	mainShip := orm.OwnedShip{ID: 63001, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 50}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, NOW(), NOW())", int64(mainShip.ID), int64(mainShip.OwnerID), int64(mainShip.ShipID), int64(mainShip.Level))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12017{
		ShipId:         proto.Uint32(mainShip.ID),
		MaterialIdList: []uint32{9999},
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ModShip(&buf, client); err != nil {
		t.Fatalf("ModShip failed: %v", err)
	}
	response := &protobuf.SC_12018{}
	decodePacket(t, client, 12018, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected failure result, got %d", response.GetResult())
	}
}

func TestToProtoOwnedShipIncludesStrengthList(t *testing.T) {
	ship := orm.OwnedShip{
		ID:     7001,
		ShipID: 1001,
		Strengths: []orm.OwnedShipStrength{
			{StrengthID: 2, Exp: 9},
			{StrengthID: 5, Exp: 12},
		},
	}
	info := orm.ToProtoOwnedShip(ship, nil, nil)
	if len(info.GetStrengthList()) != 2 {
		t.Fatalf("expected 2 strength entries, got %d", len(info.GetStrengthList()))
	}
	if info.GetStrengthList()[0].GetId() != 2 || info.GetStrengthList()[0].GetExp() != 9 {
		t.Fatalf("unexpected strength list entry")
	}
}
