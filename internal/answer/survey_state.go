package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func SurveyState(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11027
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11028, err
	}
	response := protobuf.SC_11028{Result: proto.Uint32(0)}
	activity, err := activeSurveyActivity(uint32(client.Commander.Level), payload.GetSurveyId())
	if err != nil {
		return 0, 11028, err
	}
	if activity == nil || activity.SurveyID != payload.GetSurveyId() {
		return client.SendMessage(11028, &response)
	}
	state, err := orm.GetSurveyState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return client.SendMessage(11028, &response)
		}
		return 0, 11028, err
	}
	if state.SurveyID == payload.GetSurveyId() {
		response.Result = proto.Uint32(payload.GetSurveyId())
	}
	return client.SendMessage(11028, &response)
}
