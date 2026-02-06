package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/rng"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultEventFinishCritChancePercent = 10

var eventFinishRng = rng.NewLockedRand()

var eventFinishIntn = func(n int) int { return eventFinishRng.IntN(n) }

type collectionDropObject struct {
	DropType uint32          `json:"type"`
	DropID   uint32          `json:"id"`
	Nums     json.RawMessage `json:"nums"`
}

func EventFinish(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13005
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13006, err
	}

	response := protobuf.SC_13006{
		Result:        proto.Uint32(1),
		Exp:           proto.Uint32(0),
		DropList:      []*protobuf.DROPINFO{},
		NewCollection: []*protobuf.COLLECTIONINFO{},
		IsCri:         proto.Uint32(0),
	}

	collectionID := payload.GetId()
	if collectionID == 0 {
		return client.SendMessage(13006, &response)
	}

	template, err := loadCollectionTemplate(collectionID)
	if err != nil {
		return 0, 13006, err
	}
	if template == nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(13006, &response)
	}

	now := uint32(time.Now().Unix())
	if template.OverTime > 0 && now >= template.OverTime {
		response.Result = proto.Uint32(3)
		return client.SendMessage(13006, &response)
	}

	if template.DropGoldMax > 0 && client.Commander.GetResourceCount(1) >= template.DropGoldMax {
		response.Result = proto.Uint32(4)
		return client.SendMessage(13006, &response)
	}
	if template.DropOilMax > 0 && client.Commander.GetResourceCount(2) >= template.DropOilMax {
		response.Result = proto.Uint32(4)
		return client.SendMessage(13006, &response)
	}

	event, err := orm.GetEventCollection(orm.GormDB, client.Commander.CommanderID, collectionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Result = proto.Uint32(2)
			return client.SendMessage(13006, &response)
		}
		return 0, 13006, err
	}
	if event.FinishTime == 0 || now < event.FinishTime {
		response.Result = proto.Uint32(2)
		return client.SendMessage(13006, &response)
	}

	shipIDs := orm.ToUint32List(event.ShipIDs)
	if len(shipIDs) == 0 {
		response.Result = proto.Uint32(5)
		return client.SendMessage(13006, &response)
	}
	if template.ShipNum > 0 && uint32(len(shipIDs)) != template.ShipNum {
		response.Result = proto.Uint32(5)
		return client.SendMessage(13006, &response)
	}
	for _, shipID := range shipIDs {
		if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
			response.Result = proto.Uint32(5)
			return client.SendMessage(13006, &response)
		}
	}

	drops, isCri, err := buildEventFinishDrops(template)
	if err != nil {
		return 0, 13006, err
	}

	var newCollection *protobuf.COLLECTIONINFO
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("commander_id = ? AND collection_id = ?", client.Commander.CommanderID, collectionID).Delete(&orm.EventCollection{}).Error; err != nil {
			return err
		}

		for _, shipID := range shipIDs {
			owned := client.Commander.OwnedShipsMap[shipID]
			if owned == nil {
				continue
			}
			if err := applyOwnedShipExpGain(owned, template.Exp); err != nil {
				return err
			}
			if err := tx.Save(owned).Error; err != nil {
				return err
			}
		}

		for _, drop := range drops {
			if err := applyDropTx(tx, client, drop.GetType(), drop.GetId(), drop.GetNumber()); err != nil {
				return err
			}
		}

		if err := tx.Model(&orm.Commander{}).
			Where("commander_id = ?", client.Commander.CommanderID).
			Update("collect_attack_count", gorm.Expr("collect_attack_count + ?", 1)).Error; err != nil {
			return err
		}
		client.Commander.CollectAttackCount++

		picked, err := spawnNextEventCollectionTx(tx, client, now, shipIDs, collectionID)
		if err != nil {
			return err
		}
		newCollection = picked
		return nil
	}); err != nil {
		return 0, 13006, err
	}

	response.Result = proto.Uint32(0)
	response.Exp = proto.Uint32(template.Exp)
	response.DropList = dropMapToList(drops)
	if isCri {
		response.IsCri = proto.Uint32(1)
	}
	if newCollection != nil {
		response.NewCollection = []*protobuf.COLLECTIONINFO{newCollection}
	}
	return client.SendMessage(13006, &response)
}

