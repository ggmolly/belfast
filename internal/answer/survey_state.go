package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func SurveyState(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11027
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11028, err
	}
	response := protobuf.SC_11028{
		Result: proto.Uint32(0),
	}
	completed, err := orm.IsCommanderSurveyCompleted(client.Commander.CommanderID, payload.GetSurveyId())
	if err != nil {
		return 0, 11028, err
	}
	if completed {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11028, &response)
}
