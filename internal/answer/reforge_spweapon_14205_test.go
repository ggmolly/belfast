package answer_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupReforgeSpWeaponClient(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.OwnedSpWeapon{})
	clearTable(t, &orm.Commander{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.CommanderMiscItem{})

	commander := orm.Commander{CommanderID: 1, AccountID: 1, Name: "SpWeapon Commander"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedConfigEntryTest(t *testing.T, category string, key string, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: category, Key: key, Data: json.RawMessage(payload)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}

func TestReforgeSpWeaponRollSuccessConsumesMaterialsAndPersistsTemps(t *testing.T) {
	client := setupReforgeSpWeaponClient(t)

	seedConfigEntryTest(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"upgrade_id":2001,"value_1_random":50,"value_2_random":60}`)
	seedConfigEntryTest(t, "ShareCfg/spweapon_upgrade.json", "2001", `{"id":2001,"reset_use_item":[[5001,2],[5002,3]]}`)

	items := []orm.CommanderItem{
		{CommanderID: client.Commander.CommanderID, ItemID: 5001, Count: 5},
		{CommanderID: client.Commander.CommanderID, ItemID: 5002, Count: 4},
	}
	for i := range items {
		if err := orm.GormDB.Create(&items[i]).Error; err != nil {
			t.Fatalf("failed to seed items: %v", err)
		}
	}

	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1001}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14205{ShipId: proto.Uint32(0), SpweaponId: proto.Uint32(spweapon.ID)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ReforgeSpWeapon(&buf, client); err != nil {
		t.Fatalf("ReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14206{}
	decodeTestPacket(t, client, 14206, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetAttrTemp_1() > 50 {
		t.Fatalf("expected attr_temp_1 <= 50, got %d", response.GetAttrTemp_1())
	}
	if response.GetAttrTemp_2() > 60 {
		t.Fatalf("expected attr_temp_2 <= 60, got %d", response.GetAttrTemp_2())
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.First(&stored, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.AttrTemp1 != response.GetAttrTemp_1() || stored.AttrTemp2 != response.GetAttrTemp_2() {
		t.Fatalf("expected persisted temp attrs to match response")
	}
	if client.Commander.GetItemCount(5001) != 3 {
		t.Fatalf("expected item 5001 to be consumed to 3, got %d", client.Commander.GetItemCount(5001))
	}
	if client.Commander.GetItemCount(5002) != 1 {
		t.Fatalf("expected item 5002 to be consumed to 1, got %d", client.Commander.GetItemCount(5002))
	}
}

func TestReforgeSpWeaponRollRejectsPendingTempAttrs(t *testing.T) {
	client := setupReforgeSpWeaponClient(t)

	seedConfigEntryTest(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"upgrade_id":2001,"value_1_random":50,"value_2_random":60}`)
	seedConfigEntryTest(t, "ShareCfg/spweapon_upgrade.json", "2001", `{"id":2001,"reset_use_item":[[5001,2]]}`)
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 5001, Count: 5}).Error; err != nil {
		t.Fatalf("failed to seed items: %v", err)
	}
	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1001, AttrTemp1: 1, AttrTemp2: 0}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14205{ShipId: proto.Uint32(0), SpweaponId: proto.Uint32(spweapon.ID)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ReforgeSpWeapon(&buf, client); err != nil {
		t.Fatalf("ReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14206{}
	decodeTestPacket(t, client, 14206, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.First(&stored, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.AttrTemp1 != 1 || stored.AttrTemp2 != 0 {
		t.Fatalf("expected spweapon temp attrs to remain unchanged")
	}
	if client.Commander.GetItemCount(5001) != 5 {
		t.Fatalf("expected materials not to be consumed")
	}
}

func TestReforgeSpWeaponRollRejectsUnknownUID(t *testing.T) {
	client := setupReforgeSpWeaponClient(t)

	payload := &protobuf.CS_14205{ShipId: proto.Uint32(0), SpweaponId: proto.Uint32(9999)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ReforgeSpWeapon(&buf, client); err != nil {
		t.Fatalf("ReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14206{}
	decodeTestPacket(t, client, 14206, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestReforgeSpWeaponRollRejectsInsufficientMaterials(t *testing.T) {
	client := setupReforgeSpWeaponClient(t)

	seedConfigEntryTest(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"upgrade_id":2001,"value_1_random":50,"value_2_random":60}`)
	seedConfigEntryTest(t, "ShareCfg/spweapon_upgrade.json", "2001", `{"id":2001,"reset_use_item":[[5001,2]]}`)
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 5001, Count: 1}).Error; err != nil {
		t.Fatalf("failed to seed items: %v", err)
	}
	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1001}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14205{ShipId: proto.Uint32(0), SpweaponId: proto.Uint32(spweapon.ID)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ReforgeSpWeapon(&buf, client); err != nil {
		t.Fatalf("ReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14206{}
	decodeTestPacket(t, client, 14206, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.First(&stored, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.AttrTemp1 != 0 || stored.AttrTemp2 != 0 {
		t.Fatalf("expected spweapon temp attrs to remain unchanged")
	}
	if client.Commander.GetItemCount(5001) != 1 {
		t.Fatalf("expected materials not to be consumed")
	}
}
