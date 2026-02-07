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

func seedConfigEntry(t *testing.T, category string, key string, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: category, Key: key, Data: json.RawMessage(payload)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}

func setupUpgradeSpWeaponClient(t *testing.T) *connection.Client {
	t.Helper()

	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.CommanderMiscItem{})
	clearTable(t, &orm.OwnedSpWeapon{})
	clearTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: 1, AccountID: 1, Name: "UpgradeSpWeapon Commander"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestUpgradeSpWeaponSuccessAppliesCostsAndUpgrades(t *testing.T) {
	client := setupUpgradeSpWeaponClient(t)

	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1000", `{"id":1000,"next":1001,"upgrade_pt":10,"upgrade_use_gold":7}`)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"next":0,"upgrade_pt":0}`)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "2000", `{"id":2000,"upgrade_get_pt":2}`)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "500", `{"id":500,"pt":1}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 20}).Error; err != nil {
		t.Fatalf("failed to create gold row: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 500, Count: 5}).Error; err != nil {
		t.Fatalf("failed to create item row: %v", err)
	}

	target := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1000, Pt: 7}
	if err := orm.GormDB.Create(&target).Error; err != nil {
		t.Fatalf("failed to create target spweapon: %v", err)
	}
	fodder := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 2000, Pt: 1}
	if err := orm.GormDB.Create(&fodder).Error; err != nil {
		t.Fatalf("failed to create fodder spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14203{
		ShipId:         proto.Uint32(0),
		SpweaponId:     proto.Uint32(target.ID),
		ItemIdList:     []uint32{500, 500},
		SpweaponIdList: []uint32{fodder.ID},
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeSpWeapon(&buf, client); err != nil {
		t.Fatalf("UpgradeSpWeapon failed: %v", err)
	}

	resp := &protobuf.SC_14204{}
	decodeTestPacket(t, client, 14204, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}

	var storedTarget orm.OwnedSpWeapon
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, target.ID).First(&storedTarget).Error; err != nil {
		t.Fatalf("failed to load target spweapon: %v", err)
	}
	if storedTarget.TemplateID != 1001 {
		t.Fatalf("expected upgraded template_id 1001, got %d", storedTarget.TemplateID)
	}
	if storedTarget.Pt != 2 {
		t.Fatalf("expected upgraded pt remainder 2, got %d", storedTarget.Pt)
	}

	var storedFodder orm.OwnedSpWeapon
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, fodder.ID).First(&storedFodder).Error; err == nil {
		t.Fatalf("expected fodder spweapon to be deleted")
	}

	var gold orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).First(&gold).Error; err != nil {
		t.Fatalf("failed to load gold: %v", err)
	}
	if gold.Amount != 13 {
		t.Fatalf("expected gold 13, got %d", gold.Amount)
	}

	var item orm.CommanderItem
	if err := orm.GormDB.Where("commander_id = ? AND item_id = ?", client.Commander.CommanderID, 500).First(&item).Error; err != nil {
		t.Fatalf("failed to load item: %v", err)
	}
	if item.Count != 3 {
		t.Fatalf("expected item count 3, got %d", item.Count)
	}
}

func TestUpgradeSpWeaponAccumulatesPtWhenNoUpgradeStep(t *testing.T) {
	client := setupUpgradeSpWeaponClient(t)

	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1000", `{"id":1000,"next":1001,"upgrade_pt":10,"upgrade_use_gold":7}`)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"next":0,"upgrade_pt":0}`)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "500", `{"id":500,"pt":1}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 20}).Error; err != nil {
		t.Fatalf("failed to create gold row: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 500, Count: 5}).Error; err != nil {
		t.Fatalf("failed to create item row: %v", err)
	}

	target := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1000, Pt: 7}
	if err := orm.GormDB.Create(&target).Error; err != nil {
		t.Fatalf("failed to create target spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14203{
		ShipId:     proto.Uint32(0),
		SpweaponId: proto.Uint32(target.ID),
		ItemIdList: []uint32{500, 500},
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeSpWeapon(&buf, client); err != nil {
		t.Fatalf("UpgradeSpWeapon failed: %v", err)
	}

	resp := &protobuf.SC_14204{}
	decodeTestPacket(t, client, 14204, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}

	var storedTarget orm.OwnedSpWeapon
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, target.ID).First(&storedTarget).Error; err != nil {
		t.Fatalf("failed to load target spweapon: %v", err)
	}
	if storedTarget.TemplateID != 1000 {
		t.Fatalf("expected template_id to remain 1000, got %d", storedTarget.TemplateID)
	}
	if storedTarget.Pt != 9 {
		t.Fatalf("expected pt to accumulate to 9, got %d", storedTarget.Pt)
	}

	var gold orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).First(&gold).Error; err != nil {
		t.Fatalf("failed to load gold: %v", err)
	}
	if gold.Amount != 20 {
		t.Fatalf("expected gold 20, got %d", gold.Amount)
	}

	var item orm.CommanderItem
	if err := orm.GormDB.Where("commander_id = ? AND item_id = ?", client.Commander.CommanderID, 500).First(&item).Error; err != nil {
		t.Fatalf("failed to load item: %v", err)
	}
	if item.Count != 3 {
		t.Fatalf("expected item count 3, got %d", item.Count)
	}
}

