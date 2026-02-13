package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupEquipCodeImpeachTest(t *testing.T) *connection.Client {
	t.Helper()
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	execAnswerExternalTestSQLT(t, "DELETE FROM equip_code_reports")
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders")
	commander := orm.Commander{CommanderID: 177, AccountID: 177, Name: "Equip Code Reporter"}
	if err := orm.CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestEquipCodeImpeachSuccessCreatesReport(t *testing.T) {
	client := setupEquipCodeImpeachTest(t)
	t.Setenv("EQUIP_CODE_IMPEACH_DAILY_LIMIT", "")

	payload := protobuf.CS_17607{
		Shipgroup:  proto.Uint32(100),
		Shareid:    proto.Uint32(200),
		ReportType: proto.Uint32(1),
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeImpeach(&buf, client); err != nil {
		t.Fatalf("EquipCodeImpeach failed: %v", err)
	}
	response := &protobuf.SC_17608{}
	decodeTestPacket(t, client, 17608, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	storedShipGroupID := queryAnswerExternalTestInt64(t, "SELECT ship_group_id FROM equip_code_reports WHERE commander_id = $1 AND share_id = $2", int64(client.Commander.CommanderID), int64(200))
	if storedShipGroupID != 100 {
		t.Fatalf("expected shipgroup 100, got %d", storedShipGroupID)
	}
	storedReportType := queryAnswerExternalTestInt64(t, "SELECT report_type FROM equip_code_reports WHERE commander_id = $1 AND share_id = $2", int64(client.Commander.CommanderID), int64(200))
	if storedReportType != 1 {
		t.Fatalf("expected report_type 1, got %d", storedReportType)
	}
}

func TestEquipCodeImpeachReportTypeValidation(t *testing.T) {
	client := setupEquipCodeImpeachTest(t)

	payload := protobuf.CS_17607{
		Shipgroup:  proto.Uint32(1),
		Shareid:    proto.Uint32(2),
		ReportType: proto.Uint32(3),
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeImpeach(&buf, client); err != nil {
		t.Fatalf("EquipCodeImpeach failed: %v", err)
	}
	response := &protobuf.SC_17608{}
	decodeTestPacket(t, client, 17608, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var count int64
	count = queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM equip_code_reports")
	if count != 0 {
		t.Fatalf("expected no report rows, got %d", count)
	}
}

func TestEquipCodeImpeachWarningEncoding(t *testing.T) {
	client := setupEquipCodeImpeachTest(t)
	t.Setenv("EQUIP_CODE_IMPEACH_DAILY_LIMIT", "2")

	for shareID := uint32(1); shareID <= 2; shareID++ {
		payload := protobuf.CS_17607{Shipgroup: proto.Uint32(1), Shareid: proto.Uint32(shareID), ReportType: proto.Uint32(1)}
		buf, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("marshal payload %d: %v", shareID, err)
		}
		client.Buffer.Reset()
		if _, _, err := answer.EquipCodeImpeach(&buf, client); err != nil {
			t.Fatalf("EquipCodeImpeach %d failed: %v", shareID, err)
		}
		response := &protobuf.SC_17608{}
		decodeTestPacket(t, client, 17608, response)
		if response.GetResult() != 0 {
			t.Fatalf("expected result 0 for share %d, got %d", shareID, response.GetResult())
		}
	}

	payload := protobuf.CS_17607{Shipgroup: proto.Uint32(1), Shareid: proto.Uint32(3), ReportType: proto.Uint32(1)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.EquipCodeImpeach(&buf, client); err != nil {
		t.Fatalf("EquipCodeImpeach failed: %v", err)
	}
	response := &protobuf.SC_17608{}
	decodeTestPacket(t, client, 17608, response)
	if response.GetResult() != ^uint32(0) {
		t.Fatalf("expected warning result, got %d", response.GetResult())
	}
}
