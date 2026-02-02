package answer

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func setupConfigTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderTB{})
	clearTable(t, &orm.ActivityPermanentState{})
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 1}}
	return client
}

func clearTable(t *testing.T, model any) {
	t.Helper()
	if err := orm.GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}

func seedConfigEntry(t *testing.T, category string, key string, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{Category: category, Key: key, Data: json.RawMessage(payload)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}

func seedActivityAllowlist(t *testing.T, ids []uint32) {
	t.Helper()
	payload, err := json.Marshal(ids)
	if err != nil {
		t.Fatalf("marshal allowlist failed: %v", err)
	}
	entry := orm.ConfigEntry{Category: "ServerCfg/activities.json", Key: "allowlist", Data: payload}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed allowlist failed: %v", err)
	}
}

func decodeResponse(t *testing.T, client *connection.Client, response proto.Message) {
	t.Helper()
	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected buffer to include header and payload")
	}
	if err := proto.Unmarshal(data[7:], response); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
}

func TestActivitiesUsesConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "1", `{"id":1,"time":["timer",[[2024,1,1],[0,0,0]],[[2024,1,2],[0,0,0]]]}`)
	seedActivityAllowlist(t, []uint32{1})

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(response.GetActivityList()))
	}
	if response.GetActivityList()[0].GetStopTime() == 0 {
		t.Fatalf("expected stop time to be set")
	}
}

func TestActivitiesFiltersFinishedPermanent(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_task_permanent.json", "6000", `{"id":6000}`)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "6000", `{"id":6000,"type":18,"time":"stop"}`)
	seedActivityAllowlist(t, []uint32{6000})
	state := orm.ActivityPermanentState{CommanderID: client.Commander.CommanderID, FinishedActivityIDs: orm.ToInt64List([]uint32{6000})}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed permanent activity state failed: %v", err)
	}

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 0 {
		t.Fatalf("expected finished permanent activity to be filtered")
	}
}

func TestActivitiesIncludesPermanentNow(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_task_permanent.json", "6000", `{"id":6000}`)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "6000", `{"id":6000,"type":18,"time":"stop"}`)
	state := orm.ActivityPermanentState{CommanderID: client.Commander.CommanderID, CurrentActivityID: 6000}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed permanent activity state failed: %v", err)
	}

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 1 || response.GetActivityList()[0].GetId() != 6000 {
		t.Fatalf("expected permanent now activity to be included")
	}
}

func TestActivitiesEmptyWithoutAllowlist(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "1", `{"id":1,"time":["timer",[[2024,1,1],[0,0,0]],[[2024,1,2],[0,0,0]]]}`)

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 0 {
		t.Fatalf("expected empty activity list without allowlist")
	}
}

func TestActivitiesBuildTownActivity(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "1", `{"id":1,"type":116,"time":["timer",[[2024,1,1],[0,0,0]],[[2024,1,2],[0,0,0]]]}`)
	seedConfigEntry(t, "ShareCfg/activity_town_level.json", "1", `{"id":1,"unlock_chara":3,"unlock_work":[[1],[]]}`)
	seedActivityAllowlist(t, []uint32{1})

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(response.GetActivityList()))
	}
	activity := response.GetActivityList()[0]
	if activity.GetData2() != 1 {
		t.Fatalf("expected town activity level 1, got %d", activity.GetData2())
	}
	var hasWorkplaces bool
	for _, entry := range activity.GetDate1KeyValueList() {
		if entry.GetKey() == 1 && len(entry.GetValueList()) > 0 {
			hasWorkplaces = true
			break
		}
	}
	if !hasWorkplaces {
		t.Fatalf("expected town activity workplaces to be populated")
	}
}

func TestActivityOperationSingleEventRefresh(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "10", `{"id":10,"type":112,"config_data":[1001,2001]}`)
	seedConfigEntry(t, "ShareCfg/activity_single_event.json", "1001", `{"id":1001,"type":1}`)
	seedConfigEntry(t, "ShareCfg/activity_single_event.json", "2001", `{"id":2001,"type":2}`)

	request := protobuf.CS_11202{
		ActivityId: proto.Uint32(10),
		Cmd:        proto.Uint32(2),
	}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := ActivityOperation(&buffer, client); err != nil {
		t.Fatalf("activity operation failed: %v", err)
	}

	var response protobuf.SC_11203
	decodeResponse(t, client, &response)
	if len(response.GetNumber()) != 1 || response.GetNumber()[0] != 2001 {
		t.Fatalf("expected daily event list to include 2001")
	}
}

