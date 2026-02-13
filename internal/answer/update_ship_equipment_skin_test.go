package answer_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedEquipSkinTemplate(t *testing.T, skinID uint32, equipTypes []uint32) {
	t.Helper()
	payload, err := json.Marshal(struct {
		ID        uint32   `json:"id"`
		EquipType []uint32 `json:"equip_type"`
	}{
		ID:        skinID,
		EquipType: equipTypes,
	})
	if err != nil {
		t.Fatalf("marshal equip skin payload: %v", err)
	}
	entry := orm.ConfigEntry{
		Category: "ShareCfg/equip_skin_template.json",
		Key:      fmt.Sprintf("%d", skinID),
		Data:     payload,
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", entry.Category, entry.Key, string(entry.Data))
}

func TestUpdateShipEquipmentSkinSuccessPersistClearAndIdempotent(t *testing.T) {
	client := setupEquipTest(t)
	ship := orm.Ship{
		TemplateID:  1001,
		Name:        "Ship",
		EnglishName: "Ship",
		RarityID:    2,
		Star:        1,
		Type:        1,
		Nationality: 1,
		BuildTime:   10,
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[2],"equip_3":[3],"equip_4":[],"equip_5":[],"equip_id_1":0,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipSkinTemplate(t, 12, []uint32{1})

	ownedShip, err := client.Commander.AddShip(ship.TemplateID)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, type) VALUES ($1, $2)", int64(2001), int64(1))
	execAnswerExternalTestSQLT(t, "UPDATE owned_ship_equipments SET equip_id = $1 WHERE owner_id = $2 AND ship_id = $3 AND pos = $4", int64(2001), int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	payload := protobuf.CS_12036{
		ShipId:      proto.Uint32(ownedShip.ID),
		EquipSkinId: proto.Uint32(12),
		Pos:         proto.Uint32(1),
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin failed: %v", err)
	}
	response := &protobuf.SC_12037{}
	decodePacket(t, client, 12037, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}

	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.SkinID != 12 {
		t.Fatalf("expected skin id 12, got %d", entry.SkinID)
	}

	memShip := client.Commander.OwnedShipsMap[ownedShip.ID]
	found := false
	for _, eq := range memShip.Equipments {
		if eq.Pos == 1 {
			found = true
			if eq.SkinID != 12 {
				t.Fatalf("expected in-memory skin id 12 at pos 1, got %d", eq.SkinID)
			}
		}
	}
	if !found {
		t.Fatalf("expected in-memory equipment entry at pos 1")
	}

	// Idempotent update should not duplicate equipment entries.
	buf, err = proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal idempotent payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin idempotent failed: %v", err)
	}
	response = &protobuf.SC_12037{}
	decodePacket(t, client, 12037, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result on idempotent update")
	}
	pos1Count := 0
	for _, eq := range memShip.Equipments {
		if eq.Pos == 1 {
			pos1Count++
		}
	}
	if pos1Count != 1 {
		t.Fatalf("expected 1 in-memory entry for pos 1, got %d", pos1Count)
	}

	// Clear skin.
	payload = protobuf.CS_12036{ShipId: proto.Uint32(ownedShip.ID), EquipSkinId: proto.Uint32(0), Pos: proto.Uint32(1)}
	buf, err = proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal clear payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin clear failed: %v", err)
	}
	response = &protobuf.SC_12037{}
	decodePacket(t, client, 12037, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result on clear")
	}
	entry, err = orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment after clear: %v", err)
	}
	if entry.SkinID != 0 {
		t.Fatalf("expected skin id 0 after clear, got %d", entry.SkinID)
	}
}

func TestUpdateShipEquipmentSkinPreservesEquipID(t *testing.T) {
	client := setupEquipTest(t)
	ship := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[2],"equip_3":[3],"equip_4":[],"equip_5":[],"equip_id_1":0,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipSkinTemplate(t, 12, []uint32{1})

	ownedShip, err := client.Commander.AddShip(ship.TemplateID)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, type) VALUES ($1, $2)", int64(2001), int64(1))
	execAnswerExternalTestSQLT(t, "UPDATE owned_ship_equipments SET equip_id = $1 WHERE owner_id = $2 AND ship_id = $3 AND pos = $4", int64(2001), int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	payload := protobuf.CS_12036{ShipId: proto.Uint32(ownedShip.ID), EquipSkinId: proto.Uint32(12), Pos: proto.Uint32(1)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin failed: %v", err)
	}
	response := &protobuf.SC_12037{}
	decodePacket(t, client, 12037, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result")
	}

	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != 2001 {
		t.Fatalf("expected equip id 2001 to be preserved, got %d", entry.EquipID)
	}
}

