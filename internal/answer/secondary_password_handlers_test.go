package answer

import (
	"reflect"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedSecondaryPasswordSettings(t *testing.T, commanderID uint32, password string, notice string, systemList []uint32, failCount uint32, failCd *int64) {
	t.Helper()
	hash, err := hashSecondaryPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	settings := orm.SecondaryPasswordSettings{
		CommanderID:  commanderID,
		PasswordHash: hash,
		Notice:       notice,
		SystemList:   orm.ToInt64List(systemList),
		FailCount:    failCount,
		FailCd:       failCd,
	}
	if err := orm.UpsertSecondaryPasswordSettings(orm.GormDB, settings); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
}

func TestSetSecondaryPasswordCommandResponseSuccess(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SecondaryPasswordSettings{})

	payload := protobuf.CS_11605{
		Password:   proto.String("123456"),
		Notice:     proto.String("hint"),
		SystemList: []uint32{2, 1},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SetSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_11606
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	stored, err := orm.GetSecondaryPasswordSettings(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if stored.PasswordHash == "" || stored.Notice != "hint" {
		t.Fatalf("expected settings stored")
	}
	if !reflect.DeepEqual(stored.SystemList, orm.Int64List{1, 2}) {
		t.Fatalf("expected sorted system list")
	}
}

func TestSetSecondaryPasswordCommandResponseInvalid(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SecondaryPasswordSettings{})

	payload := protobuf.CS_11605{
		Password:   proto.String("12345"),
		Notice:     proto.String("hint"),
		SystemList: []uint32{1},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SetSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_11606
	decodeResponse(t, client, &response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
}

func TestSetSecondaryPasswordSettingsWrongPassword(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SecondaryPasswordSettings{})

	seedSecondaryPasswordSettings(t, client.Commander.CommanderID, "123456", "hint", []uint32{1, 2}, 0, nil)
	payload := protobuf.CS_11607{
		Password:   proto.String("654321"),
		SystemList: []uint32{1, 2},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SetSecondaryPasswordSettingsCommandResponse(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_11608
	decodeResponse(t, client, &response)
	if response.GetResult() != 9 {
		t.Fatalf("expected result 9, got %d", response.GetResult())
	}
	stored, err := orm.GetSecondaryPasswordSettings(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if stored.FailCount != 1 {
		t.Fatalf("expected fail count 1, got %d", stored.FailCount)
	}
}

func TestSetSecondaryPasswordSettingsLockout(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SecondaryPasswordSettings{})

	lockout := time.Now().Unix() + 120
	seedSecondaryPasswordSettings(t, client.Commander.CommanderID, "123456", "hint", []uint32{1}, 5, &lockout)
	payload := protobuf.CS_11607{
		Password:   proto.String("123456"),
		SystemList: []uint32{1},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SetSecondaryPasswordSettingsCommandResponse(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_11608
	decodeResponse(t, client, &response)
	if response.GetResult() != 40 {
		t.Fatalf("expected result 40, got %d", response.GetResult())
	}
}

func TestConfirmSecondaryPasswordCommandResponseSuccess(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SecondaryPasswordSettings{})

	lockout := time.Now().Unix() - 10
	seedSecondaryPasswordSettings(t, client.Commander.CommanderID, "123456", "hint", []uint32{1}, 4, &lockout)
	payload := protobuf.CS_11609{Password: proto.String("123456")}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ConfirmSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_11610
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	stored, err := orm.GetSecondaryPasswordSettings(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if stored.FailCount != 0 || stored.FailCd != nil {
		t.Fatalf("expected lockout cleared")
	}
}

func TestFetchSecondaryPasswordCommandResponseState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SecondaryPasswordSettings{})

	seedSecondaryPasswordSettings(t, client.Commander.CommanderID, "123456", "hint", []uint32{1, 2}, 2, nil)
	buffer := []byte{}
	if _, _, err := FetchSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_11604
	decodeResponse(t, client, &response)
	if response.GetState() != 1 {
		t.Fatalf("expected state 1, got %d", response.GetState())
	}
	if response.GetNotice() != "hint" {
		t.Fatalf("expected notice hint")
	}
	if response.GetFailCount() != 2 {
		t.Fatalf("expected fail count 2, got %d", response.GetFailCount())
	}
	if !reflect.DeepEqual(response.GetSystemList(), []uint32{1, 2}) {
		t.Fatalf("expected system list [1 2]")
	}
}
