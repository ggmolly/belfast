package answer_test

import (
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
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", category, key, payload)
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

	if err := orm.CreateCommanderRoot(1, 1, "UpgradeSpWeapon Commander", 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO resources (id, item_id, name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", int64(1), int64(0), "Gold")
	execAnswerExternalTestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING", int64(500), "Spweapon PT Item", int64(1), int64(-2), int64(0), int64(0))
	commander := orm.Commander{CommanderID: 1}
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

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(20))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(500), int64(5))

	target := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, ID: 10001, TemplateID: 1000, Pt: 7}
	fodder := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, ID: 10002, TemplateID: 2000, Pt: 1}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_spweapons (owner_id, id, template_id, pt) VALUES ($1, $2, $3, $4)", int64(target.OwnerID), int64(target.ID), int64(target.TemplateID), int64(target.Pt))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_spweapons (owner_id, id, template_id, pt) VALUES ($1, $2, $3, $4)", int64(fodder.OwnerID), int64(fodder.ID), int64(fodder.TemplateID), int64(fodder.Pt))
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

	storedTarget, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, target.ID)
	if err != nil {
		t.Fatalf("failed to load target spweapon: %v", err)
	}
	if storedTarget.TemplateID != 1001 {
		t.Fatalf("expected upgraded template_id 1001, got %d", storedTarget.TemplateID)
	}
	if storedTarget.Pt != 2 {
		t.Fatalf("expected upgraded pt remainder 2, got %d", storedTarget.Pt)
	}

	if _, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, fodder.ID); err == nil {
		t.Fatalf("expected fodder spweapon to be deleted")
	}

	gold := queryAnswerExternalTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(1))
	if gold != 13 {
		t.Fatalf("expected gold 13, got %d", gold)
	}

	item := queryAnswerExternalTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(500))
	if item != 3 {
		t.Fatalf("expected item count 3, got %d", item)
	}
}

func TestUpgradeSpWeaponAccumulatesPtWhenNoUpgradeStep(t *testing.T) {
	client := setupUpgradeSpWeaponClient(t)

	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1000", `{"id":1000,"next":1001,"upgrade_pt":10,"upgrade_use_gold":7}`)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"next":0,"upgrade_pt":0}`)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "500", `{"id":500,"pt":1}`)

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(20))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(500), int64(5))

	target := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, ID: 10003, TemplateID: 1000, Pt: 7}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_spweapons (owner_id, id, template_id, pt) VALUES ($1, $2, $3, $4)", int64(target.OwnerID), int64(target.ID), int64(target.TemplateID), int64(target.Pt))
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

	storedTarget, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, target.ID)
	if err != nil {
		t.Fatalf("failed to load target spweapon: %v", err)
	}
	if storedTarget.TemplateID != 1000 {
		t.Fatalf("expected template_id to remain 1000, got %d", storedTarget.TemplateID)
	}
	if storedTarget.Pt != 9 {
		t.Fatalf("expected pt to accumulate to 9, got %d", storedTarget.Pt)
	}

	gold := queryAnswerExternalTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(1))
	if gold != 20 {
		t.Fatalf("expected gold 20, got %d", gold)
	}

	item := queryAnswerExternalTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(500))
	if item != 3 {
		t.Fatalf("expected item count 3, got %d", item)
	}
}

func TestUpgradeSpWeaponUnknownTargetDoesNotMutate(t *testing.T) {
	client := setupUpgradeSpWeaponClient(t)

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(20))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(500), int64(5))
	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, ID: 10004, TemplateID: 1000, Pt: 7}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_spweapons (owner_id, id, template_id, pt) VALUES ($1, $2, $3, $4)", int64(spweapon.OwnerID), int64(spweapon.ID), int64(spweapon.TemplateID), int64(spweapon.Pt))
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

	stored, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, spweapon.ID)
	if err != nil {
		t.Fatalf("expected spweapon to remain: %v", err)
	}
	if stored.TemplateID != 1000 || stored.Pt != 7 {
		t.Fatalf("expected target spweapon to remain unchanged")
	}

	gold := queryAnswerExternalTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(1))
	if gold != 20 {
		t.Fatalf("expected gold 20, got %d", gold)
	}

	item := queryAnswerExternalTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(500))
	if item != 5 {
		t.Fatalf("expected item count 5, got %d", item)
	}
}

func TestUpgradeSpWeaponInsufficientItemsDoesNotMutate(t *testing.T) {
	client := setupUpgradeSpWeaponClient(t)

	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1000", `{"id":1000,"next":1001,"upgrade_pt":10,"upgrade_use_gold":7}`)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1001", `{"id":1001,"next":0,"upgrade_pt":0}`)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "500", `{"id":500,"pt":1}`)

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(20))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(500), int64(1))
	target := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, ID: 10005, TemplateID: 1000, Pt: 9}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_spweapons (owner_id, id, template_id, pt) VALUES ($1, $2, $3, $4)", int64(target.OwnerID), int64(target.ID), int64(target.TemplateID), int64(target.Pt))
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

	stored, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, target.ID)
	if err != nil {
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

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(20))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(500), int64(5))
	target := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, ID: 10006, TemplateID: 1000, Pt: 9}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_spweapons (owner_id, id, template_id, pt) VALUES ($1, $2, $3, $4)", int64(target.OwnerID), int64(target.ID), int64(target.TemplateID), int64(target.Pt))
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

	stored, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, target.ID)
	if err != nil {
		t.Fatalf("expected spweapon to remain: %v", err)
	}
	if stored.TemplateID != 1000 || stored.Pt != 9 {
		t.Fatalf("expected spweapon to remain unchanged")
	}
}
