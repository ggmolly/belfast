package answer_test

import (
	"os"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupEquipCodeLikeTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearEquipTable(t, &orm.EquipCodeLike{})
	clearEquipTable(t, &orm.Commander{})
	commander := orm.Commander{CommanderID: 188, AccountID: 188, Name: "Equip Code Liker"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestEquipCodeLikeSuccessCreatesLike(t *testing.T) {
	client := setupEquipCodeLikeTest(t)
	day := uint32(time.Now().UTC().Unix() / 86400)

	payload := protobuf.CS_17605{Shipgroup: proto.Uint32(100), Shareid: proto.Uint32(200)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeLike(&buf, client); err != nil {
		t.Fatalf("EquipCodeLike failed: %v", err)
	}
	response := &protobuf.SC_17606{}
	decodeTestPacket(t, client, 17606, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var count int64
	if err := orm.GormDB.Model(&orm.EquipCodeLike{}).
		Where("commander_id = ? AND share_id = ? AND like_day = ?", client.Commander.CommanderID, 200, day).
		Count(&count).Error; err != nil {
		t.Fatalf("count likes: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 like row, got %d", count)
	}
}

func TestEquipCodeLikeDuplicateSameDayReturnsLimited(t *testing.T) {
	client := setupEquipCodeLikeTest(t)

	payload := protobuf.CS_17605{Shipgroup: proto.Uint32(100), Shareid: proto.Uint32(200)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeLike(&buf, client); err != nil {
		t.Fatalf("EquipCodeLike first failed: %v", err)
	}
	first := &protobuf.SC_17606{}
	decodeTestPacket(t, client, 17606, first)
	if first.GetResult() != 0 {
		t.Fatalf("expected first result 0, got %d", first.GetResult())
	}

	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeLike(&buf, client); err != nil {
		t.Fatalf("EquipCodeLike second failed: %v", err)
	}
	second := &protobuf.SC_17606{}
	decodeTestPacket(t, client, 17606, second)
	if second.GetResult() != 7 {
		t.Fatalf("expected second result 7, got %d", second.GetResult())
	}

	var count int64
	if err := orm.GormDB.Model(&orm.EquipCodeLike{}).Count(&count).Error; err != nil {
		t.Fatalf("count likes: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 like row after duplicate, got %d", count)
	}
}

func TestEquipCodeLikeSameShareIDDifferentShipGroupAllowed(t *testing.T) {
	client := setupEquipCodeLikeTest(t)
	day := uint32(time.Now().UTC().Unix() / 86400)

	firstPayload := protobuf.CS_17605{Shipgroup: proto.Uint32(100), Shareid: proto.Uint32(200)}
	firstBuf, err := proto.Marshal(&firstPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	secondPayload := protobuf.CS_17605{Shipgroup: proto.Uint32(101), Shareid: proto.Uint32(200)}
	secondBuf, err := proto.Marshal(&secondPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeLike(&firstBuf, client); err != nil {
		t.Fatalf("EquipCodeLike first failed: %v", err)
	}
	first := &protobuf.SC_17606{}
	decodeTestPacket(t, client, 17606, first)
	if first.GetResult() != 0 {
		t.Fatalf("expected first result 0, got %d", first.GetResult())
	}

	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeLike(&secondBuf, client); err != nil {
		t.Fatalf("EquipCodeLike second failed: %v", err)
	}
	second := &protobuf.SC_17606{}
	decodeTestPacket(t, client, 17606, second)
	if second.GetResult() != 0 {
		t.Fatalf("expected second result 0, got %d", second.GetResult())
	}

	var count int64
	if err := orm.GormDB.Model(&orm.EquipCodeLike{}).
		Where("commander_id = ? AND share_id = ? AND like_day = ?", client.Commander.CommanderID, 200, day).
		Count(&count).Error; err != nil {
		t.Fatalf("count likes: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 like rows for same share/day, got %d", count)
	}

	for _, shipGroupID := range []uint32{100, 101} {
		var perShipGroup int64
		if err := orm.GormDB.Model(&orm.EquipCodeLike{}).
			Where("commander_id = ? AND ship_group_id = ? AND share_id = ? AND like_day = ?", client.Commander.CommanderID, shipGroupID, 200, day).
			Count(&perShipGroup).Error; err != nil {
			t.Fatalf("count likes for shipgroup %d: %v", shipGroupID, err)
		}
		if perShipGroup != 1 {
			t.Fatalf("expected 1 like row for shipgroup %d, got %d", shipGroupID, perShipGroup)
		}
	}
}

func TestEquipCodeLikeInvalidInputReturnsFailure(t *testing.T) {
	client := setupEquipCodeLikeTest(t)

	cases := []protobuf.CS_17605{
		{Shipgroup: proto.Uint32(0), Shareid: proto.Uint32(1)},
		{Shipgroup: proto.Uint32(1), Shareid: proto.Uint32(0)},
	}
	for _, payload := range cases {
		buf, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		client.Buffer.Reset()
		if _, _, err := answer.EquipCodeLike(&buf, client); err != nil {
			t.Fatalf("EquipCodeLike failed: %v", err)
		}
		response := &protobuf.SC_17606{}
		decodeTestPacket(t, client, 17606, response)
		if response.GetResult() != 1 {
			t.Fatalf("expected result 1, got %d", response.GetResult())
		}
	}

	var count int64
	if err := orm.GormDB.Model(&orm.EquipCodeLike{}).Count(&count).Error; err != nil {
		t.Fatalf("count likes: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no like rows after invalid input, got %d", count)
	}
}
