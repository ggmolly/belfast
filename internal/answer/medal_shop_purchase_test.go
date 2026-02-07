package answer

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupMedalShopPurchaseTest(t *testing.T, currency uint32) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.MedalShopGood{})
	clearTable(t, &orm.MedalShopState{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: 1, AccountID: 1, Name: "Medal Shop Purchase Tester"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if currency > 0 {
		if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: 1, ItemID: medalShopCurrencyItemID, Count: currency}).Error; err != nil {
			t.Fatalf("seed currency: %v", err)
		}
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedHonorMedalGoodsList(t *testing.T, group uint32, price uint32, num uint32, goodsType uint32, goods []uint32) {
	t.Helper()
	payload, err := json.Marshal([]honorMedalGoodsListEntry{{
		Group:     group,
		Price:     price,
		Goods:     goods,
		GoodsType: goodsType,
		Num:       num,
		IsShip:    0,
	}})
	if err != nil {
		t.Fatalf("marshal honormedal goods: %v", err)
	}
	seedConfigEntry(t, "ShareCfg/honormedal_goods_list.json", "1", string(payload))
}

func seedMedalShopStateAndGood(t *testing.T, commanderID uint32, flash uint32, goodsID uint32, count uint32) {
	t.Helper()
	if err := orm.GormDB.Create(&orm.MedalShopState{CommanderID: commanderID, NextRefreshTime: flash}).Error; err != nil {
		t.Fatalf("seed state: %v", err)
	}
	if err := orm.GormDB.Create(&orm.MedalShopGood{CommanderID: commanderID, Index: 1, GoodsID: goodsID, Count: count}).Error; err != nil {
		t.Fatalf("seed good: %v", err)
	}
}

func getMedalShopGoodCount(t *testing.T, commanderID uint32, goodsID uint32) uint32 {
	t.Helper()
	var good orm.MedalShopGood
	if err := orm.GormDB.Where("commander_id = ? AND goods_id = ?", commanderID, goodsID).First(&good).Error; err != nil {
		t.Fatalf("load good: %v", err)
	}
	return good.Count
}

func findDrop(drops []*protobuf.DROPINFO, id uint32) *protobuf.DROPINFO {
	for _, d := range drops {
		if d.GetId() == id {
			return d
		}
	}
	return nil
}

func TestMedalShopPurchaseSuccessGrantsDropsAndDecrementsStock(t *testing.T) {
	client := setupMedalShopPurchaseTest(t, 100)
	seedHonorMedalGoodsList(t, 10000, 5, 2, 2, []uint32{20001, 20002})
	seedMedalShopStateAndGood(t, client.Commander.CommanderID, 999, 10000, 5)

	request := &protobuf.CS_16108{
		FlashTime: proto.Uint32(999),
		Shopid:    proto.Uint32(10000),
		Selected: []*protobuf.SELECTED_INFO{
			{Id: proto.Uint32(20001), Count: proto.Uint32(1)},
			{Id: proto.Uint32(20002), Count: proto.Uint32(2)},
		},
	}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := MedalShopPurchase(&buf, client); err != nil {
		t.Fatalf("MedalShopPurchase: %v", err)
	}
	var resp protobuf.SC_16109
	decodePacketAt(t, client, 0, 16109, &resp)
	client.Buffer.Reset()

	if resp.GetResult() != 0 {
		t.Fatalf("expected result=0, got %d", resp.GetResult())
	}
	if len(resp.GetDropList()) != 2 {
		t.Fatalf("expected 2 drops, got %d", len(resp.GetDropList()))
	}
	if drop := findDrop(resp.GetDropList(), 20001); drop == nil || drop.GetNumber() != 2 {
		t.Fatalf("expected drop 20001 x2")
	}
	if drop := findDrop(resp.GetDropList(), 20002); drop == nil || drop.GetNumber() != 4 {
		t.Fatalf("expected drop 20002 x4")
	}
	if client.Commander.GetItemCount(medalShopCurrencyItemID) != 85 {
		t.Fatalf("expected currency consumed")
	}
	if client.Commander.GetItemCount(20001) != 2 || client.Commander.GetItemCount(20002) != 4 {
		t.Fatalf("expected rewards granted")
	}
	if got := getMedalShopGoodCount(t, client.Commander.CommanderID, 10000); got != 2 {
		t.Fatalf("expected stock count=2, got %d", got)
	}
}

func TestMedalShopPurchaseInsufficientCurrencyNoStateChange(t *testing.T) {
	client := setupMedalShopPurchaseTest(t, 10)
	seedHonorMedalGoodsList(t, 10000, 5, 2, 2, []uint32{20001, 20002})
	seedMedalShopStateAndGood(t, client.Commander.CommanderID, 999, 10000, 5)

	request := &protobuf.CS_16108{FlashTime: proto.Uint32(999), Shopid: proto.Uint32(10000), Selected: []*protobuf.SELECTED_INFO{{Id: proto.Uint32(20001), Count: proto.Uint32(3)}}}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := MedalShopPurchase(&buf, client); err != nil {
		t.Fatalf("MedalShopPurchase: %v", err)
	}
	var resp protobuf.SC_16109
	decodePacketAt(t, client, 0, 16109, &resp)
	client.Buffer.Reset()

	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	if client.Commander.GetItemCount(medalShopCurrencyItemID) != 10 {
		t.Fatalf("expected currency unchanged")
	}
	if client.Commander.GetItemCount(20001) != 0 {
		t.Fatalf("expected no reward")
	}
	if got := getMedalShopGoodCount(t, client.Commander.CommanderID, 10000); got != 5 {
		t.Fatalf("expected stock unchanged")
	}
}

func TestMedalShopPurchaseStockExhaustedNoStateChange(t *testing.T) {
	client := setupMedalShopPurchaseTest(t, 100)
	seedHonorMedalGoodsList(t, 10000, 5, 2, 2, []uint32{20001})
	seedMedalShopStateAndGood(t, client.Commander.CommanderID, 999, 10000, 1)

	request := &protobuf.CS_16108{FlashTime: proto.Uint32(999), Shopid: proto.Uint32(10000), Selected: []*protobuf.SELECTED_INFO{{Id: proto.Uint32(20001), Count: proto.Uint32(2)}}}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := MedalShopPurchase(&buf, client); err != nil {
		t.Fatalf("MedalShopPurchase: %v", err)
	}
	var resp protobuf.SC_16109
	decodePacketAt(t, client, 0, 16109, &resp)
	client.Buffer.Reset()

	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	if client.Commander.GetItemCount(medalShopCurrencyItemID) != 100 {
		t.Fatalf("expected currency unchanged")
	}
	if client.Commander.GetItemCount(20001) != 0 {
		t.Fatalf("expected no reward")
	}
	if got := getMedalShopGoodCount(t, client.Commander.CommanderID, 10000); got != 1 {
		t.Fatalf("expected stock unchanged")
	}
}

func TestMedalShopPurchaseStaleFlashTimeNoStateChange(t *testing.T) {
	client := setupMedalShopPurchaseTest(t, 100)
	seedHonorMedalGoodsList(t, 10000, 5, 2, 2, []uint32{20001})
	seedMedalShopStateAndGood(t, client.Commander.CommanderID, 999, 10000, 5)

	request := &protobuf.CS_16108{FlashTime: proto.Uint32(555), Shopid: proto.Uint32(10000), Selected: []*protobuf.SELECTED_INFO{{Id: proto.Uint32(20001), Count: proto.Uint32(1)}}}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := MedalShopPurchase(&buf, client); err != nil {
		t.Fatalf("MedalShopPurchase: %v", err)
	}
	var resp protobuf.SC_16109
	decodePacketAt(t, client, 0, 16109, &resp)
	client.Buffer.Reset()

	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	if client.Commander.GetItemCount(medalShopCurrencyItemID) != 100 {
		t.Fatalf("expected currency unchanged")
	}
	if client.Commander.GetItemCount(20001) != 0 {
		t.Fatalf("expected no reward")
	}
	if got := getMedalShopGoodCount(t, client.Commander.CommanderID, 10000); got != 5 {
		t.Fatalf("expected stock unchanged")
	}
}