func TestActivitiesBuildBossBattleMark2(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "1", `{"id":1,"type":52,"time":["timer",[[2024,1,1],[0,0,0]],[[2024,1,2],[0,0,0]]]}`)
	seedActivityAllowlist(t, []uint32{1})

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(response.GetActivityList()))
	}
	activity := response.GetActivityList()[0]
	if len(activity.GetDate1KeyValueList()) != 2 {
		t.Fatalf("expected 2 key value lists, got %d", len(activity.GetDate1KeyValueList()))
	}
}

func TestChallengeInfoBuildsConfigResponse(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "1", `{"id":1,"type":37,"config_id":1}`)
	seedConfigEntry(t, "ShareCfg/activity_event_challenge.json", "1", `{"id":1,"buff":[9],"infinite_stage":[[[10001,10002,10003,10004,10005]]]}`)

	request := protobuf.CS_24004{
		ActivityId: proto.Uint32(1),
	}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := ChallengeInfo(&buffer, client); err != nil {
		t.Fatalf("challenge info failed: %v", err)
	}

	var response protobuf.SC_24005
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	if response.GetCurrentChallenge().GetSeasonId() != 1 {
		t.Fatalf("expected season id 1")
	}
	if len(response.GetCurrentChallenge().GetDungeonIdList()) != 5 {
		t.Fatalf("expected dungeon list size 5")
	}
}

func TestAtelierRequestBuildsResponse(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "1", `{"id":1,"type":88}`)

	request := protobuf.CS_26051{
		ActId: proto.Uint32(1),
	}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := AtelierRequest(&buffer, client); err != nil {
		t.Fatalf("atelier request failed: %v", err)
	}

	var response protobuf.SC_26052
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	if len(response.GetItems()) != 0 || len(response.GetRecipes()) != 0 || len(response.GetSlots()) != 0 {
		t.Fatalf("expected empty atelier lists")
	}
}

func TestActivitiesSkipPuzzleWithoutConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "334", `{"id":334,"type":21,"time":["timer",[[2024,1,1],[0,0,0]],[[2024,1,2],[0,0,0]]]}`)
	seedActivityAllowlist(t, []uint32{334})

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 0 {
		t.Fatalf("expected puzzle activity to be skipped")
	}
}

func TestActivitiesSkipNewServerTaskWithoutTasks(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "1", `{"id":1,"type":82,"config_data":[[1001,1002]],"time":["timer",[[2024,1,1],[0,0,0]],[[2024,1,2],[0,0,0]]]}`)
	seedActivityAllowlist(t, []uint32{1})

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 0 {
		t.Fatalf("expected new server task activity to be skipped")
	}
}

func TestActivitiesSkipTaskListWithoutTasks(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "5680", `{"id":5680,"type":13,"config_data":[20820,20821],"time":"stop"}`)
	seedActivityAllowlist(t, []uint32{5680})

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 0 {
		t.Fatalf("expected task list activity to be skipped")
	}
}

func TestActivitiesSkipPuzzleConnectWithoutTimeConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "5691", `{"id":5691,"type":1001,"time":"stop"}`)
	seedActivityAllowlist(t, []uint32{5691})

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 0 {
		t.Fatalf("expected puzzle connect activity to be skipped")
	}
}

func TestActivityOperationNoopForUnsupportedType(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "10", `{"id":10,"type":15}`)

	request := protobuf.CS_11202{
		ActivityId: proto.Uint32(10),
		Cmd:        proto.Uint32(1),
	}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := ActivityOperation(&buffer, client); err != nil {
		t.Fatalf("activity operation failed: %v", err)
	}

	var response protobuf.SC_11203
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestPermanentActivitiesUsesConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/activity_task_permanent.json", "6000", `{"id":6000}`)
	seedConfigEntry(t, "ShareCfg/activity_task_permanent.json", "6001", `{"id":6001}`)
	state := orm.ActivityPermanentState{
		CommanderID:         client.Commander.CommanderID,
		CurrentActivityID:   6001,
		FinishedActivityIDs: orm.ToInt64List([]uint32{6000, 6001}),
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed permanent state failed: %v", err)
	}

	buffer := []byte{}
	if _, _, err := PermanentActivites(&buffer, client); err != nil {
		t.Fatalf("permanent activities failed: %v", err)
	}

	var response protobuf.SC_11210
	decodeResponse(t, client, &response)
	if len(response.GetPermanentActivity()) != 2 || response.GetPermanentActivity()[0] != 6000 || response.GetPermanentActivity()[1] != 6001 {
		t.Fatalf("expected permanent activity list to include 6000 and 6001")
	}
	if response.GetPermanentNow() != 6001 {
		t.Fatalf("expected permanent now to be 6001")
	}
}

func TestEventDataUsesGameRoomConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/game_room_template.json", "1", `{"id":1}`)

	buffer := []byte{}
	if _, _, err := EventData(&buffer, client); err != nil {
		t.Fatalf("event data failed: %v", err)
	}

	var response protobuf.SC_26120
	decodeResponse(t, client, &response)
	if len(response.GetRooms()) != 1 || response.GetRooms()[0].GetRoomid() != 1 {
		t.Fatalf("expected room list to include room 1")
	}
}

func TestShopDataUsesMonthShopConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/month_shop_template.json", "1", `{"core_shop_goods":[100],"blueprint_shop_goods":[200],"blueprint_shop_limit_goods":[201],"honormedal_shop_goods":[300]}`)

	buffer := []byte{}
	if _, _, err := ShopData(&buffer, client); err != nil {
		t.Fatalf("shop data failed: %v", err)
	}

	var response protobuf.SC_16200
	decodeResponse(t, client, &response)
	if len(response.GetCoreShopList()) != 1 || response.GetCoreShopList()[0].GetShopId() != 100 {
		t.Fatalf("expected core shop list to include 100")
	}
	if len(response.GetBlueShopList()) != 2 {
		t.Fatalf("expected blue shop list size 2")
	}
	if len(response.GetNormalShopList()) != 1 || response.GetNormalShopList()[0].GetShopId() != 300 {
		t.Fatalf("expected normal shop list to include 300")
	}
}

func TestShipyardDataUsesBlueprintConfig(t *testing.T) {
	client := setupConfigTest(t)

	buffer := []byte{}
	if _, _, err := ShipyardData(&buffer, client); err != nil {
		t.Fatalf("shipyard data failed: %v", err)
	}

	var response protobuf.SC_63100
	decodeResponse(t, client, &response)
	if len(response.GetBlueprintList()) != 0 {
		t.Fatalf("expected blueprint list to be empty")
	}
}

func TestTechnologyNationProxyUsesFleetTechConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/fleet_tech_group.json", "1", `{"id":1}`)
	seedConfigEntry(t, "ShareCfg/fleet_tech_template.json", "1001", `{"add":[[[1,2],3,4]]}`)

	buffer := []byte{}
	if _, _, err := TechnologyNationProxy(&buffer, client); err != nil {
		t.Fatalf("technology nation proxy failed: %v", err)
	}

	var response protobuf.SC_64000
	decodeResponse(t, client, &response)
	if len(response.GetTechList()) != 1 || response.GetTechList()[0].GetGroupId() != 1 {
		t.Fatalf("expected tech list to include group 1")
	}
	if len(response.GetTechsetList()) != 2 {
		t.Fatalf("expected tech set list size 2")
	}
	if response.GetTechsetList()[0].GetAttrType() != 3 || response.GetTechsetList()[0].GetSetValue() != 4 {
		t.Fatalf("expected tech set to use attr 3 value 4")
	}
}

func TestDormDataUsesDormTemplate(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/dorm_data_template.json", "1", `{"capacity":123}`)

	buffer := []byte{}
	if _, _, err := DormData(&buffer, client); err != nil {
		t.Fatalf("dorm data failed: %v", err)
	}

	var response protobuf.SC_19001
	decodeResponse(t, client, &response)
	if response.GetFloorNum() != 1 {
		t.Fatalf("expected floor num 1")
	}
	if response.GetFoodMaxIncrease() != 123 {
		t.Fatalf("expected food max increase 123")
	}
}

