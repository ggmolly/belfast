package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type gameRoomTemplate struct {
	ID uint32 `json:"id"`
}

func EventData(buffer *[]byte, client *connection.Client) (int, int, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/game_room_template.json")
	if err != nil {
		return 0, 26120, err
	}
	response := protobuf.SC_26120{
		WeeklyFree:    proto.Uint32(0),
		MonthlyTicket: proto.Uint32(0),
		PayCoinCount:  proto.Uint32(0),
		FirstEnter:    proto.Uint32(0),
		Rooms:         make([]*protobuf.GAMEROOM, 0, len(entries)),
	}
	for _, entry := range entries {
		var room gameRoomTemplate
		if err := json.Unmarshal(entry.Data, &room); err != nil {
			return 0, 26120, err
		}
		response.Rooms = append(response.Rooms, &protobuf.GAMEROOM{
			Roomid:   proto.Uint32(room.ID),
			MaxScore: proto.Uint32(0),
		})
	}
	return client.SendMessage(26120, &response)
}
