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

func seedFinishPhantomShipTemplate(t *testing.T, templateID uint32) {
	t.Helper()
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (template_id) DO NOTHING", int64(templateID), "Phantom Ship", "Phantom Ship", int64(1), int64(1), int64(1), int64(1), int64(0))
}

func TestFinishPhantomQuestSuccessPersistsAndEmitsShadow(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})
	clearTable(t, &orm.OwnedResource{})

	seedTechnologyShadowUnlock(t, 1, 1, 0)
	seedShipDataStatistics(t, 1001, 9999)
	seedFinishPhantomShipTemplate(t, 1001)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.Energy))
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

	stored := queryAnswerTestInt64(t, "SELECT skin_id FROM owned_ship_shadow_skins WHERE commander_id = $1 AND ship_id = $2 AND shadow_id = $3", int64(client.Commander.CommanderID), int64(owned.ID), int64(1))
	if stored != 9999 {
		t.Fatalf("expected stored skin_id 9999, got %d", stored)
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
	seedFinishPhantomShipTemplate(t, 1001)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.Energy))
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

	count := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM owned_ship_shadow_skins WHERE commander_id = $1 AND ship_id = $2 AND shadow_id = $3", int64(client.Commander.CommanderID), int64(owned.ID), int64(1))
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
	seedFinishPhantomShipTemplate(t, 1001)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.Energy))
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
	seedFinishPhantomShipTemplate(t, 1001)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.Energy))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if err := client.Commander.SetResource(4, 60); err != nil {
		t.Fatalf("seed gems: %v", err)
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
	gems := queryAnswerTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(4))
	if gems != 10 {
		t.Fatalf("expected gems 10, got %d", gems)
	}
}

func TestFinishPhantomQuestGemQuestIdempotentDoesNotDoubleCharge(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedShipShadowSkin{})
	clearTable(t, &orm.OwnedResource{})

	seedTechnologyShadowUnlock(t, 2, 5, 50)
	seedShipDataStatistics(t, 1001, 9999)
	seedFinishPhantomShipTemplate(t, 1001)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.Energy))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if err := client.Commander.SetResource(4, 60); err != nil {
		t.Fatalf("seed gems: %v", err)
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

	gems := queryAnswerTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(4))
	if gems != 10 {
		t.Fatalf("expected gems 10 after retry, got %d", gems)
	}
	count := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM owned_ship_shadow_skins WHERE commander_id = $1 AND ship_id = $2 AND shadow_id = $3", int64(client.Commander.CommanderID), int64(owned.ID), int64(2))
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
	seedFinishPhantomShipTemplate(t, 1001)

	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.Energy))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if err := client.Commander.SetResource(4, 40); err != nil {
		t.Fatalf("seed gems: %v", err)
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
	gems := queryAnswerTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(4))
	if gems != 40 {
		t.Fatalf("expected gems 40, got %d", gems)
	}
	count := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM owned_ship_shadow_skins WHERE commander_id = $1", int64(client.Commander.CommanderID))
	if count != 0 {
		t.Fatalf("expected no stored shadows, got %d", count)
	}
}
