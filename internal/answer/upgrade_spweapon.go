package answer

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const spweaponDataStatisticsCategory = "ShareCfg/spweapon_data_statistics.json"

func UpgradeSpWeapon(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_14203
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 14204, err
	}

	response := protobuf.SC_14204{Result: proto.Uint32(1)}
	if client == nil || client.Commander == nil {
		return client.SendMessage(14204, &response)
	}

	commander := client.Commander
	if commander.OwnedSpWeaponsMap == nil || commander.OwnedShipsMap == nil || commander.OwnedResourcesMap == nil || commander.CommanderItemsMap == nil || commander.MiscItemsMap == nil {
		if err := commander.Load(); err != nil {
			return 0, 14204, err
		}
	}

	spweaponID := payload.GetSpweaponId()
	if spweaponID == 0 {
		return client.SendMessage(14204, &response)
	}
	target, ok := commander.OwnedSpWeaponsMap[spweaponID]
	if !ok {
		return client.SendMessage(14204, &response)
	}

	shipID := payload.GetShipId()
	if shipID != 0 {
		if _, ok := commander.OwnedShipsMap[shipID]; !ok {
			return client.SendMessage(14204, &response)
		}
		if target.EquippedShipID != 0 && target.EquippedShipID != shipID {
			return client.SendMessage(14204, &response)
		}
	}

	consumeSpweaponIDs := payload.GetSpweaponIdList()
	seenSpweapons := make(map[uint32]struct{}, len(consumeSpweaponIDs))
	ptGain := uint32(0)
	for _, consumeID := range consumeSpweaponIDs {
		if consumeID == 0 || consumeID == spweaponID {
			return client.SendMessage(14204, &response)
		}
		if _, ok := seenSpweapons[consumeID]; ok {
			return client.SendMessage(14204, &response)
		}
		seenSpweapons[consumeID] = struct{}{}
		consume, ok := commander.OwnedSpWeaponsMap[consumeID]
		if !ok {
			return client.SendMessage(14204, &response)
		}
		gain, err := spweaponConsumePt(consume.TemplateID)
		if err != nil {
			return 0, 14204, err
		}
		ptGain += gain
		ptGain += consume.Pt
	}

	itemIDs := payload.GetItemIdList()
	itemCounts := countIDList(itemIDs)
	for itemID, count := range itemCounts {
		if itemID == 0 {
			return client.SendMessage(14204, &response)
		}
		if !commander.HasEnoughItem(itemID, count) {
			return client.SendMessage(14204, &response)
		}
		gain, err := itemConsumePt(itemID)
		if err != nil {
			return 0, 14204, err
		}
		if gain == 0 {
			continue
		}
		gain64 := uint64(gain) * uint64(count)
		if gain64 > math.MaxUint32 {
			return client.SendMessage(14204, &response)
		}
		ptGain += uint32(gain64)
	}

	ptTotal := target.Pt
	if uint64(ptTotal)+uint64(ptGain) > math.MaxUint32 {
		return client.SendMessage(14204, &response)
	}
	ptTotal += ptGain

	upgradedTemplateID, remainderPt, goldCost, upgraded, err := computeSpweaponUpgrade(target.TemplateID, ptTotal)
	if err != nil {
		return 0, 14204, err
	}
	if !upgraded && ptGain == 0 {
		return client.SendMessage(14204, &response)
	}
	if goldCost != 0 && !commander.HasEnoughGold(goldCost) {
		return client.SendMessage(14204, &response)
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(14204, &response)
	}

	if goldCost != 0 {
		if err := commander.ConsumeResourceTx(tx, 1, goldCost); err != nil {
			tx.Rollback()
			return client.SendMessage(14204, &response)
		}
	}
	for itemID, count := range itemCounts {
		if count == 0 {
			continue
		}
		if err := commander.ConsumeItemTx(tx, itemID, count); err != nil {
			tx.Rollback()
			return client.SendMessage(14204, &response)
		}
	}
	for _, consumeID := range consumeSpweaponIDs {
		if err := commander.RemoveOwnedSpWeaponTx(tx, consumeID); err != nil {
			tx.Rollback()
			return client.SendMessage(14204, &response)
		}
	}

	// RemoveOwnedSpWeaponTx can reallocate OwnedSpWeapons and invalidate earlier pointers.
	target, ok = commander.OwnedSpWeaponsMap[spweaponID]
	if !ok {
		tx.Rollback()
		return client.SendMessage(14204, &response)
	}
	target.TemplateID = upgradedTemplateID
	target.Pt = remainderPt
	if err := orm.UpsertOwnedSpWeaponTx(tx, target); err != nil {
		tx.Rollback()
		return 0, 14204, err
	}
	if err := tx.Commit().Error; err != nil {
		return client.SendMessage(14204, &response)
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(14204, &response)
}

