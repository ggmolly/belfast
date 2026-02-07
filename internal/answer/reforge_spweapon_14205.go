package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/rng"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var reforgeSpWeaponRng = rng.NewLockedRand()

func ReforgeSpWeapon(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14205
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14205, err
	}

	response := protobuf.SC_14206{
		Result:     proto.Uint32(1),
		AttrTemp_1: proto.Uint32(0),
		AttrTemp_2: proto.Uint32(0),
	}

	if client.Commander == nil || client.Commander.OwnedSpWeaponsMap == nil {
		return client.SendMessage(14206, &response)
	}
	spweaponID := data.GetSpweaponId()
	spweapon, ok := client.Commander.OwnedSpWeaponsMap[spweaponID]
	if !ok {
		return client.SendMessage(14206, &response)
	}

	shipID := data.GetShipId()
	if shipID != 0 {
		if client.Commander.OwnedShipsMap == nil {
			return client.SendMessage(14206, &response)
		}
		if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
			return client.SendMessage(14206, &response)
		}
		if spweapon.EquippedShipID != 0 && spweapon.EquippedShipID != shipID {
			return client.SendMessage(14206, &response)
		}
	}

	if spweapon.AttrTemp1 != 0 || spweapon.AttrTemp2 != 0 {
		return client.SendMessage(14206, &response)
	}

	spweaponConfig, err := orm.GetSpWeaponDataStatisticsConfigTx(orm.GormDB, spweapon.TemplateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return client.SendMessage(14206, &response)
		}
		return 0, 14205, err
	}
	upgradeConfig, err := orm.GetSpWeaponUpgradeConfigTx(orm.GormDB, spweaponConfig.UpgradeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return client.SendMessage(14206, &response)
		}
		return 0, 14205, err
	}

	for _, cost := range upgradeConfig.ResetUseItem {
		if commanderItemCount(client.Commander, cost.ItemID) < cost.Count {
			return client.SendMessage(14206, &response)
		}
	}

	attrTemp1 := rollSpWeaponTempAttr(spweaponConfig.Value1Random)
	attrTemp2 := rollSpWeaponTempAttr(spweaponConfig.Value2Random)

	tx := orm.GormDB.Begin()
	for _, cost := range upgradeConfig.ResetUseItem {
		if cost.Count == 0 {
			continue
		}
		if err := client.Commander.ConsumeItemTx(tx, cost.ItemID, cost.Count); err != nil {
			tx.Rollback()
			return client.SendMessage(14206, &response)
		}
	}
	spweapon.AttrTemp1 = attrTemp1
	spweapon.AttrTemp2 = attrTemp2
	if err := tx.Save(spweapon).Error; err != nil {
		tx.Rollback()
		return 0, 14205, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 14205, err
	}

	response.Result = proto.Uint32(0)
	response.AttrTemp_1 = proto.Uint32(attrTemp1)
	response.AttrTemp_2 = proto.Uint32(attrTemp2)
	return client.SendMessage(14206, &response)
}

func commanderItemCount(commander *orm.Commander, itemID uint32) uint32 {
	if commander.CommanderItemsMap != nil {
		if item, ok := commander.CommanderItemsMap[itemID]; ok {
			return item.Count
		}
	}
	if commander.MiscItemsMap != nil {
		if item, ok := commander.MiscItemsMap[itemID]; ok {
			return item.Data
		}
	}
	return 0
}

func rollSpWeaponTempAttr(max uint32) uint32 {
	if max == 0 {
		return 0
	}
	// inclusive upper bound
	return reforgeSpWeaponRng.Uint32N(max + 1)
}
