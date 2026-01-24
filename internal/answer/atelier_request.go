package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func AtelierRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_26051
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 26052, err
	}

	activity, err := loadActivityTemplate(payload.GetActId())
	if err != nil {
		return 0, 26052, err
	}
	if activity.Type != activityTypeAtelierLink {
		return 0, 26052, fmt.Errorf("unexpected atelier activity type: %d", activity.Type)
	}

	response := protobuf.SC_26052{
		Result:  proto.Uint32(0),
		Items:   []*protobuf.KVDATA{},
		Recipes: []*protobuf.KVDATA{},
		// TODO: Populate atelier buff slots once atelier state is stored.
		Slots: []*protobuf.BUFF_SLOT{},
	}
	return client.SendMessage(26052, &response)
}
