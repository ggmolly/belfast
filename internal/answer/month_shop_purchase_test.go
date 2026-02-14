package answer

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupMonthShopPurchaseTest(t *testing.T, gold uint32) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.MonthShopPurchase{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})

	if err := orm.CreateCommanderRoot(1, 1, "Month Shop Purchase Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 1}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if err := commander.SetResource(1, gold); err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedMonthShopTemplateCore(t *testing.T, ids []uint32) {
	t.Helper()
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		t.Fatalf("marshal ids: %v", err)
	}
	payload := fmt.Sprintf(`{"core_shop_goods":%s,"blueprint_shop_goods":[],"blueprint_shop_limit_goods":[],"honormedal_shop_goods":[],"blueprint_shop_limit_goods_2":[],"blueprint_shop_goods_2":[],"blueprint_shop_limit_goods_3":[],"blueprint_shop_goods_3":[],"blueprint_shop_goods_4":[],"blueprint_shop_limit_goods_4":[]}`,
		string(idsJSON),
	)
	seedConfigEntry(t, "ShareCfg/month_shop_template.json", "0", payload)
}

func seedActivityShopGood(t *testing.T, id uint32, resourceCategory uint32, resourceType uint32, resourceNum uint32, commodityType uint32, commodityID uint32, num uint32, numLimit uint32) {
	t.Helper()
	payload := fmt.Sprintf(`{"id":%d,"resource_category":%d,"resource_type":%d,"resource_num":%d,"commodity_type":%d,"commodity_id":%d,"num":%d,"num_limit":%d}`,
		id,
		resourceCategory,
		resourceType,
		resourceNum,
		commodityType,
		commodityID,
		num,
		numLimit,
	)
	seedConfigEntry(t, "ShareCfg/activity_shop_template.json", fmt.Sprintf("%d", id), payload)
}

func TestMonthShopPurchaseSuccessUpdatesPayCount(t *testing.T) {
	client := setupMonthShopPurchaseTest(t, 100)
	seedMonthShopTemplateCore(t, []uint32{10031})
	seedActivityShopGood(t, 10031, 1, 1, 10, 2, 20001, 1, 2)

	request := &protobuf.CS_16201{Type: proto.Uint32(1), Id: proto.Uint32(10031), Count: proto.Uint32(1)}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := MonthShopPurchase(&buf, client); err != nil {
		t.Fatalf("MonthShopPurchase: %v", err)
	}
	var resp protobuf.SC_16202
	decodePacketAt(t, client, 0, 16202, &resp)
	client.Buffer.Reset()
	if resp.GetResult() != 0 {
		t.Fatalf("expected result=0, got %d", resp.GetResult())
	}
	if len(resp.GetDropList()) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(resp.GetDropList()))
	}
	if client.Commander.GetResourceCount(1) != 90 {
		t.Fatalf("expected gold to be consumed")
	}
	if client.Commander.GetItemCount(20001) != 1 {
		t.Fatalf("expected item reward")
	}

	shopBuf := []byte{}
	if _, _, err := ShopData(&shopBuf, client); err != nil {
		t.Fatalf("ShopData: %v", err)
	}
	var shopResp protobuf.SC_16200
	decodePacketAt(t, client, 0, 16200, &shopResp)
	client.Buffer.Reset()
	if len(shopResp.GetCoreShopList()) != 1 {
		t.Fatalf("expected core shop list to include item")
	}
	if shopResp.GetCoreShopList()[0].GetPayCount() != 1 {
		t.Fatalf("expected pay_count=1, got %d", shopResp.GetCoreShopList()[0].GetPayCount())
	}
}

