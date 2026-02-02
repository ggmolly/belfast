package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestSecondaryPasswordFetchDefaults(t *testing.T) {
	client := setupHandlerCommander(t)
	buffer := []byte{}
	if _, _, err := FetchSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("fetch secondary password failed: %v", err)
	}
	var response protobuf.SC_11604
	decodeResponse(t, client, &response)
	if response.GetState() != 0 {
		t.Fatalf("expected state 0, got %d", response.GetState())
	}
	if response.GetFailCount() != 0 {
		t.Fatalf("expected fail_count 0, got %d", response.GetFailCount())
	}
	if response.GetFailCd() != 0 {
		t.Fatalf("expected fail_cd 0, got %d", response.GetFailCd())
	}
	if response.GetNotice() != "" {
		t.Fatalf("expected empty notice, got %q", response.GetNotice())
	}
	if len(response.GetSystemList()) != 0 {
		t.Fatalf("expected empty system_list, got %v", response.GetSystemList())
	}
}

func TestSecondaryPasswordSetAndConfirm(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_11605{
		Password:   proto.String("123456"),
		Notice:     proto.String("test note"),
		SystemList: []uint32{3, 1, 3},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := SetSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("set secondary password failed: %v", err)
	}
	var setResponse protobuf.SC_11606
	decodeResponse(t, client, &setResponse)
	if setResponse.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", setResponse.GetResult())
	}
	state, err := orm.GetOrCreateSecondaryPasswordState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load secondary password state: %v", err)
	}
	if state.State != 1 {
		t.Fatalf("expected state 1, got %d", state.State)
	}
	if state.Notice != "test note" {
		t.Fatalf("expected notice to be stored")
	}
	if got := orm.ToUint32List(state.SystemList); len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("expected system_list [1 3], got %v", got)
	}

	confirmPayload := protobuf.CS_11609{Password: proto.String("123456")}
	confirmBuffer, err := proto.Marshal(&confirmPayload)
	if err != nil {
		t.Fatalf("marshal confirm payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ConfirmSecondaryPasswordCommandResponse(&confirmBuffer, client); err != nil {
		t.Fatalf("confirm secondary password failed: %v", err)
	}
	var confirmResponse protobuf.SC_11610
	decodeResponse(t, client, &confirmResponse)
	if confirmResponse.GetResult() != 0 {
		t.Fatalf("expected confirm result 0, got %d", confirmResponse.GetResult())
	}
	state, err = orm.GetOrCreateSecondaryPasswordState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load secondary password state: %v", err)
	}
	if state.State != 2 {
		t.Fatalf("expected state 2, got %d", state.State)
	}
}

func TestSecondaryPasswordWrongPasswordLockout(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_11605{
		Password:   proto.String("123456"),
		Notice:     proto.String(""),
		SystemList: []uint32{1},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SetSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("set secondary password failed: %v", err)
	}

	wrongPayload := protobuf.CS_11609{Password: proto.String("000000")}
	wrongBuffer, err := proto.Marshal(&wrongPayload)
	if err != nil {
		t.Fatalf("marshal wrong payload: %v", err)
	}
	for i := 0; i < secondaryPasswordMaxFailures; i++ {
		client.Buffer.Reset()
		if _, _, err := ConfirmSecondaryPasswordCommandResponse(&wrongBuffer, client); err != nil {
			t.Fatalf("confirm secondary password failed: %v", err)
		}
		var response protobuf.SC_11610
		decodeResponse(t, client, &response)
		if response.GetResult() != 9 {
			t.Fatalf("expected result 9, got %d", response.GetResult())
		}
	}

	state, err := orm.GetOrCreateSecondaryPasswordState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load secondary password state: %v", err)
	}
	if state.FailCount != secondaryPasswordMaxFailures {
		t.Fatalf("expected fail_count %d, got %d", secondaryPasswordMaxFailures, state.FailCount)
	}
	now := uint32(time.Now().Unix())
	if state.FailCd <= now {
		t.Fatalf("expected fail_cd in the future, got %d", state.FailCd)
	}

	confirmPayload := protobuf.CS_11609{Password: proto.String("123456")}
	confirmBuffer, err := proto.Marshal(&confirmPayload)
	if err != nil {
		t.Fatalf("marshal confirm payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ConfirmSecondaryPasswordCommandResponse(&confirmBuffer, client); err != nil {
		t.Fatalf("confirm secondary password failed: %v", err)
	}
	var lockedResponse protobuf.SC_11610
	decodeResponse(t, client, &lockedResponse)
	if lockedResponse.GetResult() != 1 {
		t.Fatalf("expected result 1 for locked state, got %d", lockedResponse.GetResult())
	}
}

func TestSecondaryPasswordSettingsUpdate(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_11605{
		Password:   proto.String("123456"),
		Notice:     proto.String(""),
		SystemList: []uint32{1},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SetSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("set secondary password failed: %v", err)
	}

	settingsPayload := protobuf.CS_11607{
		Password:   proto.String("123456"),
		SystemList: []uint32{2, 1},
	}
	settingsBuffer, err := proto.Marshal(&settingsPayload)
	if err != nil {
		t.Fatalf("marshal settings payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := SetSecondaryPasswordSettingsCommandResponse(&settingsBuffer, client); err != nil {
		t.Fatalf("set secondary password settings failed: %v", err)
	}
	var response protobuf.SC_11608
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	state, err := orm.GetOrCreateSecondaryPasswordState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load secondary password state: %v", err)
	}
	if state.State != 2 {
		t.Fatalf("expected state 2, got %d", state.State)
	}
	if got := orm.ToUint32List(state.SystemList); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("expected system_list [1 2], got %v", got)
	}
}

func TestSecondaryPasswordSettingsDisableClearsHash(t *testing.T) {
	client := setupHandlerCommander(t)
	payload := protobuf.CS_11605{
		Password:   proto.String("123456"),
		Notice:     proto.String("notice"),
		SystemList: []uint32{1},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SetSecondaryPasswordCommandResponse(&buffer, client); err != nil {
		t.Fatalf("set secondary password failed: %v", err)
	}

	settingsPayload := protobuf.CS_11607{
		Password:   proto.String("123456"),
		SystemList: []uint32{},
	}
	settingsBuffer, err := proto.Marshal(&settingsPayload)
	if err != nil {
		t.Fatalf("marshal settings payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := SetSecondaryPasswordSettingsCommandResponse(&settingsBuffer, client); err != nil {
		t.Fatalf("set secondary password settings failed: %v", err)
	}
	var response protobuf.SC_11608
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	state, err := orm.GetOrCreateSecondaryPasswordState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load secondary password state: %v", err)
	}
	if state.State != 0 {
		t.Fatalf("expected state 0, got %d", state.State)
	}
	if state.PasswordHash != "" {
		t.Fatalf("expected empty password hash")
	}
	if state.Notice != "" {
		t.Fatalf("expected empty notice")
	}
	if len(state.SystemList) != 0 {
		t.Fatalf("expected empty system list")
	}

	reenablePayload := protobuf.CS_11605{
		Password:   proto.String("654321"),
		Notice:     proto.String(""),
		SystemList: []uint32{2},
	}
	reenableBuffer, err := proto.Marshal(&reenablePayload)
	if err != nil {
		t.Fatalf("marshal reenable payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := SetSecondaryPasswordCommandResponse(&reenableBuffer, client); err != nil {
		t.Fatalf("reenable secondary password failed: %v", err)
	}
	var reenableResponse protobuf.SC_11606
	decodeResponse(t, client, &reenableResponse)
	if reenableResponse.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", reenableResponse.GetResult())
	}
}