func buildEventFinishDrops(template *collectionTemplate) (map[string]*protobuf.DROPINFO, bool, error) {
	result := make(map[string]*protobuf.DROPINFO)
	entries, err := parseCollectionDropObjects(template.DropDisplay)
	if err != nil {
		return nil, false, err
	}
	for _, entry := range entries {
		key := fmt.Sprintf("%d_%d", entry.DropType, entry.DropID)
		if existing, ok := result[key]; ok {
			existing.Number = proto.Uint32(existing.GetNumber() + entry.Count)
			continue
		}
		result[key] = newDropInfo(entry.DropType, entry.DropID, entry.Count)
	}

	specialDrops, err := parseCollectionDropObjects(template.SpecialDrop)
	if err != nil {
		return nil, false, err
	}
	if len(specialDrops) == 0 {
		return result, false, nil
	}

	critChance := defaultEventFinishCritChancePercent
	isCri := eventFinishIntn(100) < critChance
	if !isCri {
		return result, false, nil
	}
	for _, entry := range specialDrops {
		key := fmt.Sprintf("%d_%d", entry.DropType, entry.DropID)
		if existing, ok := result[key]; ok {
			existing.Number = proto.Uint32(existing.GetNumber() + entry.Count)
			continue
		}
		result[key] = newDropInfo(entry.DropType, entry.DropID, entry.Count)
	}
	return result, true, nil
}

type resolvedCollectionDrop struct {
	DropType uint32
	DropID   uint32
	Count    uint32
}

func parseCollectionDropObjects(raw json.RawMessage) ([]resolvedCollectionDrop, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var objects []collectionDropObject
	if err := json.Unmarshal(raw, &objects); err == nil {
		resolved := make([]resolvedCollectionDrop, 0, len(objects))
		for _, obj := range objects {
			count, err := resolveDropCount(obj.Nums)
			if err != nil {
				return nil, err
			}
			if obj.DropType == 0 || obj.DropID == 0 || count == 0 {
				continue
			}
			resolved = append(resolved, resolvedCollectionDrop{DropType: obj.DropType, DropID: obj.DropID, Count: count})
		}
		if len(resolved) > 0 {
			return resolved, nil
		}
	}

	var tuples []json.RawMessage
	if err := json.Unmarshal(raw, &tuples); err != nil {
		return nil, err
	}
	resolved := make([]resolvedCollectionDrop, 0, len(tuples))
	for _, tupleRaw := range tuples {
		var tuple []json.RawMessage
		if err := json.Unmarshal(tupleRaw, &tuple); err != nil {
			continue
		}
		if len(tuple) < 3 {
			continue
		}
		dropType, ok := parseUint32Raw(tuple[0])
		if !ok {
			continue
		}
		dropID, ok := parseUint32Raw(tuple[1])
		if !ok {
			continue
		}
		count, err := resolveDropCount(tuple[2])
		if err != nil {
			return nil, err
		}
		if dropType == 0 || dropID == 0 || count == 0 {
			continue
		}
		resolved = append(resolved, resolvedCollectionDrop{DropType: dropType, DropID: dropID, Count: count})
	}
	return resolved, nil
}

func parseUint32Raw(raw json.RawMessage) (uint32, bool) {
	var value uint32
	if err := json.Unmarshal(raw, &value); err == nil {
		return value, true
	}
	var value64 uint64
	if err := json.Unmarshal(raw, &value64); err == nil {
		return uint32(value64), true
	}
	var floatValue float64
	if err := json.Unmarshal(raw, &floatValue); err == nil {
		if floatValue < 0 {
			return 0, false
		}
		return uint32(floatValue), true
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		parsed, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return 0, false
		}
		return uint32(parsed), true
	}
	return 0, false
}

