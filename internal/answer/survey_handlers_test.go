package answer

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedSurveyActivityTemplate(t *testing.T, activityID uint32, surveyID uint32, minLevel int) {
	t.Helper()
	payload := fmt.Sprintf(`{"id":%d,"type":101,"config_id":%d,"config_data":[1,%d],"time":["timer",[[2025,1,1],[0,0,0]],[[2099,1,1],[0,0,0]]]}`, activityID, surveyID, minLevel)
	seedConfigEntry(t, "ShareCfg/activity_template.json", fmt.Sprintf("%d", activityID), payload)
}

func decodeSurveyPacket(t *testing.T, client *connection.Client, expectedID int, message proto.Message) {
	t.Helper()
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
}

func TestSurveyRequestSuccess(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.CommanderSurvey{})
	clearTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: 1, AccountID: 1, Name: "Survey Commander", Level: 30}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	client.Commander = &commander

	seedSurveyActivityTemplate(t, 7101, 1001, 30)
	seedActivityAllowlist(t, []uint32{7101})

	payload := &protobuf.CS_11025{SurveyId: proto.Uint32(1001)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyRequest(&buf, client); err != nil {
		t.Fatalf("SurveyRequest failed: %v", err)
	}

	response := &protobuf.SC_11026{}
	decodeSurveyPacket(t, client, 11026, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	completed, err := orm.IsCommanderSurveyCompleted(commander.CommanderID, 1001)
	if err != nil {
		t.Fatalf("check survey completion: %v", err)
	}
	if !completed {
		t.Fatalf("expected survey completion to be stored")
	}
}

func TestSurveyRequestLevelGateFailure(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.CommanderSurvey{})
	clearTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: 2, AccountID: 2, Name: "Survey Commander", Level: 10}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	client.Commander = &commander

	seedSurveyActivityTemplate(t, 7102, 1001, 30)
	seedActivityAllowlist(t, []uint32{7102})

	payload := &protobuf.CS_11025{SurveyId: proto.Uint32(1001)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyRequest(&buf, client); err != nil {
		t.Fatalf("SurveyRequest failed: %v", err)
	}

	response := &protobuf.SC_11026{}
	decodeSurveyPacket(t, client, 11026, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result for level gate failure")
	}

	completed, err := orm.IsCommanderSurveyCompleted(commander.CommanderID, 1001)
	if err != nil {
		t.Fatalf("check survey completion: %v", err)
	}
	if completed {
		t.Fatalf("expected survey completion to be unset")
	}
}

func TestSurveyStateCompleted(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.CommanderSurvey{})
	clearTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: 3, AccountID: 3, Name: "Survey Commander", Level: 30}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	client.Commander = &commander

	if err := orm.SetCommanderSurveyCompleted(orm.GormDB, commander.CommanderID, 1001, time.Now().UTC()); err != nil {
		t.Fatalf("set survey completion: %v", err)
	}

	payload := &protobuf.CS_11027{SurveyId: proto.Uint32(1001)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyState(&buf, client); err != nil {
		t.Fatalf("SurveyState failed: %v", err)
	}

	response := &protobuf.SC_11028{}
	decodeSurveyPacket(t, client, 11028, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
}

func TestSurveyStateNotCompleted(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.CommanderSurvey{})
	clearTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: 4, AccountID: 4, Name: "Survey Commander", Level: 30}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	client.Commander = &commander

	payload := &protobuf.CS_11027{SurveyId: proto.Uint32(1001)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := SurveyState(&buf, client); err != nil {
		t.Fatalf("SurveyState failed: %v", err)
	}

	response := &protobuf.SC_11028{}
	decodeSurveyPacket(t, client, 11028, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}
