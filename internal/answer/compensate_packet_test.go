package answer_test

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestCompensateNotificationSummary(t *testing.T) {
	commander := orm.Commander{
		CommanderID: 50,
		AccountID:   50,
		Name:        "Comp Summary",
		LastLogin:   time.Now(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}

	now := time.Now()
	comp1 := orm.Compensation{
		CommanderID: commander.CommanderID,
		Title:       "One",
		Text:        "First",
		SendTime:    now,
		ExpiresAt:   now.Add(2 * time.Hour),
	}
	comp2 := orm.Compensation{
		CommanderID: commander.CommanderID,
		Title:       "Two",
		Text:        "Second",
		SendTime:    now,
		ExpiresAt:   now.Add(3 * time.Hour),
		AttachFlag:  true,
	}
	comp3 := orm.Compensation{
		CommanderID: commander.CommanderID,
		Title:       "Three",
		Text:        "Third",
		SendTime:    now,
		ExpiresAt:   now.Add(-2 * time.Hour),
	}
	if err := orm.GormDB.Create(&comp1).Error; err != nil {
		t.Fatalf("failed to create compensation 1: %v", err)
	}
	if err := orm.GormDB.Create(&comp2).Error; err != nil {
		t.Fatalf("failed to create compensation 2: %v", err)
	}
	if err := orm.GormDB.Create(&comp3).Error; err != nil {
		t.Fatalf("failed to create compensation 3: %v", err)
	}

	commander.Compensations = []orm.Compensation{comp1, comp2, comp3}
	client := &connection.Client{Commander: &commander}
	buffer := []byte{}
	if _, _, err := answer.CompensateNotification(&buffer, client); err != nil {
		t.Fatalf("CompensateNotification failed: %v", err)
	}

	response := &protobuf.SC_30101{}
	decodeTestPacket(t, client, 30101, response)
	if response.GetNumber() != 1 {
		t.Fatalf("expected number 1, got %d", response.GetNumber())
	}
	expectedMax := uint32(comp2.ExpiresAt.Unix())
	if response.GetMaxTimestamp() != expectedMax {
		t.Fatalf("expected max_timestamp %d, got %d", expectedMax, response.GetMaxTimestamp())
	}
}

func TestCompensateListAndClaim(t *testing.T) {
	commander := orm.Commander{
		CommanderID: 51,
		AccountID:   51,
		Name:        "Comp Claim",
		LastLogin:   time.Now(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)
	commander.MiscItemsMap = make(map[uint32]*orm.CommanderMiscItem)
	commander.OwnedResourcesMap = make(map[uint32]*orm.OwnedResource)
	commander.OwnedSkinsMap = make(map[uint32]*orm.OwnedSkin)
	commander.OwnedShipsMap = make(map[uint32]*orm.OwnedShip)

	now := time.Now()
	comp := orm.Compensation{
		CommanderID: commander.CommanderID,
		Title:       "Reward",
		Text:        "Claim",
		SendTime:    now,
		ExpiresAt:   now.Add(2 * time.Hour),
		Attachments: []orm.CompensationAttachment{{
			Type:     consts.DROP_TYPE_ITEM,
			ItemID:   20001,
			Quantity: 1,
		}},
	}
	if err := orm.GormDB.Create(&comp).Error; err != nil {
		t.Fatalf("failed to create compensation: %v", err)
	}

	commander.Compensations = []orm.Compensation{comp}
	commander.CompensationsMap = map[uint32]*orm.Compensation{comp.ID: &commander.Compensations[0]}
	client := &connection.Client{Commander: &commander}

	listPayload := &protobuf.CS_30102{Type: proto.Uint32(0)}
	listBuffer, err := proto.Marshal(listPayload)
	if err != nil {
		t.Fatalf("failed to marshal list payload: %v", err)
	}
	if _, _, err := answer.GetCompensateList(&listBuffer, client); err != nil {
		t.Fatalf("GetCompensateList failed: %v", err)
	}
	listResponse := &protobuf.SC_30103{}
	decodeTestPacket(t, client, 30103, listResponse)
	if len(listResponse.GetTimeRewardList()) != 1 {
		t.Fatalf("expected 1 reward entry, got %d", len(listResponse.GetTimeRewardList()))
	}

	claimPayload := &protobuf.CS_30104{RewardId: proto.Uint32(comp.ID)}
	claimBuffer, err := proto.Marshal(claimPayload)
	if err != nil {
		t.Fatalf("failed to marshal claim payload: %v", err)
	}
	if _, _, err := answer.GetCompensateReward(&claimBuffer, client); err != nil {
		t.Fatalf("GetCompensateReward failed: %v", err)
	}
	claimResponse := &protobuf.SC_30105{}
	decodeTestPacket(t, client, 30105, claimResponse)
	if claimResponse.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", claimResponse.GetResult())
	}
	if claimResponse.GetNumber() != 0 {
		t.Fatalf("expected number 0, got %d", claimResponse.GetNumber())
	}
	expectedMax := uint32(comp.ExpiresAt.Unix())
	if claimResponse.GetMaxTimestamp() != expectedMax {
		t.Fatalf("expected max_timestamp %d, got %d", expectedMax, claimResponse.GetMaxTimestamp())
	}
	if len(claimResponse.GetDropList()) != 1 {
		t.Fatalf("expected drop list length 1, got %d", len(claimResponse.GetDropList()))
	}
}
