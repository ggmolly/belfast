package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestTaskUpdateAndSubmitSingle(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	seedConfigEntry(t, taskDataTemplateCategory, "100", `{"id":100,"target_num":2,"award_display":[[2,2001,3]],"award_choice":[]}`)

	progress := protobuf.CS_20009{Progressinfo: []*protobuf.TASK_UPDATE{{Id: proto.Uint32(100), Mode: proto.Uint32(1), Progress: proto.Uint32(2)}}}
	progressBuffer, err := proto.Marshal(&progress)
	if err != nil {
		t.Fatalf("marshal progress payload: %v", err)
	}
	if _, _, err := TaskUpdateProgress(&progressBuffer, client); err != nil {
		t.Fatalf("task update progress failed: %v", err)
	}

	var progressResponse protobuf.SC_20010
	decodeResponse(t, client, &progressResponse)
	if progressResponse.GetResult() != taskResultSuccess {
		t.Fatalf("expected progress result 0, got %d", progressResponse.GetResult())
	}

	submit := protobuf.CS_20005{Id: proto.Uint32(100)}
	submitBuffer, err := proto.Marshal(&submit)
	if err != nil {
		t.Fatalf("marshal submit payload: %v", err)
	}
	if _, _, err := TaskSubmitSingle(&submitBuffer, client); err != nil {
		t.Fatalf("task submit failed: %v", err)
	}

	var submitResponse protobuf.SC_20006
	decodeResponse(t, client, &submitResponse)
	if submitResponse.GetResult() != taskResultSuccess {
		t.Fatalf("expected submit result 0, got %d", submitResponse.GetResult())
	}
	if len(submitResponse.GetAwardList()) != 1 {
		t.Fatalf("expected one award, got %d", len(submitResponse.GetAwardList()))
	}

	count := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(2001))
	if count != 3 {
		t.Fatalf("expected item count 3, got %d", count)
	}

	if _, _, err := TaskSubmitSingle(&submitBuffer, client); err != nil {
		t.Fatalf("task submit replay failed: %v", err)
	}
	decodeResponse(t, client, &submitResponse)
	if submitResponse.GetResult() != taskResultFailure {
		t.Fatalf("expected replay submit to fail, got %d", submitResponse.GetResult())
	}
	count = queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(2001))
	if count != 3 {
		t.Fatalf("expected replay to keep item count 3, got %d", count)
	}
}

func TestTaskSubmitChoiceValidation(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	seedConfigEntry(t, taskDataTemplateCategory, "101", `{"id":101,"target_num":1,"award_display":[[2,2002,1]],"award_choice":[[[2,3001,1]],[[2,3002,2]]]}`)

	progress := protobuf.CS_20009{Progressinfo: []*protobuf.TASK_UPDATE{{Id: proto.Uint32(101), Mode: proto.Uint32(1), Progress: proto.Uint32(1)}}}
	progressBuffer, err := proto.Marshal(&progress)
	if err != nil {
		t.Fatalf("marshal progress payload: %v", err)
	}
	if _, _, err := TaskUpdateProgress(&progressBuffer, client); err != nil {
		t.Fatalf("task update progress failed: %v", err)
	}
	var progressResponse protobuf.SC_20010
	decodeResponse(t, client, &progressResponse)

	invalidChoice := protobuf.CS_20005{Id: proto.Uint32(101)}
	invalidChoiceBuffer, err := proto.Marshal(&invalidChoice)
	if err != nil {
		t.Fatalf("marshal invalid choice payload: %v", err)
	}
	if _, _, err := TaskSubmitSingle(&invalidChoiceBuffer, client); err != nil {
		t.Fatalf("task submit invalid choice failed: %v", err)
	}
	var submitResponse protobuf.SC_20006
	decodeResponse(t, client, &submitResponse)
	if submitResponse.GetResult() != taskResultFailure {
		t.Fatalf("expected missing choice to fail, got %d", submitResponse.GetResult())
	}

	validChoice := protobuf.CS_20005{Id: proto.Uint32(101), ChoiceAward: []*protobuf.DROPINFO{{Type: proto.Uint32(2), Id: proto.Uint32(3002), Number: proto.Uint32(2)}}}
	validChoiceBuffer, err := proto.Marshal(&validChoice)
	if err != nil {
		t.Fatalf("marshal valid choice payload: %v", err)
	}
	if _, _, err := TaskSubmitSingle(&validChoiceBuffer, client); err != nil {
		t.Fatalf("task submit valid choice failed: %v", err)
	}
	decodeResponse(t, client, &submitResponse)
	if submitResponse.GetResult() != taskResultSuccess {
		t.Fatalf("expected valid choice to succeed, got %d", submitResponse.GetResult())
	}

	count := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(3002))
	if count != 2 {
		t.Fatalf("expected selected choice reward count 2, got %d", count)
	}
}

