package answer

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestStartLearnTacticsSuccess(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	clearTable(t, &orm.CommanderSkillClass{})
	clearTable(t, &orm.CommanderShipSkill{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.Ship{})
	clearTable(t, &orm.ConfigEntry{})

	seedTacticsShip(t, client, 901, 200001)
	seedCommanderItem(t, client, 16001, 2)
	seedTacticsConfig(t, 200001, 501, 1, 3)

	payload := protobuf.CS_22201{RoomId: proto.Uint32(1), ShipId: proto.Uint32(901), SkillPos: proto.Uint32(1), ItemId: proto.Uint32(16001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := StartLearnTactics(&buffer, client); err != nil {
		t.Fatalf("start learn tactics failed: %v", err)
	}

	var response protobuf.SC_22202
	decodeResponse(t, client, &response)
	if response.GetResult() != lessonResultOK {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetClassInfo() == nil {
		t.Fatalf("expected class info in response")
	}
	if response.GetClassInfo().GetShipId() != 901 || response.GetClassInfo().GetRoomId() != 1 || response.GetClassInfo().GetSkillPos() != 1 {
		t.Fatalf("unexpected class info payload")
	}
	if response.GetClassInfo().GetExp() != 150 {
		t.Fatalf("expected lesson exp 150, got %d", response.GetClassInfo().GetExp())
	}

	itemCount := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(16001))
	if itemCount != 1 {
		t.Fatalf("expected item count 1, got %d", itemCount)
	}

	classCount := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM commander_skill_classes WHERE commander_id = $1 AND room_id = $2", int64(client.Commander.CommanderID), int64(1))
	if classCount != 1 {
		t.Fatalf("expected one class row, got %d", classCount)
	}
}

func TestStartLearnTacticsFailsWhenSkillAtMaxLevel(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	clearTable(t, &orm.CommanderSkillClass{})
	clearTable(t, &orm.CommanderShipSkill{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.Ship{})
	clearTable(t, &orm.ConfigEntry{})

	seedTacticsShip(t, client, 902, 200002)
	seedCommanderItem(t, client, 16001, 1)
	seedTacticsConfig(t, 200002, 502, 2, 2)
	execAnswerTestSQLT(t, "INSERT INTO commander_ship_skills (commander_id, ship_id, skill_pos, skill_id, level, exp) VALUES ($1, $2, $3, $4, $5, $6)", int64(client.Commander.CommanderID), int64(902), int64(1), int64(502), int64(2), int64(0))

	payload := protobuf.CS_22201{RoomId: proto.Uint32(1), ShipId: proto.Uint32(902), SkillPos: proto.Uint32(1), ItemId: proto.Uint32(16001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := StartLearnTactics(&buffer, client); err != nil {
		t.Fatalf("start learn tactics failed: %v", err)
	}

	var response protobuf.SC_22202
	decodeResponse(t, client, &response)
	if response.GetResult() != lessonResultFailed {
		t.Fatalf("expected failure result, got %d", response.GetResult())
	}

	itemCount := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(16001))
	if itemCount != 1 {
		t.Fatalf("expected unchanged item count, got %d", itemCount)
	}
}

func TestCancelLearnTacticsSuccess(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.CommanderSkillClass{})
	clearTable(t, &orm.CommanderShipSkill{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.Ship{})
	clearTable(t, &orm.ConfigEntry{})

	seedTacticsShip(t, client, 903, 200003)
	seedTacticsConfig(t, 200003, 503, 1, 3)
	now := uint32(time.Now().UTC().Unix())
	execAnswerTestSQLT(t, "INSERT INTO commander_skill_classes (commander_id, room_id, ship_id, skill_pos, skill_id, start_time, finish_time, exp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(client.Commander.CommanderID), int64(1), int64(903), int64(1), int64(503), int64(now-500), int64(now-1), int64(200))

	payload := protobuf.CS_22203{RoomId: proto.Uint32(1), Type: proto.Uint32(skillCancelTypeManual)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := CancelLearnTactics(&buffer, client); err != nil {
		t.Fatalf("cancel learn tactics failed: %v", err)
	}

	var response protobuf.SC_22204
	decodeResponse(t, client, &response)
	if response.GetResult() != lessonResultOK {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetExp() != 200 {
		t.Fatalf("expected granted exp 200, got %d", response.GetExp())
	}

	classCount := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM commander_skill_classes WHERE commander_id = $1 AND room_id = $2", int64(client.Commander.CommanderID), int64(1))
	if classCount != 0 {
		t.Fatalf("expected class row deleted, got %d", classCount)
	}
	skillLevel := queryAnswerTestInt64(t, "SELECT level FROM commander_ship_skills WHERE commander_id = $1 AND ship_id = $2 AND skill_pos = $3", int64(client.Commander.CommanderID), int64(903), int64(1))
	if skillLevel != 3 {
		t.Fatalf("expected level 3, got %d", skillLevel)
	}
}

func TestCancelLearnTacticsFailsForUnknownRoom(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.CommanderSkillClass{})

	payload := protobuf.CS_22203{RoomId: proto.Uint32(999), Type: proto.Uint32(skillCancelTypeAuto)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := CancelLearnTactics(&buffer, client); err != nil {
		t.Fatalf("cancel learn tactics failed: %v", err)
	}

	var response protobuf.SC_22204
	decodeResponse(t, client, &response)
	if response.GetResult() != lessonResultFailed {
		t.Fatalf("expected failure result, got %d", response.GetResult())
	}
}

func TestResourcesInfoIncludesSkillClassList(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.CommanderSkillClass{})
	seedConfigEntry(t, "ShareCfg/oilfield_template.json", "1", `{"level":1,"time":0}`)
	seedConfigEntry(t, "ShareCfg/class_upgrade_template.json", "1", `{"level":1,"time":0}`)
	seedConfigEntry(t, "ShareCfg/navalacademy_data_template.json", "1", `{"id":1}`)
	seedConfigEntry(t, "ShareCfg/navalacademy_shoppingstreet_template.json", "1", `{"special_goods_num":0}`)
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(300001), "Ship", "Ship", int64(2), int64(1), int64(1), int64(1), int64(1))
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (owner_id, ship_id, id) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(300001), int64(9901))
	execAnswerTestSQLT(t, "INSERT INTO commander_skill_classes (commander_id, room_id, ship_id, skill_pos, skill_id, start_time, finish_time, exp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(client.Commander.CommanderID), int64(1), int64(9901), int64(1), int64(777), int64(10), int64(20), int64(30))

	buffer := []byte{}
	if _, _, err := ResourcesInfo(&buffer, client); err != nil {
		t.Fatalf("resources info failed: %v", err)
	}

	var response protobuf.SC_22001
	decodeResponse(t, client, &response)
	if len(response.GetSkillClassList()) != 1 {
		t.Fatalf("expected one active class, got %d", len(response.GetSkillClassList()))
	}
	if response.GetSkillClassList()[0].GetShipId() != 9901 {
		t.Fatalf("unexpected class ship id %d", response.GetSkillClassList()[0].GetShipId())
	}
}

func seedTacticsShip(t *testing.T, client *connection.Client, ownedShipID uint32, templateID uint32) {
	t.Helper()
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(templateID), "Tactics Ship", "Tactics Ship", int64(3), int64(1), int64(1), int64(1), int64(1))
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (owner_id, ship_id, id) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(templateID), int64(ownedShipID))
	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: templateID, ID: ownedShipID}
	client.Commander.Ships = append(client.Commander.Ships, ship)
	client.Commander.OwnedShipsMap[ownedShipID] = &client.Commander.Ships[len(client.Commander.Ships)-1]
}

func seedTacticsConfig(t *testing.T, templateID uint32, skillID uint32, skillType uint32, maxLevel uint32) {
	t.Helper()
	seedConfigEntry(t, "ShareCfg/navalacademy_data_template.json", "1", `{"id":1}`)
	seedConfigEntry(t, shipTemplateCategory, fmt.Sprintf("%d", templateID), fmt.Sprintf(`{"id":%d,"buff_list_display":[%d]}`, templateID, skillID))
	seedConfigEntry(t, itemConfigCategory, "16001", `{"id":16001,"type":10,"usage":"usage_book","usage_arg":[3600,100,1,50]}`)
	seedConfigEntry(t, skillTemplateCategory, fmt.Sprintf("%d", skillID), fmt.Sprintf(`{"id":%d,"type":%d,"max_level":%d}`, skillID, skillType, maxLevel))
}
