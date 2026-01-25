package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestUseItemDropTemplateAddsResources(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	initCommanderMaps(client)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "30121", `{"id":30121,"usage":"usgae_drop_template","usage_arg":[30121,0,1000]}`)
	seedCommanderItem(t, client, 30121, 2)
	seedCommanderResource(t, client, 2, 0)

	payload := protobuf.CS_15002{Id: proto.Uint32(30121), Count: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UseItem(&buffer, client); err != nil {
		t.Fatalf("use item failed: %v", err)
	}

	var response protobuf.SC_15003
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result")
	}
	if len(response.GetDropList()) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(response.GetDropList()))
	}
	drop := response.GetDropList()[0]
	if drop.GetType() != consts.DROP_TYPE_RESOURCE || drop.GetId() != 2 || drop.GetNumber() != 2000 {
		t.Fatalf("unexpected drop: %+v", drop)
	}
	var resource orm.OwnedResource
	if err := orm.GormDB.First(&resource, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 2).Error; err != nil {
		t.Fatalf("load resource: %v", err)
	}
	if resource.Amount != 2000 {
		t.Fatalf("expected oil to be 2000, got %d", resource.Amount)
	}
}

func TestUseItemDropAppointedSelection(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "14001", `{"id":14001,"usage":"usage_drop_appointed","usage_arg":[[2,13001,1],[2,13003,1]]}`)
	seedCommanderItem(t, client, 14001, 1)

	payload := protobuf.CS_15002{Id: proto.Uint32(14001), Count: proto.Uint32(1), Arg: []uint32{2, 13003, 1}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UseItem(&buffer, client); err != nil {
		t.Fatalf("use item failed: %v", err)
	}

	var response protobuf.SC_15003
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result")
	}
	if len(response.GetDropList()) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(response.GetDropList()))
	}
	drop := response.GetDropList()[0]
	if drop.GetType() != consts.DROP_TYPE_ITEM || drop.GetId() != 13003 || drop.GetNumber() != 1 {
		t.Fatalf("unexpected drop: %+v", drop)
	}
	var reward orm.CommanderItem
	if err := orm.GormDB.First(&reward, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 13003).Error; err != nil {
		t.Fatalf("load reward item: %v", err)
	}
	if reward.Count != 1 {
		t.Fatalf("expected reward count 1, got %d", reward.Count)
	}
}

func TestUseItemDropUsesDropRestore(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	initCommanderMaps(client)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "15010", `{"id":15010,"usage":"usage_drop","usage_arg":4901}`)
	seedConfigEntry(t, "ShareCfg/drop_data_restore.json", "1", `{"id":1,"drop_id":4901,"resource_type":1,"resource_num":100,"target_id":0,"target_type":0,"type":1}`)
	seedConfigEntry(t, "ShareCfg/drop_data_restore.json", "2", `{"id":2,"drop_id":4901,"resource_type":2,"resource_num":200,"target_id":0,"target_type":0,"type":1}`)
	seedCommanderItem(t, client, 15010, 2)
	seedCommanderResource(t, client, 1, 0)
	seedCommanderResource(t, client, 2, 0)

	payload := protobuf.CS_15002{Id: proto.Uint32(15010), Count: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UseItem(&buffer, client); err != nil {
		t.Fatalf("use item failed: %v", err)
	}

	var response protobuf.SC_15003
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result")
	}
	if len(response.GetDropList()) != 2 {
		t.Fatalf("expected 2 drops, got %d", len(response.GetDropList()))
	}
	var gold orm.OwnedResource
	if err := orm.GormDB.First(&gold, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).Error; err != nil {
		t.Fatalf("load gold: %v", err)
	}
	if gold.Amount != 200 {
		t.Fatalf("expected gold to be 200, got %d", gold.Amount)
	}
	var oil orm.OwnedResource
	if err := orm.GormDB.First(&oil, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 2).Error; err != nil {
		t.Fatalf("load oil: %v", err)
	}
	if oil.Amount != 400 {
		t.Fatalf("expected oil to be 400, got %d", oil.Amount)
	}
}

