package answer

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/misc"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/region"
	"google.golang.org/protobuf/proto"
)

func setupHandlerCommander(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.Commander{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.OwnedSkin{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.CommanderMiscItem{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.OwnedSpWeapon{})
	clearTable(t, &orm.Mail{})
	clearTable(t, &orm.MailAttachment{})
	clearTable(t, &orm.Notice{})
	clearTable(t, &orm.Build{})
	clearTable(t, &orm.Ship{})
	clearTable(t, &orm.RandomFlagShip{})
	clearTable(t, &orm.Like{})
	clearTable(t, &orm.CommanderAppreciationState{})
	clearTable(t, &orm.CommanderMedalDisplay{})
	clearTable(t, &orm.CommanderTrophyProgress{})
	clearTable(t, &orm.CommanderStoreupAwardProgress{})
	clearTable(t, &orm.SecondaryPasswordState{})
	clearTable(t, &orm.ActivityPermanentState{})
	clearTable(t, &orm.EscortState{})
	commanderID := uint32(time.Now().UnixNano())
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   commanderID,
		Name:        fmt.Sprintf("Handler Commander %d", commanderID),
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	client := &connection.Client{Commander: &commander}
	client.Server = connection.NewServer("127.0.0.1", 0, func(pkt *[]byte, c *connection.Client, size int) {})
	return client
}

func seedShipTemplate(t *testing.T, templateID uint32, poolID uint32, rarity uint32, shipType uint32, englishName string, star uint32) {
	t.Helper()
	var existing orm.Ship
	if err := orm.GormDB.First(&existing, "template_id = ?", templateID).Error; err == nil {
		return
	}
	ship := orm.Ship{
		TemplateID:  templateID,
		Name:        fmt.Sprintf("Ship %d", templateID),
		EnglishName: englishName,
		RarityID:    rarity,
		Star:        star,
		Type:        shipType,
		Nationality: 1,
		BuildTime:   1,
		PoolID:      &poolID,
	}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
}

func seedOwnedShip(t *testing.T, client *connection.Client, shipTemplateID uint32) *orm.OwnedShip {
	t.Helper()
	ship := orm.OwnedShip{
		OwnerID: client.Commander.CommanderID,
		ShipID:  shipTemplateID,
	}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	return client.Commander.OwnedShipsMap[ship.ID]
}

func seedHandlerCommanderItem(t *testing.T, client *connection.Client, itemID uint32, count uint32) {
	t.Helper()
	entry := orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: itemID, Count: count}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
}

func seedHandlerCommanderResource(t *testing.T, client *connection.Client, resourceID uint32, amount uint32) {
	t.Helper()
	entry := orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: resourceID, Amount: amount}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
}

func parseVersionForTest(t *testing.T, version string) [4]uint32 {
	t.Helper()
	var parts [4]uint32
	segments := strings.Split(version, ".")
	if len(segments) < 3 || len(segments) > 4 {
		t.Fatalf("unexpected version format %q", version)
	}
	for i, segment := range segments {
		value, err := strconv.ParseUint(segment, 10, 32)
		if err != nil {
			t.Fatalf("parse version segment %q: %v", segment, err)
		}
		parts[i] = uint32(value)
	}
	return parts
}

func TestTrackingNoops(t *testing.T) {
	client := setupHandlerCommander(t)
	buffer := []byte{}
	if _, _, err := NewTracking(&buffer, client); err != nil {
		t.Fatalf("new tracking failed: %v", err)
	}
	if _, _, err := MainSceneTracking(&buffer, client); err != nil {
		t.Fatalf("main scene tracking failed: %v", err)
	}
	if _, _, err := TrackCommand(&buffer, client); err != nil {
		t.Fatalf("track command failed: %v", err)
	}
	if _, _, err := UrExchangeTracking(&buffer, client); err != nil {
		t.Fatalf("ur exchange tracking failed: %v", err)
	}
}