func TestMonthShopPurchaseInsufficientCurrencyNoStateChange(t *testing.T) {
	client := setupMonthShopPurchaseTest(t, 5)
	seedMonthShopTemplateCore(t, []uint32{10031})
	seedActivityShopGood(t, 10031, 1, 1, 10, 2, 20001, 1, 2)

	request := &protobuf.CS_16201{Type: proto.Uint32(1), Id: proto.Uint32(10031), Count: proto.Uint32(1)}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := MonthShopPurchase(&buf, client); err != nil {
		t.Fatalf("MonthShopPurchase: %v", err)
	}
	var resp protobuf.SC_16202
	decodePacketAt(t, client, 0, 16202, &resp)
	client.Buffer.Reset()
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}
	if client.Commander.GetResourceCount(1) != 5 {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(20001) != 0 {
		t.Fatalf("expected no reward")
	}

	monthKey := uint32(time.Now().Year()*100 + int(time.Now().Month()))
	counts, err := orm.ListMonthShopPurchaseCounts(client.Commander.CommanderID, monthKey)
	if err != nil {
		t.Fatalf("list counts: %v", err)
	}
	if len(counts) != 0 {
		t.Fatalf("expected no purchase count persisted")
	}
}

func TestMonthShopPurchaseLimitReachedNoStateChange(t *testing.T) {
	client := setupMonthShopPurchaseTest(t, 100)
	seedMonthShopTemplateCore(t, []uint32{10031})
	seedActivityShopGood(t, 10031, 1, 1, 10, 2, 20001, 1, 2)

	monthKey := uint32(time.Now().Year()*100 + int(time.Now().Month()))
	execAnswerTestSQLT(t, "INSERT INTO month_shop_purchases (commander_id, goods_id, month, buy_count) VALUES ($1, $2, $3, $4)", int64(client.Commander.CommanderID), int64(10031), int64(monthKey), int64(2))

	request := &protobuf.CS_16201{Type: proto.Uint32(1), Id: proto.Uint32(10031), Count: proto.Uint32(1)}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := MonthShopPurchase(&buf, client); err != nil {
		t.Fatalf("MonthShopPurchase: %v", err)
	}
	var resp protobuf.SC_16202
	decodePacketAt(t, client, 0, 16202, &resp)
	client.Buffer.Reset()
	if resp.GetResult() == 0 {
		t.Fatalf("expected limit failure")
	}
	if client.Commander.GetResourceCount(1) != 100 {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(20001) != 0 {
		t.Fatalf("expected no reward")
	}
}

func TestMonthShopPurchaseFurniturePersistsToDormData(t *testing.T) {
	client := setupMonthShopPurchaseTest(t, 0)
	clearTable(t, &orm.CommanderFurniture{})

	seedMonthShopTemplateCore(t, []uint32{20001})
	seedConfigEntry(t, "ShareCfg/furniture_shop_template.json", "20001", `{"id":20001,"gem_price":10,"dorm_icon_price":0,"time":[[[2021,1,1],[0,0,0]],[[2035,1,1],[0,0,0]]]}`)
	if err := client.Commander.SetResource(4, 20); err != nil {
		t.Fatalf("seed gems: %v", err)
	}

	request := &protobuf.CS_16201{Type: proto.Uint32(1), Id: proto.Uint32(20001), Count: proto.Uint32(1)}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := MonthShopPurchase(&buf, client); err != nil {
		t.Fatalf("MonthShopPurchase: %v", err)
	}
	var resp protobuf.SC_16202
	decodePacketAt(t, client, 0, 16202, &resp)
	client.Buffer.Reset()
	if resp.GetResult() != 0 {
		t.Fatalf("expected result=0, got %d", resp.GetResult())
	}
	if client.Commander.GetResourceCount(4) != 10 {
		t.Fatalf("expected gems to be consumed")
	}

	dormBuf := []byte{}
	if _, _, err := DormData(&dormBuf, client); err != nil {
		t.Fatalf("DormData: %v", err)
	}
	var dormResp protobuf.SC_19001
	decodePacketAt(t, client, 0, 19001, &dormResp)
	client.Buffer.Reset()
	if len(dormResp.GetFurnitureIdList()) != 1 {
		t.Fatalf("expected 1 furniture entry, got %d", len(dormResp.GetFurnitureIdList()))
	}
	if dormResp.GetFurnitureIdList()[0].GetId() != 20001 || dormResp.GetFurnitureIdList()[0].GetCount() != 1 {
		t.Fatalf("expected furniture 20001 count 1")
	}
}
