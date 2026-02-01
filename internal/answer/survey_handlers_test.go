package answer

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func seedSurveyActivity(t *testing.T, activityID uint32, surveyID uint32, requiredLevel uint32) {
	t.Helper()
	seedConfigEntry(t, "ShareCfg/activity_template.json", fmt.Sprintf("%d", activityID),
		fmt.Sprintf("{\"id\":%d,\"type\":101,\"config_id\":%d,\"config_data\":[1,%d]}", activityID, surveyID, requiredLevel))
	seedActivityAllowlist(t, []uint32{activityID})
}

func seedSurveyActivities(t *testing.T, activities []surveyActivitySeed) {
	t.Helper()
	allowlist := make([]uint32, 0, len(activities))
	for _, activity := range activities {
		seedConfigEntry(t, "ShareCfg/activity_template.json", fmt.Sprintf("%d", activity.ActivityID),
			fmt.Sprintf("{\"id\":%d,\"type\":101,\"config_id\":%d,\"config_data\":[1,%d]}", activity.ActivityID, activity.SurveyID, activity.RequiredLevel))
		allowlist = append(allowlist, activity.ActivityID)
	}
	seedActivityAllowlist(t, allowlist)
}

type surveyActivitySeed struct {
	ActivityID    uint32
	SurveyID      uint32
	RequiredLevel uint32
}

func TestSurveyRequestMarksComplete(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SurveyState{})
	seedSurveyActivity(t, 1, 1001, 30)

	payload := protobuf.CS_11025{SurveyId: proto.Uint32(1001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyRequest(&buffer, client); err != nil {
		t.Fatalf("survey request failed: %v", err)
	}
	var response protobuf.SC_11026
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	state, err := orm.GetSurveyState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load survey state: %v", err)
	}
	if state.SurveyID != 1001 {
		t.Fatalf("expected survey id 1001, got %d", state.SurveyID)
	}
}

func TestSurveyRequestRejectsMismatchedSurvey(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SurveyState{})
	seedSurveyActivity(t, 1, 1001, 30)

	payload := protobuf.CS_11025{SurveyId: proto.Uint32(1002)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyRequest(&buffer, client); err != nil {
		t.Fatalf("survey request failed: %v", err)
	}
	var response protobuf.SC_11026
	decodeResponse(t, client, &response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	_, err = orm.GetSurveyState(orm.GormDB, client.Commander.CommanderID)
	if err == nil {
		t.Fatalf("expected no survey state")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found, got %v", err)
	}
}

func TestSurveyRequestHonorsRequestedSurvey(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SurveyState{})
	seedSurveyActivities(t, []surveyActivitySeed{
		{ActivityID: 1, SurveyID: 1001, RequiredLevel: 30},
		{ActivityID: 2, SurveyID: 1002, RequiredLevel: 30},
	})

	payload := protobuf.CS_11025{SurveyId: proto.Uint32(1002)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyRequest(&buffer, client); err != nil {
		t.Fatalf("survey request failed: %v", err)
	}
	var response protobuf.SC_11026
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	state, err := orm.GetSurveyState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load survey state: %v", err)
	}
	if state.SurveyID != 1002 {
		t.Fatalf("expected survey id 1002, got %d", state.SurveyID)
	}
}

func TestSurveyStateReportsCompletion(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SurveyState{})
	seedSurveyActivity(t, 1, 1001, 30)
	if err := orm.UpsertSurveyState(orm.GormDB, &orm.SurveyState{CommanderID: client.Commander.CommanderID, SurveyID: 1001}); err != nil {
		t.Fatalf("seed survey state: %v", err)
	}

	payload := protobuf.CS_11027{SurveyId: proto.Uint32(1001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyState(&buffer, client); err != nil {
		t.Fatalf("survey state failed: %v", err)
	}
	var response protobuf.SC_11028
	decodeResponse(t, client, &response)
	if response.GetResult() != 1001 {
		t.Fatalf("expected result 1001, got %d", response.GetResult())
	}
}

func TestSurveyStateReturnsZeroWhenNotCompleted(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.SurveyState{})
	seedSurveyActivity(t, 1, 1001, 30)

	payload := protobuf.CS_11027{SurveyId: proto.Uint32(1001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyState(&buffer, client); err != nil {
		t.Fatalf("survey state failed: %v", err)
	}
	var response protobuf.SC_11028
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}
