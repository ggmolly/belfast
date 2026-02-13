package answer_test

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
	expiresAt := time.Now().Add(2 * time.Hour).UTC()
	if err := orm.CreateCommanderRoot(1, 1, "Skins Commander", 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 1}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	if err := commander.GiveSkinWithExpiry(9001, &expiresAt); err != nil {
		t.Fatalf("failed to create owned skin: %v", err)
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO global_skin_restrictions (skin_id, type) VALUES ($1, $2)", int64(1001), int64(0))
	execAnswerExternalTestSQLT(t, "INSERT INTO global_skin_restrictions (skin_id, type) VALUES ($1, $2)", int64(2002), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO global_skin_restriction_windows (id, skin_id, type, start_time, stop_time) VALUES ($1, $2, $3, $4, $5)", int64(1), int64(3003), int64(1), int64(100), int64(200))
	execAnswerExternalTestSQLT(t, "INSERT INTO global_skin_restriction_windows (id, skin_id, type, start_time, stop_time) VALUES ($1, $2, $3, $4, $5)", int64(2), int64(4004), int64(2), int64(300), int64(400))

	client := connection.Client{Commander: &commander}
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
