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
	"gorm.io/gorm"
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
	if err := orm.GormDB.Create(&commander).Error; err != nil {
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
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
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
	if err := orm.GormDB.Create(&equipConfig).Error; err != nil {
		t.Fatalf("create equipment: %v", err)
	}
	ownedShip, err := client.Commander.AddShip(ship.TemplateID)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2001, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
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
	var entry orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != 2001 {
		t.Fatalf("expected equip id 2001, got %d", entry.EquipID)
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
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != 0 {
		t.Fatalf("expected equip id 0 after unequip, got %d", entry.EquipID)
	}
	if owned := client.Commander.GetOwnedEquipment(2001); owned == nil || owned.Count != 1 {
		t.Fatalf("expected bag count 1 after unequip")
	}
}

func clearEquipTable(t *testing.T, model any) {
	t.Helper()
	if err := orm.GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}

func seedShipEquipConfig(t *testing.T, shipID uint32, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: "sharecfgdata/ship_data_template.json", Key: fmt.Sprintf("%d", shipID), Data: json.RawMessage(payload)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}