func TestUpdateShipEquipmentSkinValidationFailures(t *testing.T) {
	client := setupEquipTest(t)
	ship := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[2],"equip_3":[3],"equip_4":[],"equip_5":[],"equip_id_1":0,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipSkinTemplate(t, 12, []uint32{1})
	seedEquipSkinTemplate(t, 13, []uint32{9})

	ownedShip, err := client.Commander.AddShip(ship.TemplateID)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, type) VALUES ($1, $2)", int64(2001), int64(1))
	execAnswerExternalTestSQLT(t, "UPDATE owned_ship_equipments SET equip_id = $1 WHERE owner_id = $2 AND ship_id = $3 AND pos = $4", int64(2001), int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1))

	// Unknown ship.
	payload := protobuf.CS_12036{ShipId: proto.Uint32(999), EquipSkinId: proto.Uint32(12), Pos: proto.Uint32(1)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal unknown ship payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin unknown ship failed: %v", err)
	}
	resp := &protobuf.SC_12037{}
	decodePacket(t, client, 12037, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result for unknown ship")
	}

	// Out-of-range pos.
	payload = protobuf.CS_12036{ShipId: proto.Uint32(ownedShip.ID), EquipSkinId: proto.Uint32(12), Pos: proto.Uint32(4)}
	buf, err = proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal invalid pos payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin invalid pos failed: %v", err)
	}
	resp = &protobuf.SC_12037{}
	decodePacket(t, client, 12037, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result for invalid pos")
	}

	// Incompatible skin type.
	payload = protobuf.CS_12036{ShipId: proto.Uint32(ownedShip.ID), EquipSkinId: proto.Uint32(13), Pos: proto.Uint32(1)}
	buf, err = proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal incompatible skin payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin incompatible failed: %v", err)
	}
	resp = &protobuf.SC_12037{}
	decodePacket(t, client, 12037, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result for incompatible skin")
	}

	// DB should remain unchanged.
	entries, err := orm.ListOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID)
	if err != nil {
		t.Fatalf("load owned ship equipments: %v", err)
	}
	for _, eq := range entries {
		if eq.SkinID != 0 {
			t.Fatalf("expected skin ids to remain 0 after validation failures")
		}
	}
}

func TestUpdateShipEquipmentSkinValidatesAgainstEquippedItemType(t *testing.T) {
	client := setupEquipTest(t)
	ship := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1,2],"equip_2":[2],"equip_3":[3],"equip_4":[],"equip_5":[],"equip_id_1":0,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipSkinTemplate(t, 12, []uint32{1})
	seedEquipSkinTemplate(t, 14, []uint32{2})

	ownedShip, err := client.Commander.AddShip(ship.TemplateID)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, type) VALUES ($1, $2)", int64(2002), int64(2))
	execAnswerExternalTestSQLT(t, "UPDATE owned_ship_equipments SET equip_id = $1 WHERE owner_id = $2 AND ship_id = $3 AND pos = $4", int64(2002), int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	// Slot allows type 1, but equipped item is type 2; skin type 1 should be rejected.
	payload := protobuf.CS_12036{ShipId: proto.Uint32(ownedShip.ID), EquipSkinId: proto.Uint32(12), Pos: proto.Uint32(1)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin failed: %v", err)
	}
	resp := &protobuf.SC_12037{}
	decodePacket(t, client, 12037, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result for mismatched skin type")
	}

	// Matching type should succeed.
	payload = protobuf.CS_12036{ShipId: proto.Uint32(ownedShip.ID), EquipSkinId: proto.Uint32(14), Pos: proto.Uint32(1)}
	buf, err = proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpdateShipEquipmentSkin(&buf, client); err != nil {
		t.Fatalf("UpdateShipEquipmentSkin failed: %v", err)
	}
	resp = &protobuf.SC_12037{}
	decodePacket(t, client, 12037, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected success result")
	}
}
