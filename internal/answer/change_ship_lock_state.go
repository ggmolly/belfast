package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChangeShipLockState(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12022
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12023, err
	}
	response := protobuf.SC_12023{
		Result: proto.Uint32(1),
	}

	shipList := make([]*orm.OwnedShip, len(data.GetShipIdList()))

	for i, shipId := range data.GetShipIdList() {
		ship, ok := client.Commander.OwnedShipsMap[shipId]
		if !ok {
			return client.SendMessage(12023, &response)
		}
		shipList[i] = ship
	}

	var newState bool
	if *data.IsLocked != 0 {
		newState = true
	}
	tx := orm.GormDB.Begin()
	for _, ship := range shipList {
		ship.IsLocked = newState
		if err := orm.GormDB.Save(ship).Error; err != nil {
			tx.Rollback()
			return client.SendMessage(12023, &response)
		}
	}
	tx.Commit()
	response.Result = proto.Uint32(0)
	return client.SendMessage(12023, &response)
}
