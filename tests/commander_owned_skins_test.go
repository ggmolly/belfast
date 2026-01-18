package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestCommanderOwnedSkinsForbiddenLists(t *testing.T) {
	if err := orm.GormDB.Exec("DELETE FROM global_skin_restrictions").Error; err != nil {
		t.Fatalf("failed to clear restrictions: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM global_skin_restriction_windows").Error; err != nil {
		t.Fatalf("failed to clear restriction windows: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM owned_skins").Error; err != nil {
		t.Fatalf("failed to clear owned skins: %v", err)
	}

	expiresAt := time.Now().Add(2 * time.Hour).UTC()
	ownedSkin := orm.OwnedSkin{
		CommanderID: fakeCommander.CommanderID,
		SkinID:      9001,
		ExpiresAt:   &expiresAt,
	}
	if err := orm.GormDB.Create(&ownedSkin).Error; err != nil {
		t.Fatalf("failed to create owned skin: %v", err)
	}

	restrictions := []orm.GlobalSkinRestriction{
		{SkinID: 1001, Type: 0},
		{SkinID: 2002, Type: 1},
	}
	if err := orm.GormDB.Create(&restrictions).Error; err != nil {
		t.Fatalf("failed to create restrictions: %v", err)
	}

	windows := []orm.GlobalSkinRestrictionWindow{
		{ID: 1, SkinID: 3003, Type: 1, StartTime: 100, StopTime: 200},
		{ID: 2, SkinID: 4004, Type: 2, StartTime: 300, StopTime: 400},
	}
	if err := orm.GormDB.Create(&windows).Error; err != nil {
		t.Fatalf("failed to create restriction windows: %v", err)
	}

	if err := fakeCommander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	client := connection.Client{Commander: &fakeCommander}
	buffer := []byte{}
	if _, _, err := answer.CommanderOwnedSkins(&buffer, &client); err != nil {
		t.Fatalf("failed to build response: %v", err)
	}

	payload := client.Buffer.Bytes()
	if packets.GetPacketId(0, &payload) != 12201 {
		t.Fatalf("expected packet 12201, got %d", packets.GetPacketId(0, &payload))
	}
	payload = payload[packets.HEADER_SIZE:]

	var response protobuf.SC_12201
	if err := proto.Unmarshal(payload, &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.SkinList) != 1 {
		t.Fatalf("expected 1 owned skin, got %d", len(response.SkinList))
	}
	if response.GetSkinList()[0].GetId() != 9001 {
		t.Fatalf("expected skin id 9001, got %d", response.GetSkinList()[0].GetId())
	}
	if response.GetSkinList()[0].GetTime() != uint32(expiresAt.Unix()) {
		t.Fatalf("expected expiry %d, got %d", expiresAt.Unix(), response.GetSkinList()[0].GetTime())
	}

	expectedForbiddenList := []uint32{1001, 2002}
	expectedForbiddenTypes := []uint32{0, 1}
	if !reflect.DeepEqual(expectedForbiddenList, response.ForbiddenSkinList) {
		t.Fatalf("forbidden skin list mismatch: %v", response.ForbiddenSkinList)
	}
	if !reflect.DeepEqual(expectedForbiddenTypes, response.ForbiddenSkinType) {
		t.Fatalf("forbidden skin type mismatch: %v", response.ForbiddenSkinType)
	}

	if len(response.ForbiddenList) != 2 {
		t.Fatalf("expected 2 forbidden windows, got %d", len(response.ForbiddenList))
	}
	first := response.ForbiddenList[0]
	second := response.ForbiddenList[1]
	if first.GetId() != 3003 || first.GetType() != 1 || first.GetStartTime() != 100 || first.GetStopTime() != 200 {
		t.Fatalf("unexpected first forbidden window: %+v", first)
	}
	if second.GetId() != 4004 || second.GetType() != 2 || second.GetStartTime() != 300 || second.GetStopTime() != 400 {
		t.Fatalf("unexpected second forbidden window: %+v", second)
	}
}
