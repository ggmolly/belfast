package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EventGiveUp(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13007
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13008, err
	}

	response := protobuf.SC_13008{Result: proto.Uint32(1)}
	collectionID := payload.GetId()
	if collectionID == 0 {
		return client.SendMessage(13008, &response)
	}

	now := uint32(time.Now().Unix())
	template, err := loadCollectionTemplate(collectionID)
	if err != nil {
		return 0, 13008, err
	}
	if template != nil && template.OverTime > 0 && now >= template.OverTime {
		response.Result = proto.Uint32(3)
		return client.SendMessage(13008, &response)
	}

	event, err := orm.GetEventCollection(nil, client.Commander.CommanderID, collectionID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			response.Result = proto.Uint32(2)
			return client.SendMessage(13008, &response)
		}
		return 0, 13008, err
	}
	if event.FinishTime == 0 || now >= event.FinishTime {
		response.Result = proto.Uint32(2)
		return client.SendMessage(13008, &response)
	}

	event.StartTime = 0
	event.FinishTime = 0
	event.ShipIDs = orm.Int64List{}
	if err := orm.SaveEventCollection(nil, event); err != nil {
		return 0, 13008, err
	}
	response.Result = proto.Uint32(0)

	return client.SendMessage(13008, &response)
}