func TestResourcesInfoUsesTemplates(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/oilfield_template.json", "1", `{"level":2,"time":30}`)
	seedConfigEntry(t, "ShareCfg/class_upgrade_template.json", "1", `{"level":3,"time":40}`)
	seedConfigEntry(t, "ShareCfg/navalacademy_data_template.json", "1", `{"id":1}`)
	seedConfigEntry(t, "ShareCfg/navalacademy_shoppingstreet_template.json", "1", `{"special_goods_num":9}`)

	buffer := []byte{}
	if _, _, err := ResourcesInfo(&buffer, client); err != nil {
		t.Fatalf("resources info failed: %v", err)
	}

	var response protobuf.SC_22001
	decodeResponse(t, client, &response)
	if response.GetOilWellLevel() != 2 || response.GetClassLv() != 3 {
		t.Fatalf("expected oil level 2 and class level 3")
	}
	if response.GetSkillClassNum() != 1 || response.GetDailyFinishBuffCnt() != 9 {
		t.Fatalf("expected academy counts to be set")
	}
}

func TestEquipedSpecialWeaponsUsesConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "1", `{"id":1}`)
	seedConfigEntry(t, "ShareCfg/spweapon_data_statistics.json", "2", `{"id":2}`)

	buffer := []byte{}
	if _, _, err := EquipedSpecialWeapons(&buffer, client); err != nil {
		t.Fatalf("special weapons failed: %v", err)
	}

	var response protobuf.SC_14001
	decodeResponse(t, client, &response)
	if response.GetSpweaponBagSize() != 2 {
		t.Fatalf("expected bag size 2")
	}
}

func TestEquippedWeaponSkinUsesConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/equip_skin_template.json", "12", `{"id":12}`)

	buffer := []byte{}
	if _, _, err := EquippedWeaponSkin(&buffer, client); err != nil {
		t.Fatalf("weapon skin failed: %v", err)
	}

	var response protobuf.SC_14101
	decodeResponse(t, client, &response)
	if len(response.GetEquipSkinList()) != 1 || response.GetEquipSkinList()[0].GetId() != 12 {
		t.Fatalf("expected equip skin list to include 12")
	}
}

func TestCommanderManualUsesConfig(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/tutorial_handbook.json", "100", `{"id":100,"tag_list":[1001]}`)
	seedConfigEntry(t, "ShareCfg/tutorial_handbook_task.json", "1001", `{"id":1001,"pt":10}`)

	buffer := []byte{}
	if _, _, err := CommanderManualInfo(&buffer, client); err != nil {
		t.Fatalf("commander manual failed: %v", err)
	}

	var response protobuf.SC_22300
	decodeResponse(t, client, &response)
	if len(response.GetHandbooks()) != 1 {
		t.Fatalf("expected handbooks list size 1")
	}
	if response.GetHandbooks()[0].GetId() != 1001 || response.GetHandbooks()[0].GetPt() != 10 {
		t.Fatalf("expected handbook id 1001 pt 10")
	}
}

func TestNewEducateRequestPersistsTBState(t *testing.T) {
	client := setupConfigTest(t)
	payload := protobuf.CS_29001{Id: proto.Uint32(7)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}

	if _, _, err := NewEducateRequest(&data, client); err != nil {
		t.Fatalf("new educate request failed: %v", err)
	}

	var response protobuf.SC_29002
	decodeResponse(t, client, &response)
	if response.GetTb().GetId() != 7 {
		t.Fatalf("expected tb id 7")
	}
	if _, err := orm.GetCommanderTB(orm.GormDB, client.Commander.CommanderID); err != nil {
		t.Fatalf("expected tb state persisted: %v", err)
	}
}

func TestNewEducateSetCallPersistsName(t *testing.T) {
	client := setupConfigTest(t)
	request := protobuf.CS_29001{Id: proto.Uint32(3)}
	requestData, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}
	if _, _, err := NewEducateRequest(&requestData, client); err != nil {
		t.Fatalf("new educate request failed: %v", err)
	}
	setCall := protobuf.CS_29009{Id: proto.Uint32(3), Name: proto.String("Commander")}
	callData, err := proto.Marshal(&setCall)
	if err != nil {
		t.Fatalf("marshal set call failed: %v", err)
	}
	if _, _, err := NewEducateSetCall(&callData, client); err != nil {
		t.Fatalf("new educate set call failed: %v", err)
	}
	entry, err := orm.GetCommanderTB(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load commander tb failed: %v", err)
	}
	info, _, err := entry.Decode()
	if err != nil {
		t.Fatalf("decode commander tb failed: %v", err)
	}
	if info.GetName() != "Commander" {
		t.Fatalf("expected name Commander, got %q", info.GetName())
	}
}
