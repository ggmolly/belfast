package answer_test

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupEquipCodeShareTest(t *testing.T) *connection.Client {
	t.Helper()
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	execAnswerExternalTestSQLT(t, "DELETE FROM equip_code_shares")
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders")
	commander := orm.Commander{CommanderID: 199, AccountID: 199, Name: "Equip Code Sharer"}
	if err := orm.CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func encodeConversionBase32(n uint32) string {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUV"
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 8)
	for n > 0 {
		d := n % 32
		buf = append(buf, alphabet[d])
		n /= 32
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func makeEqCode(shipGroupID uint32) string {
	return "build&" + encodeConversionBase32(shipGroupID) + "&tag1&tag2"
}

func TestEquipCodeShareSuccessStoresShare(t *testing.T) {
	client := setupEquipCodeShareTest(t)
	day := uint32(time.Now().UTC().Unix() / 86400)

	payload := protobuf.CS_17603{Shipgroup: proto.Uint32(100), Eqcode: proto.String(makeEqCode(100))}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeShare(&buf, client); err != nil {
		t.Fatalf("EquipCodeShare failed: %v", err)
	}
	response := &protobuf.SC_17604{}
	decodeTestPacket(t, client, 17604, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var count int64
	count = queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM equip_code_shares WHERE commander_id = $1 AND ship_group_id = $2 AND share_day = $3", int64(client.Commander.CommanderID), int64(100), int64(day))
	if count != 1 {
		t.Fatalf("expected 1 share row, got %d", count)
	}
}

func TestEquipCodeShareInvalidInputReturnsFailure(t *testing.T) {
	client := setupEquipCodeShareTest(t)

	cases := []protobuf.CS_17603{
		{Shipgroup: proto.Uint32(0), Eqcode: proto.String(makeEqCode(0))},
		{Shipgroup: proto.Uint32(1), Eqcode: proto.String("")},
		{Shipgroup: proto.Uint32(10), Eqcode: proto.String("bad&10&tag1")},
		{Shipgroup: proto.Uint32(10), Eqcode: proto.String(makeEqCode(11))},
	}
	for _, payload := range cases {
		buf, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		client.Buffer.Reset()
		if _, _, err := answer.EquipCodeShare(&buf, client); err != nil {
			t.Fatalf("EquipCodeShare failed: %v", err)
		}
		response := &protobuf.SC_17604{}
		decodeTestPacket(t, client, 17604, response)
		if response.GetResult() != 1 {
			t.Fatalf("expected result 1, got %d", response.GetResult())
		}
	}

	var count int64
	count = queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM equip_code_shares")
	if count != 0 {
		t.Fatalf("expected no share rows after invalid input, got %d", count)
	}
}

func TestEquipCodeShareDuplicateSameDayReturnsAlreadyShared(t *testing.T) {
	client := setupEquipCodeShareTest(t)

	payload := protobuf.CS_17603{Shipgroup: proto.Uint32(100), Eqcode: proto.String(makeEqCode(100))}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeShare(&buf, client); err != nil {
		t.Fatalf("EquipCodeShare first failed: %v", err)
	}
	first := &protobuf.SC_17604{}
	decodeTestPacket(t, client, 17604, first)
	if first.GetResult() != 0 {
		t.Fatalf("expected first result 0, got %d", first.GetResult())
	}

	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeShare(&buf, client); err != nil {
		t.Fatalf("EquipCodeShare second failed: %v", err)
	}
	second := &protobuf.SC_17604{}
	decodeTestPacket(t, client, 17604, second)
	if second.GetResult() != 7 {
		t.Fatalf("expected second result 7, got %d", second.GetResult())
	}

	var count int64
	count = queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM equip_code_shares")
	if count != 1 {
		t.Fatalf("expected 1 share row after duplicate, got %d", count)
	}
}

func TestEquipCodeShareGlobalDailyLimitReturnsLimited(t *testing.T) {
	client := setupEquipCodeShareTest(t)
	t.Setenv("EQUIP_CODE_SHARE_DAILY_LIMIT", "2")

	for _, shipGroupID := range []uint32{100, 101} {
		payload := protobuf.CS_17603{Shipgroup: proto.Uint32(shipGroupID), Eqcode: proto.String(makeEqCode(shipGroupID))}
		buf, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		client.Buffer.Reset()
		if _, _, err := answer.EquipCodeShare(&buf, client); err != nil {
			t.Fatalf("EquipCodeShare failed: %v", err)
		}
		response := &protobuf.SC_17604{}
		decodeTestPacket(t, client, 17604, response)
		if response.GetResult() != 0 {
			t.Fatalf("expected result 0 for shipgroup %d, got %d", shipGroupID, response.GetResult())
		}
	}

	thirdPayload := protobuf.CS_17603{Shipgroup: proto.Uint32(102), Eqcode: proto.String(makeEqCode(102))}
	thirdBuf, err := proto.Marshal(&thirdPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeShare(&thirdBuf, client); err != nil {
		t.Fatalf("EquipCodeShare third failed: %v", err)
	}
	third := &protobuf.SC_17604{}
	decodeTestPacket(t, client, 17604, third)
	if third.GetResult() != 44 {
		t.Fatalf("expected result 44, got %d", third.GetResult())
	}

	var count int64
	count = queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM equip_code_shares")
	if count != 2 {
		t.Fatalf("expected 2 share rows after limit, got %d", count)
	}
}
