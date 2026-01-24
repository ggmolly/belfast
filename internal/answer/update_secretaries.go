package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateSecretaries(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_11011
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 11012, err
	}
	response := protobuf.SC_11012{
		Result: proto.Uint32(0),
	}

	updates := make([]orm.SecretaryUpdate, 0, len(data.GetCharacter()))
	for _, ship := range data.GetCharacter() {
		if ship == nil {
			continue
		}
		updates = append(updates, orm.SecretaryUpdate{
			ShipID:    ship.GetKey(),
			PhantomID: ship.GetValue(),
		})
	}

	// Check if all ships are owned by the player
	for _, update := range updates {
		if _, ok := client.Commander.OwnedShipsMap[update.ShipID]; !ok {
			response.Result = proto.Uint32(1)
			break
		}
	}

	if *response.Result == 0 {
		if err := client.Commander.RemoveSecretaries(); err != nil {
			response.Result = proto.Uint32(1)
		} else if err := client.Commander.UpdateSecretaries(updates); err != nil { // Update secretaries
			response.Result = proto.Uint32(1)
		}
	}
	if *response.Result == 0 && len(updates) > 0 {
		if ship, ok := client.Commander.OwnedShipsMap[updates[0].ShipID]; ok {
			client.Commander.DisplayIconID = ship.ShipID
			client.Commander.DisplaySkinID = ship.SkinID
			_ = orm.GormDB.Model(client.Commander).Updates(map[string]interface{}{
				"display_icon_id": ship.ShipID,
				"display_skin_id": ship.SkinID,
			}).Error
		}
	}

	return client.SendMessage(11012, &response)
}
