package answer

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const composeDataTemplateCategory = "ShareCfg/compose_data_template.json"

func setupCompositeEquipmentTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.OwnedEquipment{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.Equipment{})
	clearTable(t, &orm.Resource{})
	clearTable(t, &orm.Item{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: 810, AccountID: 810, Name: "Composite Tester"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	client := &connection.Client{Commander: &commander}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return client
}

func seedCompositeResource(t *testing.T, id uint32) {
	t.Helper()
	resource := orm.Resource{ID: id, Name: fmt.Sprintf("res-%d", id)}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}
}

func seedCompositeItem(t *testing.T, id uint32) {
	t.Helper()
	item := orm.Item{ID: id, Name: fmt.Sprintf("item-%d", id), Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	if err := orm.GormDB.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
}

func seedCompositeEquipment(t *testing.T, id uint32) {
	t.Helper()
	equip := orm.Equipment{
		ID:                id,
		DestroyGold:       0,
		DestroyItem:       json.RawMessage(`[]`),
		EquipLimit:        0,
		Group:             1,
		Important:         0,
		Level:             1,
		Next:              0,
		Prev:              0,
		RestoreGold:       0,
		RestoreItem:       json.RawMessage(`[]`),
		ShipTypeForbidden: json.RawMessage(`[]`),
		TransUseGold:      0,
		TransUseItem:      json.RawMessage(`[]`),
		Type:              1,
		UpgradeFormulaID:  json.RawMessage(`[]`),
	}
	if err := orm.GormDB.Create(&equip).Error; err != nil {
		t.Fatalf("seed equipment: %v", err)
	}
}

func seedCompositeCommanderGold(t *testing.T, commanderID uint32, amount uint32) {
	t.Helper()
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: commanderID, ResourceID: 1, Amount: amount}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
}

func seedCompositeCommanderItem(t *testing.T, commanderID uint32, itemID uint32, count uint32) {
	t.Helper()
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: commanderID, ItemID: itemID, Count: count}).Error; err != nil {
		t.Fatalf("seed commander item: %v", err)
	}
}

func sendCS14006(t *testing.T, client *connection.Client, composeID uint32, num uint32) *protobuf.SC_14007 {
	t.Helper()
	payload := protobuf.CS_14006{Id: proto.Uint32(composeID), Num: proto.Uint32(num)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := CompositeEquipment(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	resp := &protobuf.SC_14007{}
	decodePacket(t, client, 14007, resp)
	return resp
}

func TestCompositeEquipment_SuccessConsumesGoldAndMaterialsAndAddsEquipment(t *testing.T) {
	client := setupCompositeEquipmentTest(t)
	seedCompositeResource(t, 1)
	seedCompositeItem(t, 3001)
	seedCompositeEquipment(t, 2001)
	seedCompositeCommanderGold(t, client.Commander.CommanderID, 200)
	seedCompositeCommanderItem(t, client.Commander.CommanderID, 3001, 10)
	seedConfigEntry(t, composeDataTemplateCategory, "9001", `{"id":9001,"equip_id":2001,"material_id":3001,"material_num":3,"gold_num":10}`)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	resp := sendCS14006(t, client, 9001, 2)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	if client.Commander.GetResourceCount(1) != 180 {
		t.Fatalf("expected gold 180, got %d", client.Commander.GetResourceCount(1))
	}
	if client.Commander.GetItemCount(3001) != 4 {
		t.Fatalf("expected material count 4, got %d", client.Commander.GetItemCount(3001))
	}
	owned := client.Commander.GetOwnedEquipment(2001)
	if owned == nil || owned.Count != 2 {
		t.Fatalf("expected 2 owned equipment")
	}
}

func TestCompositeEquipment_FailsWhenNotEnoughGold(t *testing.T) {
	client := setupCompositeEquipmentTest(t)
	seedCompositeResource(t, 1)
	seedCompositeItem(t, 3001)
	seedCompositeEquipment(t, 2001)
	seedCompositeCommanderGold(t, client.Commander.CommanderID, 5)
	seedCompositeCommanderItem(t, client.Commander.CommanderID, 3001, 10)
	seedConfigEntry(t, composeDataTemplateCategory, "9001", `{"id":9001,"equip_id":2001,"material_id":3001,"material_num":3,"gold_num":10}`)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(3001)

	resp := sendCS14006(t, client, 9001, 1)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	if client.Commander.GetResourceCount(1) != goldBefore {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(3001) != itemBefore {
		t.Fatalf("expected items unchanged")
	}
	if client.Commander.GetOwnedEquipment(2001) != nil {
		t.Fatalf("expected no equipment granted")
	}
}

func TestCompositeEquipment_FailsWhenNotEnoughMaterials(t *testing.T) {
	client := setupCompositeEquipmentTest(t)
	seedCompositeResource(t, 1)
	seedCompositeItem(t, 3001)
	seedCompositeEquipment(t, 2001)
	seedCompositeCommanderGold(t, client.Commander.CommanderID, 200)
	seedCompositeCommanderItem(t, client.Commander.CommanderID, 3001, 2)
	seedConfigEntry(t, composeDataTemplateCategory, "9001", `{"id":9001,"equip_id":2001,"material_id":3001,"material_num":3,"gold_num":10}`)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(3001)

	resp := sendCS14006(t, client, 9001, 1)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	if client.Commander.GetResourceCount(1) != goldBefore {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(3001) != itemBefore {
		t.Fatalf("expected items unchanged")
	}
	if client.Commander.GetOwnedEquipment(2001) != nil {
		t.Fatalf("expected no equipment granted")
	}
}

func TestCompositeEquipment_FailsWhenBagCapacityExceeded(t *testing.T) {
	client := setupCompositeEquipmentTest(t)
	seedCompositeResource(t, 1)
	seedCompositeItem(t, 3001)
	seedCompositeEquipment(t, 2001)
	seedCompositeEquipment(t, 9999)
	seedCompositeCommanderGold(t, client.Commander.CommanderID, 200)
	seedCompositeCommanderItem(t, client.Commander.CommanderID, 3001, 10)
	seedConfigEntry(t, composeDataTemplateCategory, "9001", `{"id":9001,"equip_id":2001,"material_id":3001,"material_num":3,"gold_num":10}`)
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 9999, Count: equipBagMax}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(3001)

	resp := sendCS14006(t, client, 9001, 1)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	if client.Commander.GetResourceCount(1) != goldBefore {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(3001) != itemBefore {
		t.Fatalf("expected items unchanged")
	}
	if client.Commander.GetOwnedEquipment(2001) != nil {
		t.Fatalf("expected no equipment granted")
	}
}

func TestCompositeEquipment_FailsWhenRecipeMissing(t *testing.T) {
	client := setupCompositeEquipmentTest(t)
	seedCompositeResource(t, 1)
	seedCompositeItem(t, 3001)
	seedCompositeEquipment(t, 2001)
	seedCompositeCommanderGold(t, client.Commander.CommanderID, 200)
	seedCompositeCommanderItem(t, client.Commander.CommanderID, 3001, 10)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(3001)

	resp := sendCS14006(t, client, 9999, 1)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	if client.Commander.GetResourceCount(1) != goldBefore {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(3001) != itemBefore {
		t.Fatalf("expected items unchanged")
	}
	if client.Commander.GetOwnedEquipment(2001) != nil {
		t.Fatalf("expected no equipment granted")
	}
}
