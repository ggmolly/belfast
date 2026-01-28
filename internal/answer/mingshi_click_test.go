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
	if err := orm.GormDB.Unscoped().Delete(&orm.Commander{}, commanderID).Error; err != nil {
		t.Fatalf("failed to cleanup commander: %v", err)
	}
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   commanderID,
		Name:        fmt.Sprintf("MingShi Commander %d", commanderID),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	defer func() {
		if err := orm.GormDB.Unscoped().Delete(&orm.Commander{}, commanderID).Error; err != nil {
			t.Fatalf("failed to cleanup commander: %v", err)
		}
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
	var updated orm.Commander
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&updated).Error; err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}
	if updated.AccPayLv != 5 {
		t.Fatalf("expected acc_pay_lv 5, got %d", updated.AccPayLv)
	}
}
