package answer

import (
	"context"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

func EquipSpWeapon(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14201
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14202, err
	}

	response := protobuf.SC_14202{Result: proto.Uint32(1)}
	if client.Commander == nil || client.Commander.OwnedSpWeaponsMap == nil || client.Commander.OwnedShipsMap == nil {
		return client.SendMessage(14202, &response)
	}

	spweaponID := data.GetSpweaponId()
	shipID := data.GetShipId()
	if spweaponID == 0 {
		return client.SendMessage(14202, &response)
	}
	if shipID != 0 {
		if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
			return client.SendMessage(14202, &response)
		}
	}

	if _, ok := client.Commander.OwnedSpWeaponsMap[spweaponID]; !ok {
		return client.SendMessage(14202, &response)
	}

	ownerID := client.Commander.CommanderID
	ctx := context.Background()
	if err := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if shipID != 0 {
			if _, err := tx.Exec(ctx, `
UPDATE owned_spweapons
SET equipped_ship_id = 0
WHERE owner_id = $1
  AND equipped_ship_id = $2
  AND id <> $3
`, int64(ownerID), int64(shipID), int64(spweaponID)); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `
UPDATE owned_spweapons
SET equipped_ship_id = $3
WHERE owner_id = $1
  AND id = $2
`, int64(ownerID), int64(spweaponID), int64(shipID))
		return err
	}); err != nil {
		return 0, 14202, err
	}

	// Keep the in-memory commander snapshot in sync with persisted equip state.
	if shipID != 0 {
		for i := range client.Commander.OwnedSpWeapons {
			entry := &client.Commander.OwnedSpWeapons[i]
			if entry.ID != spweaponID && entry.EquippedShipID == shipID {
				entry.EquippedShipID = 0
			}
		}
	}
	if spweapon := client.Commander.OwnedSpWeaponsMap[spweaponID]; spweapon != nil {
		spweapon.EquippedShipID = shipID
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(14202, &response)
}
