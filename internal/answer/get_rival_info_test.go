package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func decodeSC18105(t *testing.T, client *connection.Client) *protobuf.SC_18105 {
	t.Helper()
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != 18105 {
		t.Fatalf("expected packet 18105, got %d", packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	var response protobuf.SC_18105
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	client.Buffer.Reset()
	return &response
}

func TestGetRivalInfo_NotFound_ReturnsSentinelIdZero(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.Fleet{})

	payload := protobuf.CS_18104{Id: proto.Uint32(999999)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := GetRivalInfo(&data, client); err != nil {
		t.Fatalf("GetRivalInfo failed: %v", err)
	}
	response := decodeSC18105(t, client)
	if response.GetInfo().GetId() != 0 {
		t.Fatalf("expected sentinel info.id == 0, got %d", response.GetInfo().GetId())
	}
}

func TestGetRivalInfo_Success_PopulatesTargetInfoAndDisplay(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.Fleet{})

	seedShipTemplate(t, 1001, 1, 2, 1, "Vanguard-1", 1)
	seedShipTemplate(t, 1002, 1, 2, 1, "Vanguard-2", 1)
	seedShipTemplate(t, 1003, 1, 2, 1, "Vanguard-3", 1)
	seedShipTemplate(t, 2001, 1, 2, 5, "Main-1", 1)
	seedShipTemplate(t, 2002, 1, 2, 5, "Main-2", 1)
	seedShipTemplate(t, 2003, 1, 2, 5, "Main-3", 1)

	targetID := uint32(424242)
	target := orm.Commander{
		CommanderID:         targetID,
		AccountID:           targetID,
		Name:                "Target Commander",
		Level:               77,
		LastLogin:           time.Now().UTC(),
		DisplayIconID:       123,
		DisplaySkinID:       456,
		SelectedIconFrameID: 7,
		SelectedChatFrameID: 8,
		DisplayIconThemeID:  9,
		Manifesto:           "",
		NameChangeCooldown:  time.Unix(0, 0),
	}
	if err := orm.GormDB.Create(&target).Error; err != nil {
		t.Fatalf("create target commander: %v", err)
	}

	ownedTemplates := []uint32{1001, 1002, 1003, 2001, 2002, 2003}
	ownedIDs := make([]uint32, 0, len(ownedTemplates))
	for _, templateID := range ownedTemplates {
		owned := orm.OwnedShip{OwnerID: targetID, ShipID: templateID}
		if err := orm.GormDB.Create(&owned).Error; err != nil {
			t.Fatalf("create owned ship: %v", err)
		}
		ownedIDs = append(ownedIDs, owned.ID)
	}

	if err := target.Load(); err != nil {
		t.Fatalf("load target commander: %v", err)
	}
	if err := orm.CreateFleet(&target, 1, "", ownedIDs); err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	payload := protobuf.CS_18104{Id: proto.Uint32(targetID)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := GetRivalInfo(&data, client); err != nil {
		t.Fatalf("GetRivalInfo failed: %v", err)
	}
	response := decodeSC18105(t, client)
	info := response.GetInfo()
	if info.GetId() != targetID {
		t.Fatalf("expected info.id %d, got %d", targetID, info.GetId())
	}
	if info.GetLevel() != uint32(target.Level) {
		t.Fatalf("expected info.level %d, got %d", target.Level, info.GetLevel())
	}
	if info.GetName() != target.Name {
		t.Fatalf("expected info.name %q, got %q", target.Name, info.GetName())
	}
	if info.GetScore() != 0 || info.GetRank() != 0 {
		t.Fatalf("expected score/rank placeholders 0, got score=%d rank=%d", info.GetScore(), info.GetRank())
	}
	if info.GetDisplay() == nil {
		t.Fatalf("expected display to be populated")
	}
	if got := info.GetDisplay().GetIcon(); got != target.DisplayIconID {
		t.Fatalf("expected display.icon %d, got %d", target.DisplayIconID, got)
	}
	if got := info.GetDisplay().GetSkin(); got != target.DisplaySkinID {
		t.Fatalf("expected display.skin %d, got %d", target.DisplaySkinID, got)
	}
	if got := info.GetDisplay().GetIconFrame(); got != target.SelectedIconFrameID {
		t.Fatalf("expected display.icon_frame %d, got %d", target.SelectedIconFrameID, got)
	}
	if got := info.GetDisplay().GetChatFrame(); got != target.SelectedChatFrameID {
		t.Fatalf("expected display.chat_frame %d, got %d", target.SelectedChatFrameID, got)
	}
	if got := info.GetDisplay().GetIconTheme(); got != target.DisplayIconThemeID {
		t.Fatalf("expected display.icon_theme %d, got %d", target.DisplayIconThemeID, got)
	}

	if len(info.GetVanguardShipList()) == 0 && len(info.GetMainShipList()) == 0 {
		t.Fatalf("expected at least one ship in vanguard or main list")
	}
}

func TestGetRivalInfo_StaleFleetShipIDs_FallsBackToOwnedShips(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.Fleet{})

	seedShipTemplate(t, 4001, 1, 2, 1, "Vanguard", 1)
	seedShipTemplate(t, 4002, 1, 2, 5, "Main", 1)

	targetID := uint32(565656)
	target := orm.Commander{
		CommanderID:        targetID,
		AccountID:          targetID,
		Name:               "Stale Fleet Commander",
		Level:              25,
		LastLogin:          time.Now().UTC(),
		NameChangeCooldown: time.Unix(0, 0),
	}
	if err := orm.GormDB.Create(&target).Error; err != nil {
		t.Fatalf("create target commander: %v", err)
	}

	ownedIDs := make(map[uint32]bool, 2)
	for _, templateID := range []uint32{4001, 4002} {
		owned := orm.OwnedShip{OwnerID: targetID, ShipID: templateID}
		if err := orm.GormDB.Create(&owned).Error; err != nil {
			t.Fatalf("create owned ship: %v", err)
		}
		ownedIDs[owned.ID] = true
	}

	staleA := int64(9999999)
	staleB := int64(8888888)
	fleet := orm.Fleet{
		CommanderID:    targetID,
		GameID:         1,
		Name:           "",
		ShipList:       orm.Int64List{staleA, staleB},
		MeowfficerList: orm.Int64List{},
	}
	if err := orm.GormDB.Create(&fleet).Error; err != nil {
		t.Fatalf("create stale fleet: %v", err)
	}

	payload := protobuf.CS_18104{Id: proto.Uint32(targetID)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := GetRivalInfo(&data, client); err != nil {
		t.Fatalf("GetRivalInfo failed: %v", err)
	}
	response := decodeSC18105(t, client)
	info := response.GetInfo()

	got := append(info.GetVanguardShipList(), info.GetMainShipList()...)
	if len(got) == 0 {
		t.Fatalf("expected ships even when fleet ship ids are stale")
	}
	for _, ship := range got {
		if ship.GetId() == uint32(staleA) || ship.GetId() == uint32(staleB) {
			t.Fatalf("expected stale fleet ids to be ignored, got ship id %d", ship.GetId())
		}
		if !ownedIDs[ship.GetId()] {
			t.Fatalf("expected returned ship id %d to exist in owned ships", ship.GetId())
		}
	}
}

func TestGetRivalInfo_DisplayFallback_UsesFirstSecretary(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.Fleet{})

	seedShipTemplate(t, 3001, 1, 2, 1, "Secretary", 1)

	targetID := uint32(333333)
	target := orm.Commander{
		CommanderID:        targetID,
		AccountID:          targetID,
		Name:               "Fallback Commander",
		Level:              10,
		LastLogin:          time.Now().UTC(),
		DisplayIconID:      0,
		DisplaySkinID:      0,
		NameChangeCooldown: time.Unix(0, 0),
	}
	if err := orm.GormDB.Create(&target).Error; err != nil {
		t.Fatalf("create target commander: %v", err)
	}
	owned := orm.OwnedShip{OwnerID: targetID, ShipID: 3001, SkinID: 888, IsSecretary: true}
	pos := uint32(0)
	owned.SecretaryPosition = &pos
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("create owned secretary ship: %v", err)
	}

	payload := protobuf.CS_18104{Id: proto.Uint32(targetID)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := GetRivalInfo(&data, client); err != nil {
		t.Fatalf("GetRivalInfo failed: %v", err)
	}
	response := decodeSC18105(t, client)
	info := response.GetInfo()
	if info.GetDisplay() == nil {
		t.Fatalf("expected display to be populated")
	}
	if got := info.GetDisplay().GetIcon(); got != 3001 {
		t.Fatalf("expected display.icon fallback to secretary ship_id 3001, got %d", got)
	}
	if got := info.GetDisplay().GetSkin(); got != 888 {
		t.Fatalf("expected display.skin fallback to secretary skin_id 888, got %d", got)
	}
}
