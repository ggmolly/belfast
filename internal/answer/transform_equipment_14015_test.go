package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func sendCS14015(t *testing.T, client *connection.Client, equipID uint32, upgradeID uint32) *protobuf.SC_14016 {
	t.Helper()
	payload := protobuf.CS_14015{
		EquipId:   proto.Uint32(equipID),
		UpgradeId: proto.Uint32(upgradeID),
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.TransformEquipmentInBag14015(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := &protobuf.SC_14016{}
	decodePacket(t, client, 14016, response)
	return response
}

func TestTransformEquipmentInBagSuccess(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 200)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 2)
	seedEquipment(t, 2001, 1, 0)
	seedEquipment(t, 2002, 1, 0)
	seedEquipUpgradeData(t, 9001, 2001, 2002, 100, [][]uint32{{3001, 2}})

	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2001, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	resp := sendCS14015(t, client, 2001, 9001)
	if resp.GetResult() != 0 {
		t.Fatalf("expected success")
	}
	if client.Commander.GetOwnedEquipment(2001) != nil {
		t.Fatalf("expected source equipment removed")
	}
	owned := client.Commander.GetOwnedEquipment(2002)
	if owned == nil || owned.Count != 1 {
		t.Fatalf("expected target equipment added")
	}
}

func TestTransformEquipmentInBagFailsWrongUpgradePath(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 200)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 2)
	seedEquipment(t, 2001, 1, 0)
	seedEquipment(t, 2002, 1, 0)
	seedEquipUpgradeData(t, 9001, 9999, 2002, 100, [][]uint32{{3001, 2}})

	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2001, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(3001)

	resp := sendCS14015(t, client, 2001, 9001)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	owned := client.Commander.GetOwnedEquipment(2001)
	if owned == nil || owned.Count != 1 {
		t.Fatalf("expected equipment unchanged")
	}
	if client.Commander.GetResourceCount(1) != goldBefore {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(3001) != itemBefore {
		t.Fatalf("expected items unchanged")
	}
}

func TestTransformEquipmentInBagFailsInsufficientGold(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 10)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 2)
	seedEquipment(t, 2001, 1, 0)
	seedEquipment(t, 2002, 1, 0)
	seedEquipUpgradeData(t, 9001, 2001, 2002, 100, [][]uint32{{3001, 2}})

	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2001, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	resp := sendCS14015(t, client, 2001, 9001)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	if client.Commander.GetOwnedEquipment(2001) == nil {
		t.Fatalf("expected equipment unchanged")
	}
}

func TestTransformEquipmentInBagFailsInsufficientMaterial(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 200)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 1)
	seedEquipment(t, 2001, 1, 0)
	seedEquipment(t, 2002, 1, 0)
	seedEquipUpgradeData(t, 9001, 2001, 2002, 100, [][]uint32{{3001, 2}})

	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2001, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(3001)

	resp := sendCS14015(t, client, 2001, 9001)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	owned := client.Commander.GetOwnedEquipment(2001)
	if owned == nil || owned.Count != 1 {
		t.Fatalf("expected equipment unchanged")
	}
	if client.Commander.GetResourceCount(1) != goldBefore {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(3001) != itemBefore {
		t.Fatalf("expected items unchanged")
	}
}