func TestUpgradeSpWeaponUnknownTargetDoesNotMutate(t *testing.T) {
	client := setupUpgradeSpWeaponClient(t)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 20}).Error; err != nil {
		t.Fatalf("failed to create gold row: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 500, Count: 5}).Error; err != nil {
		t.Fatalf("failed to create item row: %v", err)
	}
	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1000, Pt: 7}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14203{
		ShipId:     proto.Uint32(0),
		SpweaponId: proto.Uint32(999999),
		ItemIdList: []uint32{500, 500},
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeSpWeapon(&buf, client); err != nil {
		t.Fatalf("UpgradeSpWeapon failed: %v", err)
	}

	resp := &protobuf.SC_14204{}
	decodeTestPacket(t, client, 14204, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).First(&stored).Error; err != nil {
		t.Fatalf("expected spweapon to remain: %v", err)
	}
	if stored.TemplateID != 1000 || stored.Pt != 7 {
		t.Fatalf("expected target spweapon to remain unchanged")
	}

	var gold orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).First(&gold).Error; err != nil {
		t.Fatalf("failed to load gold: %v", err)
	}
	if gold.Amount != 20 {
		t.Fatalf("expected gold 20, got %d", gold.Amount)
	}

	var item orm.CommanderItem
	if err := orm.GormDB.Where("commander_id = ? AND item_id = ?", client.Commander.CommanderID, 500).First(&item).Error; err != nil {
		t.Fatalf("failed to load item: %v", err)
	}
	if item.Count != 5 {
		t.Fatalf("expected item count 5, got %d", item.Count)
	}
}

func TestUpgradeSpWeaponInsufficientItemsDoesNotMutate(t *testing.T) {
	client := setupUpgradeSpWeaponClient(t)

	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1000", `{"id":1000,"next":1001,"upgrade_pt":10,"upgrade_use_gold":7}`)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"next":0,"upgrade_pt":0}`)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "500", `{"id":500,"pt":1}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 20}).Error; err != nil {
		t.Fatalf("failed to create gold row: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 500, Count: 1}).Error; err != nil {
		t.Fatalf("failed to create item row: %v", err)
	}
	target := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1000, Pt: 9}
	if err := orm.GormDB.Create(&target).Error; err != nil {
		t.Fatalf("failed to create target spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14203{
		ShipId:     proto.Uint32(0),
		SpweaponId: proto.Uint32(target.ID),
		ItemIdList: []uint32{500, 500},
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeSpWeapon(&buf, client); err != nil {
		t.Fatalf("UpgradeSpWeapon failed: %v", err)
	}

	resp := &protobuf.SC_14204{}
	decodeTestPacket(t, client, 14204, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, target.ID).First(&stored).Error; err != nil {
		t.Fatalf("expected spweapon to remain: %v", err)
	}
	if stored.TemplateID != 1000 || stored.Pt != 9 {
		t.Fatalf("expected spweapon to remain unchanged")
	}
}

func TestUpgradeSpWeaponUnknownConsumedSpweaponDoesNotMutate(t *testing.T) {
	client := setupUpgradeSpWeaponClient(t)

	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1000", `{"id":1000,"next":1001,"upgrade_pt":10,"upgrade_use_gold":7}`)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"next":0,"upgrade_pt":0}`)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "500", `{"id":500,"pt":1}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 20}).Error; err != nil {
		t.Fatalf("failed to create gold row: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 500, Count: 5}).Error; err != nil {
		t.Fatalf("failed to create item row: %v", err)
	}
	target := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1000, Pt: 9}
	if err := orm.GormDB.Create(&target).Error; err != nil {
		t.Fatalf("failed to create target spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14203{
		ShipId:         proto.Uint32(0),
		SpweaponId:     proto.Uint32(target.ID),
		ItemIdList:     []uint32{500},
		SpweaponIdList: []uint32{999999},
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeSpWeapon(&buf, client); err != nil {
		t.Fatalf("UpgradeSpWeapon failed: %v", err)
	}

	resp := &protobuf.SC_14204{}
	decodeTestPacket(t, client, 14204, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, target.ID).First(&stored).Error; err != nil {
		t.Fatalf("expected spweapon to remain: %v", err)
	}
	if stored.TemplateID != 1000 || stored.Pt != 9 {
		t.Fatalf("expected spweapon to remain unchanged")
	}
}
