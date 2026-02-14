package answer_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupRemouldTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_ship_transforms")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_ship_equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_ships")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_resources")
	execAnswerExternalTestSQLT(t, "DELETE FROM commander_items")
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_skins")
	execAnswerExternalTestSQLT(t, "DELETE FROM ships")
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders")
	if err := orm.CreateCommanderRoot(401, 401, "Remould Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 401}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestRemouldShipSuccess(t *testing.T) {
	client := setupRemouldTest(t)
	seedRemouldShips(t)
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 203024, Level: 90}
	if err := mainShip.Create(); err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(mainShip.Level), int64(mainShip.OwnerID), int64(mainShip.ID))
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1}
	if err := materialShip.Create(); err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(materialShip.Level), int64(materialShip.OwnerID), int64(materialShip.ID))
	equipEntry := orm.OwnedShipEquipment{OwnerID: client.Commander.CommanderID, ShipID: materialShip.ID, Pos: 1, EquipID: 3001, SkinID: 0}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_ship_equipments (owner_id, ship_id, pos, equip_id, skin_id) VALUES ($1, $2, $3, $4, $5)", int64(equipEntry.OwnerID), int64(equipEntry.ShipID), int64(equipEntry.Pos), int64(equipEntry.EquipID), int64(equipEntry.SkinID))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(5000))
	execAnswerExternalTestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING", int64(18013), "Remould Item", int64(1), int64(0), int64(1), int64(0))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(18013), int64(1))
	seedTransformConfig(t, 12011, `{"id":12011,"level_limit":85,"star_limit":5,"max_level":1,"use_gold":3000,"use_ship":1,"use_item":[[[18013,1]]],"ship_id":[[203024,203124]],"edit_trans":[],"skin_id":203029,"skill_id":0,"condition_id":[]}`)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12011{
		ShipId:     proto.Uint32(mainShip.ID),
		RemouldId:  proto.Uint32(12011),
		MaterialId: []uint32{materialShip.ID},
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RemouldShip(&buf, client); err != nil {
		t.Fatalf("RemouldShip failed: %v", err)
	}
	response := &protobuf.SC_12012{}
	decodePacket(t, client, 12012, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}

	updated, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 203124 {
		t.Fatalf("expected ship_id 203124, got %d", updated.ShipID)
	}
	if updated.SkinID != 203029 {
		t.Fatalf("expected skin_id 203029, got %d", updated.SkinID)
	}
	transforms, err := orm.ListOwnedShipTransforms(client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load transforms: %v", err)
	}
	if len(transforms) != 1 || transforms[0].TransformID != 12011 || transforms[0].Level != 1 {
		t.Fatalf("expected transform 12011 level 1")
	}
	gold := queryAnswerExternalTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(1))
	if gold != 2000 {
		t.Fatalf("expected gold 2000, got %d", gold)
	}
	item := queryAnswerExternalTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(18013))
	if item != 0 {
		t.Fatalf("expected item count 0, got %d", item)
	}
	if _, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, materialShip.ID); err == nil {
		t.Fatalf("expected material ship to be deleted")
	}
	bag := queryAnswerExternalTestInt64(t, "SELECT count FROM owned_equipments WHERE commander_id = $1 AND equipment_id = $2", int64(client.Commander.CommanderID), int64(3001))
	if bag != 1 {
		t.Fatalf("expected equipment count 1, got %d", bag)
	}
}

