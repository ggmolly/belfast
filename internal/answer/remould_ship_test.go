package answer_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func setupRemouldTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearRemouldTable(t, &orm.OwnedShipTransform{})
	clearRemouldTable(t, &orm.OwnedShipEquipment{})
	clearRemouldTable(t, &orm.OwnedShip{})
	clearRemouldTable(t, &orm.OwnedEquipment{})
	clearRemouldTable(t, &orm.OwnedResource{})
	clearRemouldTable(t, &orm.CommanderItem{})
	clearRemouldTable(t, &orm.ConfigEntry{})
	clearRemouldTable(t, &orm.OwnedSkin{})
	clearRemouldTable(t, &orm.Ship{})
	clearRemouldTable(t, &orm.Commander{})
	commander := orm.Commander{CommanderID: 401, AccountID: 401, Name: "Remould Tester"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestRemouldShipSuccess(t *testing.T) {
	client := setupRemouldTest(t)
	seedRemouldShips(t)
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 203024, Level: 90}
	if err := orm.GormDB.Create(&mainShip).Error; err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1}
	if err := orm.GormDB.Create(&materialShip).Error; err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	equipEntry := orm.OwnedShipEquipment{OwnerID: client.Commander.CommanderID, ShipID: materialShip.ID, Pos: 1, EquipID: 3001, SkinID: 0}
	if err := orm.GormDB.Create(&equipEntry).Error; err != nil {
		t.Fatalf("seed ship equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 5000}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 18013, Count: 1}).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
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

	var updated orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, mainShip.ID).First(&updated).Error; err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 203124 {
		t.Fatalf("expected ship_id 203124, got %d", updated.ShipID)
	}
	if updated.SkinID != 203029 {
		t.Fatalf("expected skin_id 203029, got %d", updated.SkinID)
	}
	transforms, err := orm.ListOwnedShipTransforms(orm.GormDB, client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load transforms: %v", err)
	}
	if len(transforms) != 1 || transforms[0].TransformID != 12011 || transforms[0].Level != 1 {
		t.Fatalf("expected transform 12011 level 1")
	}
	var gold orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).First(&gold).Error; err != nil {
		t.Fatalf("load gold: %v", err)
	}
	if gold.Amount != 2000 {
		t.Fatalf("expected gold 2000, got %d", gold.Amount)
	}
	var item orm.CommanderItem
	if err := orm.GormDB.Where("commander_id = ? AND item_id = ?", client.Commander.CommanderID, 18013).First(&item).Error; err != nil {
		t.Fatalf("load item: %v", err)
	}
	if item.Count != 0 {
		t.Fatalf("expected item count 0, got %d", item.Count)
	}
	var materialCheck orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, materialShip.ID).First(&materialCheck).Error; err == nil {
		t.Fatalf("expected material ship to be deleted")
	}
	var bag orm.OwnedEquipment
	if err := orm.GormDB.Where("commander_id = ? AND equipment_id = ?", client.Commander.CommanderID, 3001).First(&bag).Error; err != nil {
		t.Fatalf("load owned equipment: %v", err)
	}
	if bag.Count != 1 {
		t.Fatalf("expected equipment count 1, got %d", bag.Count)
	}
}

func TestRemouldShipInsufficientGold(t *testing.T) {
	client := setupRemouldTest(t)
	seedRemouldShips(t)
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 203024, Level: 90}
	if err := orm.GormDB.Create(&mainShip).Error; err != nil {
		t.Fatalf("create main ship: %v", err)
	}
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
	var updated orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, mainShip.ID).First(&updated).Error; err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 203024 {
		t.Fatalf("expected ship_id 203024, got %d", updated.ShipID)
	}
	transforms, err := orm.ListOwnedShipTransforms(orm.GormDB, client.Commander.CommanderID, mainShip.ID)
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
	if err := orm.GormDB.Create(&mainShip).Error; err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1}
	if err := orm.GormDB.Create(&materialShip).Error; err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 5000}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
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
	var materialCheck orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, materialShip.ID).First(&materialCheck).Error; err != nil {
		t.Fatalf("expected material ship to remain: %v", err)
	}
}

func TestRemouldShipRejectsSelfMaterial(t *testing.T) {
	client := setupRemouldTest(t)
	seedRemouldShips(t)
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 203024, Level: 90}
	if err := orm.GormDB.Create(&mainShip).Error; err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 5000}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
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
	var updated orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, mainShip.ID).First(&updated).Error; err != nil {
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
		if err := orm.GormDB.Create(&ship).Error; err != nil {
			t.Fatalf("seed ship %d: %v", ship.TemplateID, err)
		}
	}
}

func seedTransformConfig(t *testing.T, id uint32, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: "ShareCfg/transform_data_template.json", Key: fmt.Sprintf("%d", id), Data: json.RawMessage(payload)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed transform config failed: %v", err)
	}
}

func clearRemouldTable(t *testing.T, model any) {
	t.Helper()
	if err := orm.GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}
