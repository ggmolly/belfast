package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func SurveyRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11025
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11026, err
	}
	response := protobuf.SC_11026{Result: proto.Uint32(1)}
	activity, err := activeSurveyActivity(uint32(client.Commander.Level), payload.GetSurveyId())
	if err != nil {
		return 0, 11026, err
	}
	if activity == nil || activity.SurveyID != payload.GetSurveyId() {
		return client.SendMessage(11026, &response)
	}
	if err := upsertSurveyState(client.Commander.CommanderID, payload.GetSurveyId()); err != nil {
		return client.SendMessage(11026, &response)
	}
	response.Result = proto.Uint32(0)
	return client.SendMessage(11026, &response)
}
