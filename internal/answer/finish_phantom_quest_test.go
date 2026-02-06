package answer

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedTechnologyShadowUnlock(t *testing.T, id uint32, questType uint32, targetNum uint32) {
	t.Helper()
	seedConfigEntry(t, "ShareCfg/technology_shadow_unlock.json", fmt.Sprintf("%d", id), fmt.Sprintf(`{"id":%d,"type":%d,"target_num":%d}`, id, questType, targetNum))
}

func seedShipDataStatistics(t *testing.T, templateID uint32, skinID uint32) {
	t.Helper()
	seedConfigEntry(t, "sharecfgdata/ship_data_statistics.json", fmt.Sprintf("%d", templateID), fmt.Sprintf(`{"id":%d,"skin_id":%d}`, templateID, skinID))
}

func TestFinishPhantomQuestSuccessPersistsAndEmitsShadow(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})
	clearTable(t, &orm.OwnedResource{})

	seedTechnologyShadowUnlock(t, 1, 1, 0)
	seedShipDataStatistics(t, 1001, 9999)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12210{ShipId: proto.Uint32(owned.ID), SkinShadowId: proto.Uint32(1)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest failed: %v", err)
	}
	response := &protobuf.SC_12211{}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var stored orm.OwnedShipShadowSkin
	if err := orm.GormDB.First(&stored, "commander_id = ? AND ship_id = ? AND shadow_id = ?", client.Commander.CommanderID, owned.ID, 1).Error; err != nil {
		t.Fatalf("load stored shadow: %v", err)
	}
	if stored.SkinID != 9999 {
		t.Fatalf("expected stored skin_id 9999, got %d", stored.SkinID)
	}

	shadows, err := orm.ListOwnedShipShadowSkins(client.Commander.CommanderID, []uint32{owned.ID})
	if err != nil {
		t.Fatalf("list shadows: %v", err)
	}
	info := orm.ToProtoOwnedShip(owned, nil, shadows[owned.ID])
	if len(info.GetSkinShadowList()) != 1 {
		t.Fatalf("expected 1 skin_shadow_list entry, got %d", len(info.GetSkinShadowList()))
	}
	entry := info.GetSkinShadowList()[0]
	if entry.GetKey() != 1 || entry.GetValue() != 9999 {
		t.Fatalf("unexpected skin_shadow_list entry: key=%d value=%d", entry.GetKey(), entry.GetValue())
	}
}

func TestFinishPhantomQuestIdempotent(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})
	clearTable(t, &orm.OwnedResource{})

	seedTechnologyShadowUnlock(t, 1, 1, 0)
	seedShipDataStatistics(t, 1001, 9999)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12210{ShipId: proto.Uint32(owned.ID), SkinShadowId: proto.Uint32(1)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest first failed: %v", err)
	}
	response := &protobuf.SC_12211{}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest second failed: %v", err)
	}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var count int64
	if err := orm.GormDB.Model(&orm.OwnedShipShadowSkin{}).Where("commander_id = ? AND ship_id = ? AND shadow_id = ?", client.Commander.CommanderID, owned.ID, 1).Count(&count).Error; err != nil {
		t.Fatalf("count stored shadow: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 stored row, got %d", count)
	}
}

func TestFinishPhantomQuestFailsWhenShipNotOwned(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})

	seedTechnologyShadowUnlock(t, 1, 1, 0)

	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	payload := protobuf.CS_12210{ShipId: proto.Uint32(9999), SkinShadowId: proto.Uint32(1)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest failed: %v", err)
	}
	response := &protobuf.SC_12211{}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
}

func TestFinishPhantomQuestFailsWhenShadowIDUnknown(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})

	seedShipDataStatistics(t, 1001, 9999)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	payload := protobuf.CS_12210{ShipId: proto.Uint32(owned.ID), SkinShadowId: proto.Uint32(42)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest failed: %v", err)
	}
	response := &protobuf.SC_12211{}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
}

func TestFinishPhantomQuestGemQuestConsumesGems(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})
	clearTable(t, &orm.OwnedResource{})

	seedTechnologyShadowUnlock(t, 2, 5, 50)
	seedShipDataStatistics(t, 1001, 9999)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 4, Amount: 60}).Error; err != nil {
		t.Fatalf("seed gems: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	payload := protobuf.CS_12210{ShipId: proto.Uint32(owned.ID), SkinShadowId: proto.Uint32(2)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest failed: %v", err)
	}
	response := &protobuf.SC_12211{}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	var gems orm.OwnedResource
	if err := orm.GormDB.First(&gems, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 4).Error; err != nil {
		t.Fatalf("load gems: %v", err)
	}
	if gems.Amount != 10 {
		t.Fatalf("expected gems 10, got %d", gems.Amount)
	}
}

func TestFinishPhantomQuestGemQuestIdempotentDoesNotDoubleCharge(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})
	clearTable(t, &orm.OwnedResource{})

	seedTechnologyShadowUnlock(t, 2, 5, 50)
	seedShipDataStatistics(t, 1001, 9999)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 4, Amount: 60}).Error; err != nil {
		t.Fatalf("seed gems: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12210{ShipId: proto.Uint32(owned.ID), SkinShadowId: proto.Uint32(2)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest first failed: %v", err)
	}
	response := &protobuf.SC_12211{}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest second failed: %v", err)
	}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var gems orm.OwnedResource
	if err := orm.GormDB.First(&gems, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 4).Error; err != nil {
		t.Fatalf("load gems: %v", err)
	}
	if gems.Amount != 10 {
		t.Fatalf("expected gems 10 after retry, got %d", gems.Amount)
	}
	var count int64
	if err := orm.GormDB.Model(&orm.OwnedShipShadowSkin{}).Where("commander_id = ? AND ship_id = ? AND shadow_id = ?", client.Commander.CommanderID, owned.ID, 2).Count(&count).Error; err != nil {
		t.Fatalf("count stored shadow: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 stored row, got %d", count)
	}
}

func TestFinishPhantomQuestGemQuestInsufficientGemsNoMutation(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})
	clearTable(t, &orm.OwnedResource{})

	seedTechnologyShadowUnlock(t, 2, 5, 50)
	seedShipDataStatistics(t, 1001, 9999)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 4, Amount: 40}).Error; err != nil {
		t.Fatalf("seed gems: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	payload := protobuf.CS_12210{ShipId: proto.Uint32(owned.ID), SkinShadowId: proto.Uint32(2)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := FinishPhantomQuest(&buf, client); err != nil {
		t.Fatalf("FinishPhantomQuest failed: %v", err)
	}
	response := &protobuf.SC_12211{}
	decodePacket(t, client, 12211, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	var gems orm.OwnedResource
	if err := orm.GormDB.First(&gems, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 4).Error; err != nil {
		t.Fatalf("load gems: %v", err)
	}
	if gems.Amount != 40 {
		t.Fatalf("expected gems 40, got %d", gems.Amount)
	}
	var count int64
	if err := orm.GormDB.Model(&orm.OwnedShipShadowSkin{}).Where("commander_id = ?", client.Commander.CommanderID).Count(&count).Error; err != nil {
		t.Fatalf("count stored shadows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no stored shadows, got %d", count)
	}
}