func countIDList(ids []uint32) map[uint32]uint32 {
	counts := make(map[uint32]uint32, len(ids))
	for _, id := range ids {
		counts[id]++
	}
	return counts
}

func computeSpweaponUpgrade(startTemplateID uint32, pt uint32) (templateID uint32, remainder uint32, goldCost uint32, upgraded bool, err error) {
	templateID = startTemplateID
	remainder = pt
	goldCost = 0
	for i := 0; i < 20; i++ {
		next, needPt, stepGold, err := spweaponUpgradeStepConfig(templateID)
		if err != nil {
			return 0, 0, 0, false, err
		}
		if next == 0 || needPt == 0 {
			break
		}
		if remainder < needPt {
			break
		}
		remainder -= needPt
		goldCost += stepGold
		templateID = next
		upgraded = true
	}
	return templateID, remainder, goldCost, upgraded, nil
}

func spweaponUpgradeStepConfig(templateID uint32) (next uint32, needPt uint32, gold uint32, err error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, spweaponDataStatisticsCategory, strconv.FormatUint(uint64(templateID), 10))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, 0, nil
		}
		return 0, 0, 0, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(entry.Data, &raw); err != nil {
		return 0, 0, 0, err
	}
	if v, ok := rawUint32(raw["next"]); ok {
		next = v
	} else if v, ok := rawUint32(raw["upgrade_id"]); ok {
		next = v
	} else if v, ok := rawUint32(raw["upgrade_to"]); ok {
		next = v
	}

	if v, ok := rawUint32(raw["upgrade_pt"]); ok {
		needPt = v
	} else if v, ok := rawUint32(raw["upgrade_need_pt"]); ok {
		needPt = v
	} else if v, ok := rawUint32(raw["pt"]); ok {
		needPt = v
	}

	if v, ok := rawUint32(raw["upgrade_use_gold"]); ok {
		gold = v
	} else if v, ok := rawUint32(raw["trans_use_gold"]); ok {
		gold = v
	} else if v, ok := rawUint32(raw["use_gold"]); ok {
		gold = v
	}
	return next, needPt, gold, nil
}

func spweaponConsumePt(templateID uint32) (uint32, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, spweaponDataStatisticsCategory, strconv.FormatUint(uint64(templateID), 10))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(entry.Data, &raw); err != nil {
		return 0, err
	}
	if v, ok := rawUint32(raw["upgrade_get_pt"]); ok {
		return v, nil
	}
	if v, ok := rawUint32(raw["destory_get_pt"]); ok {
		return v, nil
	}
	if v, ok := rawUint32(raw["consume_pt"]); ok {
		return v, nil
	}
	return 0, nil
}

func itemConsumePt(itemID uint32) (uint32, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, itemDataStatisticsCategory, strconv.FormatUint(uint64(itemID), 10))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(entry.Data, &raw); err != nil {
		return 0, err
	}
	if v, ok := rawUint32(raw["spweapon_pt"]); ok {
		return v, nil
	}
	if v, ok := rawUint32(raw["pt"]); ok {
		return v, nil
	}
	if rawUsage, ok := raw["usage_arg"]; ok {
		var direct uint32
		if err := json.Unmarshal(rawUsage, &direct); err == nil {
			return direct, nil
		}
		var list []uint32
		if err := json.Unmarshal(rawUsage, &list); err == nil {
			if len(list) > 0 {
				return list[0], nil
			}
			return 0, nil
		}
	}
	return 0, nil
}

func rawUint32(raw json.RawMessage) (uint32, bool) {
	if len(raw) == 0 {
		return 0, false
	}
	var v uint64
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0, false
	}
	if v > math.MaxUint32 {
		return 0, false
	}
	return uint32(v), true
}