func resolveDropCount(raw json.RawMessage) (uint32, error) {
	if len(raw) == 0 {
		return 0, nil
	}

	var number uint32
	if err := json.Unmarshal(raw, &number); err == nil {
		return number, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		text = strings.TrimSpace(text)
		text = strings.ReplaceAll(text, "\\~", "~")
		if text == "" {
			return 0, nil
		}
		if strings.Contains(text, "~") {
			parts := strings.SplitN(text, "~", 2)
			min, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, err
			}
			max, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return 0, err
			}
			if max < min {
				min, max = max, min
			}
			return uint32(min + eventFinishIntn(max-min+1)), nil
		}
		parsed, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(parsed), nil
	}
	var pair []uint32
	if err := json.Unmarshal(raw, &pair); err == nil {
		if len(pair) == 0 {
			return 0, nil
		}
		if len(pair) == 1 {
			return pair[0], nil
		}
		min := pair[0]
		max := pair[1]
		if max < min {
			min, max = max, min
		}
		return min + uint32(eventFinishIntn(int(max-min+1))), nil
	}
	return 0, nil
}

func applyOwnedShipExpGain(owned *orm.OwnedShip, gain uint32) error {
	if gain == 0 {
		return nil
	}
	if owned.Level >= owned.MaxLevel {
		if owned.MaxLevel >= 100 {
			owned.SurplusExp = addSurplusExp(owned.SurplusExp, gain)
		}
		return nil
	}
	newExp := owned.Exp + gain
	level := owned.Level
	for level < owned.MaxLevel {
		config, err := loadShipLevelConfig(level)
		if err != nil {
			return err
		}
		if config == nil {
			break
		}
		required := config.Exp
		if owned.Ship.RarityID == 6 {
			required = config.ExpUR
		}
		if required == 0 || newExp < required {
			break
		}
		newExp -= required
		level++
	}
	owned.Exp = newExp
	owned.Level = level
	if owned.Level >= owned.MaxLevel && owned.MaxLevel >= 100 && owned.Exp > 0 {
		owned.SurplusExp = addSurplusExp(owned.SurplusExp, owned.Exp)
		owned.Exp = 0
	}
	return nil
}

func applyDropTx(tx *gorm.DB, client *connection.Client, dropType uint32, dropID uint32, dropCount uint32) error {
	if dropID == 0 || dropCount == 0 {
		return nil
	}
	switch dropType {
	case consts.DROP_TYPE_RESOURCE:
		return addResourceTx(tx, client.Commander, dropID, dropCount)
	case consts.DROP_TYPE_ITEM:
		return addItemTx(tx, client.Commander, dropID, dropCount)
	default:
		// Unsupported types are returned to the client but ignored server-side.
		return nil
	}
}

func addResourceTx(tx *gorm.DB, commander *orm.Commander, resourceID uint32, amount uint32) error {
	orm.DealiasResource(&resourceID)
	entry := orm.OwnedResource{CommanderID: commander.CommanderID, ResourceID: resourceID, Amount: amount}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "resource_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"amount": gorm.Expr("amount + ?", amount),
		}),
	}).Create(&entry).Error; err != nil {
		return err
	}
	if commander.OwnedResourcesMap == nil {
		commander.OwnedResourcesMap = make(map[uint32]*orm.OwnedResource)
	}
	if existing, ok := commander.OwnedResourcesMap[resourceID]; ok {
		existing.Amount += amount
		return nil
	}
	commander.OwnedResources = append(commander.OwnedResources, orm.OwnedResource{CommanderID: commander.CommanderID, ResourceID: resourceID, Amount: amount})
	commander.OwnedResourcesMap[resourceID] = &commander.OwnedResources[len(commander.OwnedResources)-1]
	return nil
}

