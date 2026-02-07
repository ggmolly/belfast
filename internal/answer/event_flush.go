package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EventFlush(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13009
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13010, err
	}

	var events []orm.EventCollection
	if err := orm.GormDB.
		Where("commander_id = ? AND finish_time > 0", client.Commander.CommanderID).
		Order("collection_id asc").
		Find(&events).Error; err != nil {
		return 0, 13010, err
	}

	collectionList := make([]*protobuf.COLLECTIONINFO, 0, len(events))
	for _, event := range events {
		template, err := loadCollectionTemplate(event.CollectionID)
		if err != nil {
			return 0, 13010, err
		}
		overTime := uint32(0)
		if template != nil {
			overTime = template.OverTime
		}

		collectionList = append(collectionList, &protobuf.COLLECTIONINFO{
			Id:         proto.Uint32(event.CollectionID),
			FinishTime: proto.Uint32(event.FinishTime),
			OverTime:   proto.Uint32(overTime),
			ShipIdList: orm.ToUint32List(event.ShipIDs),
		})
	}

	response := protobuf.SC_13010{
		Result:         proto.Uint32(0),
		CollectionList: collectionList,
	}
	return client.SendMessage(13010, &response)
}
