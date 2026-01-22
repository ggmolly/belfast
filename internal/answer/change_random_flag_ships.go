package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChangeRandomFlagShips(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12208
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12209, err
	}
	response := protobuf.SC_12209{
		Result: proto.Uint32(0),
	}
	updates := make([]orm.RandomFlagShipUpdate, 0, len(data.GetShipShadowList()))
	for _, entry := range data.GetShipShadowList() {
		if entry == nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(12209, &response)
		}
		shipID := entry.GetKey()
		if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
			response.Result = proto.Uint32(1)
			return client.SendMessage(12209, &response)
		}
		flag := entry.GetValue2()
		if flag > 1 {
			response.Result = proto.Uint32(1)
			return client.SendMessage(12209, &response)
		}
		updates = append(updates, orm.RandomFlagShipUpdate{
			ShipID:    shipID,
			PhantomID: entry.GetValue1(),
			Flag:      flag,
		})
	}
	if len(updates) > 0 {
		tx := orm.GormDB.Begin()
		if err := orm.ApplyRandomFlagShipUpdates(tx, client.Commander.CommanderID, updates); err != nil {
			tx.Rollback()
			response.Result = proto.Uint32(1)
			return client.SendMessage(12209, &response)
		}
		if err := tx.Commit().Error; err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(12209, &response)
		}
	}
	return client.SendMessage(12209, &response)
}