func TestTaskSubmitOneStepSkipsIneligibleTasks(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	seedConfigEntry(t, taskDataTemplateCategory, "110", `{"id":110,"target_num":1,"award_display":[[2,2010,1]],"award_choice":[]}`)
	seedConfigEntry(t, taskDataTemplateCategory, "111", `{"id":111,"target_num":1,"award_display":[[2,2011,1]],"award_choice":[[[2,3011,1]]]}`)

	progress := protobuf.CS_20009{Progressinfo: []*protobuf.TASK_UPDATE{
		{Id: proto.Uint32(110), Mode: proto.Uint32(1), Progress: proto.Uint32(1)},
		{Id: proto.Uint32(111), Mode: proto.Uint32(1), Progress: proto.Uint32(1)},
	}}
	progressBuffer, err := proto.Marshal(&progress)
	if err != nil {
		t.Fatalf("marshal progress payload: %v", err)
	}
	if _, _, err := TaskUpdateProgress(&progressBuffer, client); err != nil {
		t.Fatalf("task update progress failed: %v", err)
	}
	var progressResponse protobuf.SC_20010
	decodeResponse(t, client, &progressResponse)

	oneStep := protobuf.CS_20011{IdList: []uint32{110, 111, 999}}
	oneStepBuffer, err := proto.Marshal(&oneStep)
	if err != nil {
		t.Fatalf("marshal one-step payload: %v", err)
	}
	if _, _, err := TaskSubmitOneStep(&oneStepBuffer, client); err != nil {
		t.Fatalf("task one-step submit failed: %v", err)
	}

	var response protobuf.SC_20012
	decodeResponse(t, client, &response)
	if len(response.GetIdList()) != 1 || response.GetIdList()[0] != 110 {
		t.Fatalf("expected only task 110 to submit, got %v", response.GetIdList())
	}

	count := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(2010))
	if count != 1 {
		t.Fatalf("expected one-step reward for task 110, got %d", count)
	}

	submitted110 := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM commander_common_flags WHERE commander_id = $1 AND flag_id = $2", int64(client.Commander.CommanderID), int64(taskSubmittedFlagID(110)))
	if submitted110 != 1 {
		t.Fatalf("expected task 110 to be marked submitted")
	}
	submitted111 := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM commander_common_flags WHERE commander_id = $1 AND flag_id = $2", int64(client.Commander.CommanderID), int64(taskSubmittedFlagID(111)))
	if submitted111 != 0 {
		t.Fatalf("expected choice task 111 to stay unsubmitted")
	}
}

func TestTaskUpdateProgressUnknownTaskFails(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	initCommanderMaps(client)
	clearTable(t, &orm.ConfigEntry{})

	progress := protobuf.CS_20009{Progressinfo: []*protobuf.TASK_UPDATE{{Id: proto.Uint32(9999), Mode: proto.Uint32(1), Progress: proto.Uint32(1)}}}
	progressBuffer, err := proto.Marshal(&progress)
	if err != nil {
		t.Fatalf("marshal progress payload: %v", err)
	}
	if _, _, err := TaskUpdateProgress(&progressBuffer, client); err != nil {
		t.Fatalf("task update progress failed: %v", err)
	}

	var response protobuf.SC_20010
	decodeResponse(t, client, &response)
	if response.GetResult() != taskResultFailure {
		t.Fatalf("expected unknown task progress to fail, got %d", response.GetResult())
	}
}
