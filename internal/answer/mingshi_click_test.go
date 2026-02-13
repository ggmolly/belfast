package answer_test

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestClickMingShiIncrementsAccPayLv(t *testing.T) {
	commanderID := uint32(9101)
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   commanderID,
		Name:        fmt.Sprintf("MingShi Commander %d", commanderID),
	}
	if err := orm.CreateCommanderRoot(commanderID, commanderID, commander.Name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	defer func() {
		execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	}()
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}

	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_11506{State: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ClickMingShi(&buf, client); err != nil {
		t.Fatalf("ClickMingShi failed: %v", err)
	}
	response := &protobuf.SC_11507{}
	decodeTestPacket(t, client, 11507, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	updated, err := orm.GetCommanderCoreByID(commanderID)
	if err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}
	if updated.AccPayLv != 5 {
		t.Fatalf("expected acc_pay_lv 5, got %d", updated.AccPayLv)
	}
}