func TestUseItemSkinExpGrantsTimedSkin(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedSkin{})
	initCommanderMaps(client)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "68133", `{"id":68133,"usage":"usage_skin_exp","usage_arg":[90001]}`)
	seedConfigEntry(t, "ShareCfg/shop_template.json", "90001", `{"id":90001,"effect_args":[207031],"resource_type":125,"resource_num":1,"time_second":3600}`)
	seedCommanderItem(t, client, 68133, 1)

	start := time.Now()
	payload := protobuf.CS_15002{Id: proto.Uint32(68133), Count: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UseItem(&buffer, client); err != nil {
		t.Fatalf("use item failed: %v", err)
	}

	var owned orm.OwnedSkin
	if err := orm.GormDB.First(&owned, "commander_id = ? AND skin_id = ?", client.Commander.CommanderID, 207031).Error; err != nil {
		t.Fatalf("load skin: %v", err)
	}
	if owned.ExpiresAt == nil {
		t.Fatalf("expected expiry to be set")
	}
	if owned.ExpiresAt.Before(start) || owned.ExpiresAt.After(start.Add(2*time.Hour)) {
		t.Fatalf("unexpected expiry time: %v", owned.ExpiresAt)
	}
}

func TestUseItemSkinDiscountConsumesResources(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.OwnedSkin{})
	initCommanderMaps(client)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "59551", `{"id":59551,"usage":"usage_skin_discount","usage_arg":[[0],780]}`)
	seedConfigEntry(t, "ShareCfg/shop_template.json", "90001", `{"id":90001,"effect_args":[207031],"resource_type":1,"resource_num":1000,"time_second":0}`)
	seedCommanderItem(t, client, 59551, 1)
	seedCommanderResource(t, client, 1, 500)

	payload := protobuf.CS_15002{Id: proto.Uint32(59551), Count: proto.Uint32(1), Arg: []uint32{90001}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UseItem(&buffer, client); err != nil {
		t.Fatalf("use item failed: %v", err)
	}

	var response protobuf.SC_15003
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result")
	}
	if len(response.GetDropList()) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(response.GetDropList()))
	}
	drop := response.GetDropList()[0]
	if drop.GetType() != consts.DROP_TYPE_SKIN || drop.GetId() != 207031 {
		t.Fatalf("unexpected drop: %+v", drop)
	}
	var resource orm.OwnedResource
	if err := orm.GormDB.First(&resource, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).Error; err != nil {
		t.Fatalf("load resource: %v", err)
	}
	if resource.Amount != 280 {
		t.Fatalf("expected gold to be 280, got %d", resource.Amount)
	}
}

func TestQuickExchangeBlueprintOrder(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "30121", `{"id":30121,"usage":"usgae_drop_template","usage_arg":[30121,0,1000]}`)
	seedCommanderItem(t, client, 30121, 1)
	seedCommanderResource(t, client, 2, 0)

	list := []*protobuf.CS_15002{
		{Id: proto.Uint32(30121), Count: proto.Uint32(1)},
		{Id: proto.Uint32(30121), Count: proto.Uint32(0)},
	}
	payload := protobuf.CS_15012{UseList: list}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := QuickExchangeBlueprint(&buffer, client); err != nil {
		t.Fatalf("quick exchange failed: %v", err)
	}

	var response protobuf.SC_15013
	decodeResponse(t, client, &response)
	if len(response.GetRetList()) != 2 {
		t.Fatalf("expected 2 results, got %d", len(response.GetRetList()))
	}
	if response.GetRetList()[0].GetResult() != 0 {
		t.Fatalf("expected first entry to succeed")
	}
	if response.GetRetList()[1].GetResult() == 0 {
		t.Fatalf("expected second entry to fail")
	}
}

func initCommanderMaps(client *connection.Client) {
	client.Commander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)
	client.Commander.MiscItemsMap = make(map[uint32]*orm.CommanderMiscItem)
	client.Commander.OwnedResourcesMap = make(map[uint32]*orm.OwnedResource)
	client.Commander.OwnedShipsMap = make(map[uint32]*orm.OwnedShip)
	client.Commander.OwnedSkinsMap = make(map[uint32]*orm.OwnedSkin)
}

func seedCommanderItem(t *testing.T, client *connection.Client, itemID uint32, count uint32) {
	t.Helper()
	item := orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: itemID, Count: count}
	if err := orm.GormDB.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
	client.Commander.Items = append(client.Commander.Items, item)
	client.Commander.CommanderItemsMap[itemID] = &client.Commander.Items[len(client.Commander.Items)-1]
}

func seedCommanderResource(t *testing.T, client *connection.Client, resourceID uint32, amount uint32) {
	t.Helper()
	resource := orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: resourceID, Amount: amount}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	client.Commander.OwnedResources = append(client.Commander.OwnedResources, resource)
	client.Commander.OwnedResourcesMap[resourceID] = &client.Commander.OwnedResources[len(client.Commander.OwnedResources)-1]
}
