package answer

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var playerUpdateCommanderID uint32 = 9000

func setupPlayerUpdateTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.CommanderCommonFlag{})
	clearTable(t, &orm.CommanderStory{})
	clearTable(t, &orm.CommanderAttire{})
	clearTable(t, &orm.CommanderLivingAreaCover{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.CommanderMiscItem{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})

	commanderID := atomic.AddUint32(&playerUpdateCommanderID, 1)
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       30,
		Exp:         0,
		Name:        fmt.Sprintf("Update Tester %d", commanderID),
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestUpdateCommonFlagCommand(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_11019{FlagId: proto.Uint32(1000009)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UpdateCommonFlagCommand(&buffer, client); err != nil {
		t.Fatalf("update common flag failed: %v", err)
	}
	var entry orm.CommanderCommonFlag
	if err := orm.GormDB.First(&entry, "commander_id = ? AND flag_id = ?", client.Commander.CommanderID, 1000009).Error; err != nil {
		t.Fatalf("load flag entry: %v", err)
	}
}

func TestCancelCommonFlagCommand(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	if err := orm.SetCommanderCommonFlag(orm.GormDB, client.Commander.CommanderID, 1000001); err != nil {
		t.Fatalf("seed flag: %v", err)
	}
	payload := protobuf.CS_11021{FlagId: proto.Uint32(1000001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := CancelCommonFlagCommand(&buffer, client); err != nil {
		t.Fatalf("cancel common flag failed: %v", err)
	}
	var entry orm.CommanderCommonFlag
	if err := orm.GormDB.First(&entry, "commander_id = ? AND flag_id = ?", client.Commander.CommanderID, 1000001).Error; err == nil {
		t.Fatalf("expected flag to be removed")
	}
}

func TestUpdateGuideIndex(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_11016{GuideIndex: proto.Uint32(42), Type: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UpdateGuideIndex(&buffer, client); err != nil {
		t.Fatalf("update guide index failed: %v", err)
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.GuideIndex != 42 {
		t.Fatalf("expected guide index 42, got %d", commander.GuideIndex)
	}
}

func TestPlayerInfoBackfillsGuideIndex(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	client.Commander.GuideIndex = 0
	client.Commander.NewGuideIndex = 0
	if err := orm.GormDB.Save(client.Commander).Error; err != nil {
		t.Fatalf("save commander: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedShip{
		OwnerID:           client.Commander.CommanderID,
		ShipID:            202124,
		IsSecretary:       true,
		SecretaryPosition: proto.Uint32(0),
	}).Error; err != nil {
		t.Fatalf("seed secretary: %v", err)
	}

	buffer := []byte{}
	if _, _, err := PlayerInfo(&buffer, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}

	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.GuideIndex != 1 {
		t.Fatalf("expected guide index 1, got %d", commander.GuideIndex)
	}
	if commander.NewGuideIndex != 1 {
		t.Fatalf("expected new guide index 1, got %d", commander.NewGuideIndex)
	}
}

func TestPlayerInfoRandomShipMode(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	client.Commander.RandomShipMode = 3
	if err := orm.GormDB.Save(client.Commander).Error; err != nil {
		t.Fatalf("save commander: %v", err)
	}
	client.Commander.Ships = []orm.OwnedShip{{
		OwnerID:           client.Commander.CommanderID,
		ShipID:            202124,
		IsSecretary:       true,
		SecretaryPosition: proto.Uint32(0),
	}}

	buffer := []byte{}
	if _, _, err := PlayerInfo(&buffer, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	payload := decodeFirstPacketPayload(t, client.Buffer.Bytes())
	var response protobuf.SC_11003
	if err := proto.Unmarshal(payload, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetRandomShipMode() != 3 {
		t.Fatalf("expected random ship mode 3, got %d", response.GetRandomShipMode())
	}
}

func TestPlayerInfoPushesCommanderManual(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	seedConfigEntry(t, "ShareCfg/tutorial_handbook.json", "100", `{"id":100,"tag_list":[1001]}`)
	seedConfigEntry(t, "ShareCfg/tutorial_handbook_task.json", "1001", `{"id":1001,"pt":10}`)
	client.Commander.Ships = []orm.OwnedShip{{
		OwnerID:           client.Commander.CommanderID,
		ShipID:            202124,
		IsSecretary:       true,
		SecretaryPosition: proto.Uint32(0),
	}}

	buffer := []byte{}
	if _, _, err := PlayerInfo(&buffer, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}

	packetIDs := decodePacketIDs(t, client.Buffer.Bytes())
	if len(packetIDs) < 2 {
		t.Fatalf("expected at least 2 packets")
	}
	if packetIDs[0] != 11003 || packetIDs[1] != 22300 {
		t.Fatalf("expected packet ids 11003 and 22300, got %v", packetIDs)
	}
}

func decodePacketIDs(t *testing.T, data []byte) []uint16 {
	t.Helper()
	var packetIDs []uint16
	for offset := 0; offset < len(data); {
		if len(data[offset:]) < 7 {
			t.Fatalf("expected packet header at offset %d", offset)
		}
		payloadSize := int(data[offset])<<8 | int(data[offset+1])
		packetID := uint16(data[offset+3])<<8 | uint16(data[offset+4])
		packetIDs = append(packetIDs, packetID)
		offset += payloadSize + 2
	}
	return packetIDs
}

func decodeFirstPacketPayload(t *testing.T, data []byte) []byte {
	t.Helper()
	if len(data) < 7 {
		t.Fatalf("expected packet header")
	}
	payloadSize := int(data[0])<<8 | int(data[1])
	payloadLen := payloadSize - 5
	if len(data) < payloadLen+7 {
		t.Fatalf("expected packet payload")
	}
	return data[7 : 7+payloadLen]
}

func TestUpdateStory(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_11017{StoryId: proto.Uint32(3001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UpdateStory(&buffer, client); err != nil {
		t.Fatalf("update story failed: %v", err)
	}
	var entry orm.CommanderStory
	if err := orm.GormDB.First(&entry, "commander_id = ? AND story_id = ?", client.Commander.CommanderID, 3001).Error; err != nil {
		t.Fatalf("load story entry: %v", err)
	}
}

func TestChangeManifesto(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_11009{Adv: proto.String("Hello, Commander")}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChangeManifesto(&buffer, client); err != nil {
		t.Fatalf("change manifesto failed: %v", err)
	}
	responsePayload := decodeFirstPacketPayload(t, client.Buffer.Bytes())
	var response protobuf.SC_11010
	if err := proto.Unmarshal(responsePayload, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.Manifesto != "Hello, Commander" {
		t.Fatalf("expected manifesto to persist, got %q", commander.Manifesto)
	}
}

func TestPlayerInfoManifesto(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	client.Commander.Manifesto = "Ready to sortie"
	client.Commander.Ships = []orm.OwnedShip{{
		OwnerID:           client.Commander.CommanderID,
		ShipID:            202124,
		IsSecretary:       true,
		SecretaryPosition: proto.Uint32(0),
	}}
	buffer := []byte{}
	if _, _, err := PlayerInfo(&buffer, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	responsePayload := decodeFirstPacketPayload(t, client.Buffer.Bytes())
	var response protobuf.SC_11003
	if err := proto.Unmarshal(responsePayload, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetAdv() != "Ready to sortie" {
		t.Fatalf("expected manifesto to be returned, got %q", response.GetAdv())
	}
}

func TestUpdateStoryList(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_11032{StoryIds: []uint32{3001, 3002}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UpdateStoryList(&buffer, client); err != nil {
		t.Fatalf("update story list failed: %v", err)
	}
	for _, storyID := range payload.StoryIds {
		var entry orm.CommanderStory
		if err := orm.GormDB.First(&entry, "commander_id = ? AND story_id = ?", client.Commander.CommanderID, storyID).Error; err != nil {
			t.Fatalf("load story entry %d: %v", storyID, err)
		}
	}
	responsePayload := protobuf.SC_11033{}
	responseData := decodeFirstPacketPayload(t, client.Buffer.Bytes())
	if err := proto.Unmarshal(responseData, &responsePayload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if responsePayload.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", responsePayload.GetResult())
	}
}

func TestUpdateStoryListIdempotent(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	if err := orm.AddCommanderStory(orm.GormDB, client.Commander.CommanderID, 3001); err != nil {
		t.Fatalf("seed story: %v", err)
	}
	payload := protobuf.CS_11032{StoryIds: []uint32{3001}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UpdateStoryList(&buffer, client); err != nil {
		t.Fatalf("update story list failed: %v", err)
	}
	var count int64
	if err := orm.GormDB.Model(&orm.CommanderStory{}).Where("commander_id = ? AND story_id = ?", client.Commander.CommanderID, 3001).Count(&count).Error; err != nil {
		t.Fatalf("count stories: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 story entry, got %d", count)
	}
	responsePayload := protobuf.SC_11033{}
	responseData := decodeFirstPacketPayload(t, client.Buffer.Bytes())
	if err := proto.Unmarshal(responseData, &responsePayload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if responsePayload.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", responsePayload.GetResult())
	}
}

func TestChangeLivingAreaCover(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	seedConfigEntry(t, "ShareCfg/livingarea_cover.json", "100", `{"id":100}`)
	if err := orm.UpsertCommanderLivingAreaCover(orm.GormDB, orm.CommanderLivingAreaCover{CommanderID: client.Commander.CommanderID, CoverID: 100}); err != nil {
		t.Fatalf("seed cover: %v", err)
	}
	payload := protobuf.CS_11030{LivingareaCoverId: proto.Uint32(100)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChangeLivingAreaCover(&buffer, client); err != nil {
		t.Fatalf("change living area cover failed: %v", err)
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.LivingAreaCoverID != 100 {
		t.Fatalf("expected cover id 100, got %d", commander.LivingAreaCoverID)
	}
}

func TestAttireApply(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	entry := orm.CommanderAttire{CommanderID: client.Commander.CommanderID, Type: consts.AttireTypeIconFrame, AttireID: 101}
	if err := orm.UpsertCommanderAttire(orm.GormDB, entry); err != nil {
		t.Fatalf("seed attire: %v", err)
	}
	payload := protobuf.CS_11005{Type: proto.Uint32(consts.AttireTypeIconFrame), Id: proto.Uint32(101)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := AttireApply(&buffer, client); err != nil {
		t.Fatalf("attire apply failed: %v", err)
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.SelectedIconFrameID != 101 {
		t.Fatalf("expected icon frame 101, got %d", commander.SelectedIconFrameID)
	}
}

func TestChangePlayerName(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	seedConfigEntry(t, "ShareCfg/gameset.json", "player_name_change_lv_limit", `{"key_value":1,"description":""}`)
	seedConfigEntry(t, "ShareCfg/gameset.json", "player_name_cold_time", `{"key_value":60,"description":""}`)
	seedConfigEntry(t, "ShareCfg/gameset.json", "player_name_change_cost", `{"key_value":0,"description":[2,15009,1]}`)
	item := orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 15009, Count: 1}
	if err := orm.GormDB.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
	client.Commander.Items = []orm.CommanderItem{item}
	client.Commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{15009: &client.Commander.Items[0]}
	client.Commander.MiscItemsMap = map[uint32]*orm.CommanderMiscItem{}

	payload := protobuf.CS_11007{Type: proto.Uint32(1), Name: proto.String("NewCommander")}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChangePlayerName(&buffer, client); err != nil {
		t.Fatalf("change player name failed: %v", err)
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.Name != "NewCommander" {
		t.Fatalf("expected name updated")
	}
	var updatedItem orm.CommanderItem
	if err := orm.GormDB.First(&updatedItem, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 15009).Error; err != nil {
		t.Fatalf("load item: %v", err)
	}
	if updatedItem.Count != 0 {
		t.Fatalf("expected item count 0, got %d", updatedItem.Count)
	}
}

func TestUpdateSecretariesPhantom(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 101}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	client.Commander.Ships = []orm.OwnedShip{ship}
	client.Commander.OwnedShipsMap = map[uint32]*orm.OwnedShip{ship.ID: &client.Commander.Ships[0]}

	entries := []*protobuf.KVDATA{{Key: proto.Uint32(ship.ID), Value: proto.Uint32(2)}}
	payload := protobuf.CS_11011{Character: entries}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UpdateSecretaries(&buffer, client); err != nil {
		t.Fatalf("update secretaries failed: %v", err)
	}
	var updated orm.OwnedShip
	if err := orm.GormDB.First(&updated, ship.ID).Error; err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.SecretaryPhantomID != 2 || !updated.IsSecretary {
		t.Fatalf("expected secretary phantom id to be set")
	}
}

func TestUpdateGuideIndexNew(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_11016{GuideIndex: proto.Uint32(77), Type: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UpdateGuideIndex(&buffer, client); err != nil {
		t.Fatalf("update guide index failed: %v", err)
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.NewGuideIndex != 77 {
		t.Fatalf("expected new guide index 77, got %d", commander.NewGuideIndex)
	}
}

func TestChangePlayerNameForcedClearsFlag(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	if err := orm.SetCommanderCommonFlag(orm.GormDB, client.Commander.CommanderID, consts.IllegalityPlayerName); err != nil {
		t.Fatalf("seed flag: %v", err)
	}
	payload := protobuf.CS_11007{Type: proto.Uint32(2), Name: proto.String("ForcedName")}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChangePlayerName(&buffer, client); err != nil {
		t.Fatalf("change player name failed: %v", err)
	}
	var entry orm.CommanderCommonFlag
	if err := orm.GormDB.First(&entry, "commander_id = ? AND flag_id = ?", client.Commander.CommanderID, consts.IllegalityPlayerName).Error; err == nil {
		t.Fatalf("expected illegality flag removed")
	}
}

func TestAttireApplyExpired(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	expiresAt := time.Now().Add(-time.Hour)
	entry := orm.CommanderAttire{CommanderID: client.Commander.CommanderID, Type: consts.AttireTypeChatFrame, AttireID: 200, ExpiresAt: &expiresAt}
	if err := orm.UpsertCommanderAttire(orm.GormDB, entry); err != nil {
		t.Fatalf("seed attire: %v", err)
	}
	payload := protobuf.CS_11005{Type: proto.Uint32(consts.AttireTypeChatFrame), Id: proto.Uint32(200)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, _, err = AttireApply(&buffer, client)
	if err != nil {
		t.Fatalf("attire apply failed: %v", err)
	}
	var response protobuf.SC_11006
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected apply to fail for expired attire")
	}
}

func TestChangePlayerNameInvalidGameset(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	seedConfigEntry(t, "ShareCfg/gameset.json", "player_name_change_lv_limit", `{"key_value":1,"description":""}`)
	payload := protobuf.CS_11007{Type: proto.Uint32(1), Name: proto.String("InvalidName")}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, _, err = ChangePlayerName(&buffer, client)
	if err != nil {
		t.Fatalf("change player name failed: %v", err)
	}
	var response protobuf.SC_11008
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure without gameset config")
	}
}

func TestChangeLivingAreaCoverMissingConfig(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	if err := orm.UpsertCommanderLivingAreaCover(orm.GormDB, orm.CommanderLivingAreaCover{CommanderID: client.Commander.CommanderID, CoverID: 500}); err != nil {
		t.Fatalf("seed cover: %v", err)
	}
	payload := protobuf.CS_11030{LivingareaCoverId: proto.Uint32(500)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, _, err = ChangeLivingAreaCover(&buffer, client)
	if err != nil {
		t.Fatalf("change living area cover failed: %v", err)
	}
	var response protobuf.SC_11031
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure without config")
	}
}
