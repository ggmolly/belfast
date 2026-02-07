package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
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
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if shipID != 0 {
			if err := tx.Model(&orm.OwnedSpWeapon{}).
				Where("owner_id = ? AND equipped_ship_id = ? AND id <> ?", ownerID, shipID, spweaponID).
				Update("equipped_ship_id", 0).
				Error; err != nil {
				return err
			}
		}
		return tx.Model(&orm.OwnedSpWeapon{}).
			Where("owner_id = ? AND id = ?", ownerID, spweaponID).
			Update("equipped_ship_id", shipID).
			Error
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