func TestRemouldShipInsufficientGold(t *testing.T) {
	client := setupRemouldTest(t)
	seedRemouldShips(t)
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 203024, Level: 90}
	if err := mainShip.Create(); err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(mainShip.Level), int64(mainShip.OwnerID), int64(mainShip.ID))
	seedTransformConfig(t, 12011, `{"id":12011,"level_limit":85,"star_limit":5,"max_level":1,"use_gold":3000,"use_ship":0,"use_item":[],"ship_id":[[203024,203124]],"edit_trans":[],"skin_id":0,"skill_id":0,"condition_id":[]}`)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12011{
		ShipId:    proto.Uint32(mainShip.ID),
		RemouldId: proto.Uint32(12011),
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RemouldShip(&buf, client); err != nil {
		t.Fatalf("RemouldShip failed: %v", err)
	}
	response := &protobuf.SC_12012{}
	decodePacket(t, client, 12012, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	updated, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 203024 {
		t.Fatalf("expected ship_id 203024, got %d", updated.ShipID)
	}
	transforms, err := orm.ListOwnedShipTransforms(client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load transforms: %v", err)
	}
	if len(transforms) != 0 {
		t.Fatalf("expected no transforms")
	}
}

func TestRemouldShipRejectsUnexpectedMaterials(t *testing.T) {
	client := setupRemouldTest(t)
	seedRemouldShips(t)
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 203024, Level: 90}
	if err := mainShip.Create(); err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(mainShip.Level), int64(mainShip.OwnerID), int64(mainShip.ID))
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1}
	if err := materialShip.Create(); err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(materialShip.Level), int64(materialShip.OwnerID), int64(materialShip.ID))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(5000))
	seedTransformConfig(t, 12011, `{"id":12011,"level_limit":85,"star_limit":5,"max_level":1,"use_gold":3000,"use_ship":0,"use_item":[],"ship_id":[[203024,203124]],"edit_trans":[],"skin_id":0,"skill_id":0,"condition_id":[]}`)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12011{
		ShipId:     proto.Uint32(mainShip.ID),
		RemouldId:  proto.Uint32(12011),
		MaterialId: []uint32{materialShip.ID},
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RemouldShip(&buf, client); err != nil {
		t.Fatalf("RemouldShip failed: %v", err)
	}
	response := &protobuf.SC_12012{}
	decodePacket(t, client, 12012, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	if _, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, materialShip.ID); err != nil {
		t.Fatalf("expected material ship to remain: %v", err)
	}
}

func TestRemouldShipRejectsSelfMaterial(t *testing.T) {
	client := setupRemouldTest(t)
	seedRemouldShips(t)
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 203024, Level: 90}
	if err := mainShip.Create(); err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	execAnswerExternalTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(mainShip.Level), int64(mainShip.OwnerID), int64(mainShip.ID))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(5000))
	seedTransformConfig(t, 12011, `{"id":12011,"level_limit":85,"star_limit":5,"max_level":1,"use_gold":3000,"use_ship":1,"use_item":[],"ship_id":[[203024,203124]],"edit_trans":[],"skin_id":0,"skill_id":0,"condition_id":[]}`)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12011{
		ShipId:     proto.Uint32(mainShip.ID),
		RemouldId:  proto.Uint32(12011),
		MaterialId: []uint32{mainShip.ID},
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RemouldShip(&buf, client); err != nil {
		t.Fatalf("RemouldShip failed: %v", err)
	}
	response := &protobuf.SC_12012{}
	decodePacket(t, client, 12012, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	updated, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 203024 {
		t.Fatalf("expected ship_id 203024, got %d", updated.ShipID)
	}
}

func seedRemouldShips(t *testing.T) {
	t.Helper()
	ships := []orm.Ship{
		{TemplateID: 203024, Name: "ShipA", EnglishName: "ShipA", RarityID: 5, Star: 5, Type: 3, Nationality: 1, BuildTime: 10},
		{TemplateID: 203124, Name: "ShipB", EnglishName: "ShipB", RarityID: 5, Star: 5, Type: 3, Nationality: 1, BuildTime: 10},
		{TemplateID: 1001, Name: "Material", EnglishName: "Material", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10},
	}
	for _, ship := range ships {
		execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	}
}

func seedTransformConfig(t *testing.T, id uint32, payload string) {
	t.Helper()
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", "ShareCfg/transform_data_template.json", fmt.Sprintf("%d", id), payload)
}