func TestSimpleResponseHandlers(t *testing.T) {
	client := setupHandlerCommander(t)
	cases := []struct {
		name    string
		handler func(*[]byte, *connection.Client) (int, int, error)
		payload proto.Message
	}{
		{name: "SendHeartbeat", handler: SendHeartbeat},
		{name: "ChargeCommandAnswer", handler: ChargeCommandAnswer},
		{name: "FetchSecondaryPasswordCommandResponse", handler: FetchSecondaryPasswordCommandResponse},
		{name: "WeeklyMissions", handler: WeeklyMissions},
		{name: "WorldBaseInfo", handler: WorldBaseInfo},
		{name: "WorldBossInfo", handler: WorldBossInfo},
		{name: "WorldCheckInfo", handler: WorldCheckInfo},
		{name: "MiniGameHubData", handler: MiniGameHubData},
		{name: "GetChargeList", handler: GetChargeList},
		{name: "GetRefundInfo", handler: GetRefundInfo},
		{name: "LimitChallengeInfo", handler: LimitChallengeInfo},
		{name: "TechnologyRefreshList", handler: TechnologyRefreshList},
		{name: "Meowfficers", handler: Meowfficers},
		{name: "FleetEnergyRecoverTime", handler: FleetEnergyRecoverTime},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client.Buffer.Reset()
			buffer := []byte{}
			if tc.payload != nil {
				data, err := proto.Marshal(tc.payload)
				if err != nil {
					t.Fatalf("marshal payload: %v", err)
				}
				buffer = data
			}
			if _, _, err := tc.handler(&buffer, client); err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}
			if client.Buffer.Len() == 0 {
				t.Fatalf("%s expected response buffer", tc.name)
			}
		})
	}
}

func TestCheaterMarkAndFetchVoteInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	cheaterPayload := protobuf.CS_10994{Type: proto.Uint32(3)}
	data, err := proto.Marshal(&cheaterPayload)
	if err != nil {
		t.Fatalf("marshal cheater payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := CheaterMark(&data, client); err != nil {
		t.Fatalf("cheater mark failed: %v", err)
	}
	var cheaterResponse protobuf.SC_10995
	decodeResponse(t, client, &cheaterResponse)
	if cheaterResponse.GetResult() != 3 {
		t.Fatalf("expected result 3")
	}

	votePayload := protobuf.CS_17203{Type: proto.Uint32(1)}
	data, err = proto.Marshal(&votePayload)
	if err != nil {
		t.Fatalf("marshal vote payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := FetchVoteInfo(&data, client); err != nil {
		t.Fatalf("fetch vote info failed: %v", err)
	}
	var voteResponse protobuf.SC_17204
	decodeResponse(t, client, &voteResponse)
}

func TestFetchVoteTicketInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_17201{Type: proto.Uint32(1)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal vote ticket payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := FetchVoteTicketInfo(&data, client); err != nil {
		t.Fatalf("fetch vote ticket info failed: %v", err)
	}
	var response protobuf.SC_17202
	decodeResponse(t, client, &response)

	if response.DailyVote == nil || response.LoveVote == nil {
		t.Fatalf("expected required vote fields to be set")
	}
	if response.GetDailyVote() != 0 {
		t.Fatalf("expected daily vote 0")
	}
	if response.GetLoveVote() != 0 {
		t.Fatalf("expected love vote 0")
	}
	if len(response.GetDailyShipList()) != 0 {
		t.Fatalf("expected empty daily ship list")
	}
}

func TestVersionCheck(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_10996{State: proto.Uint32(1), Platform: proto.String("1")}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal version check payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := VersionCheck(&data, client); err != nil {
		t.Fatalf("version check failed: %v", err)
	}
	var response protobuf.SC_10997
	decodeResponse(t, client, &response)
	regionName := region.Current()
	versionString, err := misc.ResolveRegionVersion(regionName)
	if err != nil {
		t.Fatalf("resolve version: %v", err)
	}
	expectedParts := parseVersionForTest(t, versionString)
	if response.GetVersion1() != expectedParts[0] || response.GetVersion2() != expectedParts[1] || response.GetVersion3() != expectedParts[2] || response.GetVersion4() != expectedParts[3] {
		t.Fatalf("unexpected version parts: %d.%d.%d.%d", response.GetVersion1(), response.GetVersion2(), response.GetVersion3(), response.GetVersion4())
	}
	if response.GetGatewayIp() != consts.RegionGateways[regionName] {
		t.Fatalf("expected gateway ip %q", consts.RegionGateways[regionName])
	}
	if response.GetGatewayPort() != 80 {
		t.Fatalf("expected gateway port 80")
	}
	if response.GetUrl() != consts.GamePlatformUrl[regionName]["1"] {
		t.Fatalf("expected url %q", consts.GamePlatformUrl[regionName]["1"])
	}
}

func TestVersionCheckUnknownPlatform(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_10996{State: proto.Uint32(1), Platform: proto.String("9")}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal version check payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := VersionCheck(&data, client); err == nil {
		t.Fatalf("expected error for unknown platform")
	}
}

func TestVersionCheckMissingVersion(t *testing.T) {
	t.Setenv("AL_REGION", "ZZ")
	region.ResetCurrentForTest()
	t.Cleanup(region.ResetCurrentForTest)
	client := setupHandlerCommander(t)
	payload := protobuf.CS_10996{State: proto.Uint32(1), Platform: proto.String("1")}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal version check payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := VersionCheck(&data, client); err == nil {
		t.Fatalf("expected error for missing version")
	}
}

func TestCommanderHomeAndPlayerExist(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_25026{Type: proto.Uint32(1)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal commander home payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := GetCommanderHome(&data, client); err != nil {
		t.Fatalf("get commander home failed: %v", err)
	}
	var homeResponse protobuf.SC_25027
	decodeResponse(t, client, &homeResponse)
	if homeResponse.GetLevel() != 1 {
		t.Fatalf("expected level 1")
	}

	existPayload := protobuf.CS_10026{AccountId: proto.Uint32(client.Commander.AccountID)}
	data, err = proto.Marshal(&existPayload)
	if err != nil {
		t.Fatalf("marshal player exist payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := PlayerExist(&data, client); err != nil {
		t.Fatalf("player exist failed: %v", err)
	}
	var existResponse protobuf.SC_10027
	decodeResponse(t, client, &existResponse)
	if existResponse.GetUserId() == 0 {
		t.Fatalf("expected user id")
	}
}

func TestLastLoginAndOnlineInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	buffer := []byte{}
	client.Buffer.Reset()
	if _, _, err := LastLogin(&buffer, client); err != nil {
		t.Fatalf("last login failed: %v", err)
	}
	var loginResponse protobuf.SC_11000
	decodeResponse(t, client, &loginResponse)
	if loginResponse.GetTimestamp() == 0 {
		t.Fatalf("expected timestamp")
	}

	buffer = []byte{}
	client.Buffer.Reset()
	if _, _, err := LastOnlineInfo(&buffer, client); err != nil {
		t.Fatalf("last online info failed: %v", err)
	}
	var onlineResponse protobuf.SC_11752
	decodeResponse(t, client, &onlineResponse)
	if onlineResponse.GetActive() != 0 {
		t.Fatalf("expected active 0")
	}
}

func TestChatRoomChange(t *testing.T) {
	client := setupHandlerCommander(t)
	client.Commander.RoomID = 1
	client.Server.JoinRoom(1, client)
	payload := protobuf.CS_11401{RoomId: proto.Uint32(2)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal chat room payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ChatRoomChange(&data, client); err != nil {
		t.Fatalf("chat room change failed: %v", err)
	}
	var response protobuf.SC_11402
	decodeResponse(t, client, &response)
	if response.GetRoomId() != 2 {
		t.Fatalf("expected room id 2")
	}
}

func TestGameNotices(t *testing.T) {
	client := setupHandlerCommander(t)
	notice := orm.Notice{
		Version:    "1",
		BtnTitle:   "Button",
		Title:      "Notice",
		TitleImage: "",
		TimeDesc:   "",
		Content:    "Hello",
		TagType:    1,
		Icon:       1,
		Track:      "",
	}
	if err := orm.GormDB.Create(&notice).Error; err != nil {
		t.Fatalf("seed notice: %v", err)
	}
	buffer := []byte{}
	client.Buffer.Reset()
	if _, _, err := GameNotices(&buffer, client); err != nil {
		t.Fatalf("game notices failed: %v", err)
	}
	var response protobuf.SC_11300
	decodeResponse(t, client, &response)
	if len(response.GetNoticeList()) != 1 {
		t.Fatalf("expected notice list")
	}
}

func TestOwnedItemsAndGiveResources(t *testing.T) {
	client := setupHandlerCommander(t)
	seedHandlerCommanderItem(t, client, 1001, 5)
	if err := orm.GormDB.Create(&orm.CommanderMiscItem{CommanderID: client.Commander.CommanderID, ItemID: 2002, Data: 7}).Error; err != nil {
		t.Fatalf("seed misc item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	buffer := []byte{}
	client.Buffer.Reset()
	if _, _, err := OwnedItems(&buffer, client); err != nil {
		t.Fatalf("owned items failed: %v", err)
	}
	var ownedResponse protobuf.SC_15001
	decodeResponse(t, client, &ownedResponse)
	if len(ownedResponse.GetItemList()) != 1 || len(ownedResponse.GetItemMiscList()) != 1 {
		t.Fatalf("expected item lists")
	}

	seedHandlerCommanderResource(t, client, 7, 4)
	givePayload := protobuf.CS_11013{Type: proto.Uint32(1), Number: proto.Uint32(1)}
	data, err := proto.Marshal(&givePayload)
	if err != nil {
		t.Fatalf("marshal give resources payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := GiveResources(&data, client); err != nil {
		t.Fatalf("give resources failed: %v", err)
	}
	var giveResponse protobuf.SC_11014
	decodeResponse(t, client, &giveResponse)
	if give_attach := giveResponse.GetResult(); give_attach != 0 && give_attach != 1 {
		t.Fatalf("unexpected result %d", give_attach)
	}
}

func TestGiveItem(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_11202{ActivityId: proto.Uint32(42), Cmd: proto.Uint32(1)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal give item payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := GiveItem(&data, client); err != nil {
		t.Fatalf("give item failed: %v", err)
	}
	var response protobuf.SC_11203
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestAskMailBodyAndDeleteArchivedMail(t *testing.T) {
	client := setupHandlerCommander(t)
	mail := orm.Mail{ReceiverID: client.Commander.CommanderID, Title: "Mail", Body: "Body", IsArchived: true}
	if err := orm.GormDB.Create(&mail).Error; err != nil {
		t.Fatalf("seed mail: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	payload := protobuf.CS_30008{MailId: proto.Uint32(mail.ID)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal ask mail payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := AskMailBody(&data, client); err != nil {
		t.Fatalf("ask mail body failed: %v", err)
	}
	var bodyResponse protobuf.SC_30009
	decodeResponse(t, client, &bodyResponse)
	if bodyResponse.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	deletePayload := protobuf.CS_30008{MailId: proto.Uint32(mail.ID)}
	data, err = proto.Marshal(&deletePayload)
	if err != nil {
		t.Fatalf("marshal delete mail payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := DeleteArchivedMail(&data, client); err != nil {
		t.Fatalf("delete archived mail failed: %v", err)
	}
	var deleteResponse protobuf.SC_30009
	decodeResponse(t, client, &deleteResponse)
	if deleteResponse.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestChangeSelectedSkinAndFavorite(t *testing.T) {
	client := setupHandlerCommander(t)
	seedShipTemplate(t, 1001, 1, 2, 1, "Test Ship", 1)
	owned := seedOwnedShip(t, client, 1001)
	if err := orm.GormDB.Create(&orm.OwnedSkin{CommanderID: client.Commander.CommanderID, SkinID: 5001}).Error; err != nil {
		t.Fatalf("seed skin: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	payload := protobuf.CS_12202{ShipId: proto.Uint32(owned.ID), SkinId: proto.Uint32(5001), SkinShadow: proto.Uint32(0)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal change skin payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ChangeSelectedSkin(&data, client); err != nil {
		t.Fatalf("change selected skin failed: %v", err)
	}
	var skinResponse protobuf.SC_12203
	decodeResponse(t, client, &skinResponse)
	if skinResponse.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	favPayload := protobuf.CS_12040{ShipId: proto.Uint32(owned.ID), Flag: proto.Uint32(1)}
	data, err = proto.Marshal(&favPayload)
	if err != nil {
		t.Fatalf("marshal favorite payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := SetFavoriteShip(&data, client); err != nil {
		t.Fatalf("set favorite ship failed: %v", err)
	}
	var favResponse protobuf.SC_12041
	decodeResponse(t, client, &favResponse)
	if favResponse.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}
