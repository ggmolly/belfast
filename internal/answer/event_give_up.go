package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
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

	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		event, err := orm.GetEventCollection(tx, client.Commander.CommanderID, collectionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Result = proto.Uint32(2)
				return nil
			}
			return err
		}
		if event.FinishTime == 0 || now >= event.FinishTime {
			response.Result = proto.Uint32(2)
			return nil
		}

		event.StartTime = 0
		event.FinishTime = 0
		event.ShipIDs = orm.Int64List{}
		if err := tx.Save(event).Error; err != nil {
			return err
		}
		response.Result = proto.Uint32(0)
		return nil
	}); err != nil {
		return 0, 13008, err
	}

	return client.SendMessage(13008, &response)
}