func addItemTx(tx *gorm.DB, commander *orm.Commander, itemID uint32, amount uint32) error {
	entry := orm.CommanderItem{CommanderID: commander.CommanderID, ItemID: itemID, Count: amount}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "item_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"count": gorm.Expr("count + ?", amount),
		}),
	}).Create(&entry).Error; err != nil {
		return err
	}
	if commander.CommanderItemsMap == nil {
		commander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)
	}
	if existing, ok := commander.CommanderItemsMap[itemID]; ok {
		existing.Count += amount
		return nil
	}
	commander.Items = append(commander.Items, orm.CommanderItem{CommanderID: commander.CommanderID, ItemID: itemID, Count: amount})
	commander.CommanderItemsMap[itemID] = &commander.Items[len(commander.Items)-1]
	return nil
}

func spawnNextEventCollectionTx(tx *gorm.DB, client *connection.Client, now uint32, shipIDs []uint32, finishedID uint32) (*protobuf.COLLECTIONINFO, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, collectionTemplateCategory)
	if err != nil {
		return nil, err
	}

	activeCount, err := orm.GetActiveEventCount(tx, client.Commander.CommanderID)
	if err != nil {
		return nil, err
	}

	templates := make([]*collectionTemplate, 0, len(entries))
	for i := range entries {
		var tpl collectionTemplate
		if err := json.Unmarshal(entries[i].Data, &tpl); err != nil {
			return nil, err
		}
		templates = append(templates, &tpl)
	}

	allowedTypes := make(map[uint32]struct{})
	shipMeetsLevel := func(required uint32) bool {
		if required == 0 {
			return true
		}
		for _, id := range shipIDs {
			owned, ok := client.Commander.OwnedShipsMap[id]
			if ok && owned.Level >= required {
				return true
			}
		}
		return false
	}
	shipsMatchTypes := func(types []uint32) bool {
		if len(types) == 0 {
			return true
		}
		allowedTypes = make(map[uint32]struct{}, len(types))
		for _, t := range types {
			allowedTypes[t] = struct{}{}
		}
		for _, id := range shipIDs {
			owned, ok := client.Commander.OwnedShipsMap[id]
			if !ok {
				return false
			}
			if _, ok := allowedTypes[owned.Ship.Type]; !ok {
				return false
			}
		}
		return true
	}

	candidates := make([]*collectionTemplate, 0)
	for _, tpl := range templates {
		if tpl == nil {
			continue
		}
		if tpl.ID == 0 || tpl.ID == finishedID {
			continue
		}
		if tpl.OverTime > 0 && now >= tpl.OverTime {
			continue
		}
		if tpl.ShipNum > 0 && uint32(len(shipIDs)) != tpl.ShipNum {
			continue
		}
		if !shipMeetsLevel(tpl.ShipLv) {
			continue
		}
		if !shipsMatchTypes(tpl.ShipType) {
			continue
		}
		if tpl.MaxTeam > 0 && uint32(activeCount) >= tpl.MaxTeam {
			continue
		}
		if _, err := orm.GetEventCollection(tx, client.Commander.CommanderID, tpl.ID); err == nil {
			continue
		}
		candidates = append(candidates, tpl)
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	picked := candidates[eventFinishIntn(len(candidates))]
	finishTime := now
	if picked.CollectTime > 0 {
		finishTime = now + picked.CollectTime
	}
	newEvent := orm.EventCollection{
		CommanderID:  client.Commander.CommanderID,
		CollectionID: picked.ID,
		StartTime:    now,
		FinishTime:   finishTime,
		ShipIDs:      orm.ToInt64List(shipIDs),
	}
	if err := tx.Create(&newEvent).Error; err != nil {
		return nil, err
	}
	return &protobuf.COLLECTIONINFO{
		Id:         proto.Uint32(picked.ID),
		FinishTime: proto.Uint32(finishTime),
		OverTime:   proto.Uint32(picked.OverTime),
		ShipIdList: shipIDs,
	}, nil
}
