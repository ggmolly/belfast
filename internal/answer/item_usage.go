package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	itemUsageDrop              = "usage_drop"
	itemUsageDropTemplate      = "usgae_drop_template"
	itemUsageDropAppointed     = "usage_drop_appointed"
	itemUsageDropAppointedSkin = "usage_drop_appointed_skinexchange"
	itemUsageDropRandomSkin    = "usage_drop_random_skin"
	itemUsageInvitation        = "usage_invitation"
	itemUsageSkinExp           = "usage_skin_exp"
	itemUsageSkinDiscount      = "usage_skin_discount"
	itemUsageShopDiscount      = "usage_shop_discount"
)

type itemUsageConfig struct {
	ID       uint32          `json:"id"`
	Usage    string          `json:"usage"`
	UsageArg json.RawMessage `json:"usage_arg"`
	Display  string          `json:"display"`
	IconList json.RawMessage `json:"display_icon"`
}

type dropRestoreEntry struct {
	ID           uint32 `json:"id"`
	DropID       uint32 `json:"drop_id"`
	ResourceType uint32 `json:"resource_type"`
	ResourceNum  uint32 `json:"resource_num"`
	TargetType   uint32 `json:"target_type"`
	TargetID     uint32 `json:"target_id"`
	Type         uint32 `json:"type"`
}

type shopTemplateEntry struct {
	ID           uint32          `json:"id"`
	EffectArgs   json.RawMessage `json:"effect_args"`
	ResourceType uint32          `json:"resource_type"`
	ResourceNum  uint32          `json:"resource_num"`
	TimeSecond   uint32          `json:"time_second"`
}

type displayIconEntry struct {
	DropType uint32
	DropID   uint32
	Count    uint32
}

type useItemOutcome struct {
	result   uint32
	dropList []*protobuf.DROPINFO
}

func useItem(client *connection.Client, payload *protobuf.CS_15002) (*useItemOutcome, error) {
	itemId := payload.GetId()
	count := payload.GetCount()
	if count == 0 {
		return &useItemOutcome{result: 1}, nil
	}
	if client.Commander.CommanderItemsMap == nil && client.Commander.MiscItemsMap == nil {
		logger.LogEvent("Item", "Use", fmt.Sprintf("commander maps missing for item %d", itemId), logger.LOG_LEVEL_INFO)
		if err := client.Commander.Load(); err != nil {
			logger.LogEvent("Item", "Use", fmt.Sprintf("commander load failed for item %d: %v", itemId, err), logger.LOG_LEVEL_ERROR)
			return nil, err
		}
	}
	available, source, err := getCommanderItemCountFromDB(client, itemId)
	if err != nil {
		logger.LogEvent("Item", "Use", fmt.Sprintf("item count lookup failed %d: %v", itemId, err), logger.LOG_LEVEL_ERROR)
		return nil, err
	}
	logger.LogEvent("Item", "Use", fmt.Sprintf("item %d available=%d source=%s", itemId, available, source), logger.LOG_LEVEL_INFO)
	if available < count {
		logger.LogEvent("Item", "Use", fmt.Sprintf("insufficient item %d count %d", itemId, count), logger.LOG_LEVEL_INFO)
		return &useItemOutcome{result: 1}, nil
	}
	config, err := loadItemUsageConfig(itemId)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return &useItemOutcome{result: 1}, nil
	}
	plan, err := prepareItemUsage(client, config, payload.GetArg(), count)
	if err != nil {
		return nil, err
	}
	if plan.result != 0 {
		return &useItemOutcome{result: plan.result, dropList: plan.dropList}, nil
	}
	if err := consumeCommanderItemFromDB(client, itemId, count, source); err != nil {
		logger.LogEvent("Item", "Use", fmt.Sprintf("item consume failed %d: %v", itemId, err), logger.LOG_LEVEL_ERROR)
		return &useItemOutcome{result: 1}, nil
	}
	if err := plan.apply(); err != nil {
		return nil, err
	}
	return &useItemOutcome{result: 0, dropList: plan.dropList}, nil
}

type itemUsagePlan struct {
	result   uint32
	dropList []*protobuf.DROPINFO
	apply    func() error
}

func prepareItemUsage(client *connection.Client, config *itemUsageConfig, arg []uint32, count uint32) (*itemUsagePlan, error) {
	switch config.Usage {
	case itemUsageDrop:
		return prepareDropUsage(client, config, count)
	case itemUsageDropTemplate:
		return prepareDropTemplateUsage(client, config, count)
	case itemUsageDropAppointed:
		return prepareDropAppointedUsage(client, config, arg, count)
	case itemUsageDropAppointedSkin:
		return prepareSkinSelectUsage(client, config, arg, count)
	case itemUsageDropRandomSkin:
		return prepareRandomSkinUsage(client, config, count)
	case itemUsageInvitation:
		return prepareInvitationUsage(client, config, arg, count)
	case itemUsageSkinExp:
		return prepareSkinExpUsage(client, config, count)
	case itemUsageSkinDiscount, itemUsageShopDiscount:
		return prepareSkinDiscountUsage(client, config, arg, count)
	default:
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
}

func prepareDropUsage(client *connection.Client, config *itemUsageConfig, count uint32) (*itemUsagePlan, error) {
	var dropID uint32
	if err := decodeUsageArg(config.UsageArg, &dropID); err != nil {
		return nil, err
	}
	entries, err := listDropRestoreEntries(dropID)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		logger.LogEvent("Item", "Use", fmt.Sprintf("drop_id %d not found", dropID), logger.LOG_LEVEL_ERROR)
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	drops := make([]*protobuf.DROPINFO, 0, len(entries))
	for _, entry := range entries {
		drops = append(drops, &protobuf.DROPINFO{
			Type:   proto.Uint32(consts.DROP_TYPE_TRANS_ITEM),
			Id:     proto.Uint32(entry.ID),
			Number: proto.Uint32(count),
		})
	}
	return &itemUsagePlan{
		result:   0,
		dropList: drops,
		apply: func() error {
			for _, entry := range entries {
				if err := applyDropRestoreEntry(client, entry, count); err != nil {
					return err
				}
			}
			return nil
		},
	}, nil
}

func prepareDropTemplateUsage(client *connection.Client, config *itemUsageConfig, count uint32) (*itemUsagePlan, error) {
	var args []uint32
	if err := decodeUsageArg(config.UsageArg, &args); err != nil {
		return nil, err
	}
	if len(args) < 3 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	gold := args[1] * count
	oil := args[2] * count
	drops := make([]*protobuf.DROPINFO, 0, 3)
	randomDrops, err := pickDropTemplateDisplayDrops(config, count, gold > 0, oil > 0)
	if err != nil {
		return nil, err
	}
	if gold > 0 {
		drops = append(drops, &protobuf.DROPINFO{
			Type:   proto.Uint32(consts.DROP_TYPE_RESOURCE),
			Id:     proto.Uint32(1),
			Number: proto.Uint32(gold),
		})
	}
	if oil > 0 {
		drops = append(drops, &protobuf.DROPINFO{
			Type:   proto.Uint32(consts.DROP_TYPE_RESOURCE),
			Id:     proto.Uint32(2),
			Number: proto.Uint32(oil),
		})
	}
	for _, drop := range randomDrops {
		drops = append(drops, drop)
	}
	return &itemUsagePlan{
		result:   0,
		dropList: drops,
		apply: func() error {
			if gold > 0 {
				if err := client.Commander.AddResource(1, gold); err != nil {
					return err
				}
			}
			if oil > 0 {
				if err := client.Commander.AddResource(2, oil); err != nil {
					return err
				}
			}
			if err := applyDropList(client, randomDrops); err != nil {
				return err
			}
			return nil
		},
	}, nil
}

func prepareDropAppointedUsage(client *connection.Client, config *itemUsageConfig, arg []uint32, count uint32) (*itemUsagePlan, error) {
	var options [][]uint32
	if err := decodeUsageArg(config.UsageArg, &options); err != nil {
		return nil, err
	}
	selection, ok := selectDropOption(options, arg)
	if !ok {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	if len(selection) < 3 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	dropType := selection[0]
	dropID := selection[1]
	dropCount := selection[2] * count
	return &itemUsagePlan{
		result:   0,
		dropList: []*protobuf.DROPINFO{newDropInfo(dropType, dropID, dropCount)},
		apply: func() error {
			ok, err := applyDrop(client, dropType, dropID, dropCount)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("unsupported drop type %d", dropType)
			}
			return nil
		},
	}, nil
}

func prepareSkinSelectUsage(client *connection.Client, config *itemUsageConfig, arg []uint32, count uint32) (*itemUsagePlan, error) {
	selection := uint32(0)
	if len(arg) > 0 {
		selection = arg[0]
	}
	if selection == 0 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	choices, err := parseSkinExchangeChoices(config.UsageArg)
	if err != nil {
		return nil, err
	}
	if !containsUint32(choices, selection) {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	return &itemUsagePlan{
		result:   0,
		dropList: []*protobuf.DROPINFO{newDropInfo(consts.DROP_TYPE_SKIN, selection, count)},
		apply: func() error {
			for i := uint32(0); i < count; i++ {
				if err := client.Commander.GiveSkin(selection); err != nil {
					return err
				}
			}
			return nil
		},
	}, nil
}

func prepareRandomSkinUsage(client *connection.Client, config *itemUsageConfig, count uint32) (*itemUsagePlan, error) {
	choices, err := parseRandomSkinChoices(config.UsageArg)
	if err != nil {
		return nil, err
	}
	if len(choices) == 0 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	randomDrops := make(map[string]*protobuf.DROPINFO)
	for i := uint32(0); i < count; i++ {
		selection := choices[randomIndex(len(choices))]
		key := fmt.Sprintf("%d_%d", consts.DROP_TYPE_SKIN, selection)
		if existing := randomDrops[key]; existing != nil {
			existing.Number = proto.Uint32(existing.GetNumber() + 1)
			continue
		}
		randomDrops[key] = newDropInfo(consts.DROP_TYPE_SKIN, selection, 1)
	}
	return &itemUsagePlan{
		result:   0,
		dropList: dropMapToList(randomDrops),
		apply: func() error {
			return applyDropList(client, randomDrops)
		},
	}, nil
}

func prepareInvitationUsage(client *connection.Client, config *itemUsageConfig, arg []uint32, count uint32) (*itemUsagePlan, error) {
	if len(arg) == 0 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	var choices []uint32
	if err := decodeUsageArg(config.UsageArg, &choices); err != nil {
		return nil, err
	}
	selection := arg[0]
	if !containsUint32(choices, selection) {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	return &itemUsagePlan{
		result:   0,
		dropList: []*protobuf.DROPINFO{newDropInfo(consts.DROP_TYPE_SHIP, selection, count)},
		apply: func() error {
			for i := uint32(0); i < count; i++ {
				if _, err := client.Commander.AddShip(selection); err != nil {
					return err
				}
			}
			return nil
		},
	}, nil
}

func prepareSkinExpUsage(client *connection.Client, config *itemUsageConfig, count uint32) (*itemUsagePlan, error) {
	var args []uint32
	if err := decodeUsageArg(config.UsageArg, &args); err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	shopEntry, effectArgs, err := loadShopTemplate(args[0])
	if err != nil {
		return nil, err
	}
	if shopEntry == nil || len(effectArgs) == 0 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	if shopEntry.TimeSecond == 0 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	base := time.Now()
	skinId := effectArgs[0]
	if client.Commander.OwnedSkinsMap != nil {
		if owned, ok := client.Commander.OwnedSkinsMap[skinId]; ok {
			if owned.ExpiresAt != nil && owned.ExpiresAt.After(base) {
				base = *owned.ExpiresAt
			}
		}
	} else {
		var owned orm.OwnedSkin
		err := orm.GormDB.Where("commander_id = ? AND skin_id = ?", client.Commander.CommanderID, skinId).First(&owned).Error
		if err == nil && owned.ExpiresAt != nil && owned.ExpiresAt.After(base) {
			base = *owned.ExpiresAt
		}
	}
	expiry := base.Add(time.Second * time.Duration(shopEntry.TimeSecond) * time.Duration(count))
	return &itemUsagePlan{
		result: 0,
		apply: func() error {
			return client.Commander.GiveSkinWithExpiry(effectArgs[0], &expiry)
		},
	}, nil
}

func prepareSkinDiscountUsage(client *connection.Client, config *itemUsageConfig, arg []uint32, count uint32) (*itemUsagePlan, error) {
	if len(arg) == 0 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	shopID := arg[0]
	allowed, discount, err := parseDiscountArgs(config.UsageArg)
	if err != nil {
		return nil, err
	}
	if len(allowed) > 0 && !containsUint32(allowed, 0) && !containsUint32(allowed, shopID) {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	shopEntry, effectArgs, err := loadShopTemplate(shopID)
	if err != nil {
		return nil, err
	}
	if shopEntry == nil || len(effectArgs) == 0 {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	cost := uint32(0)
	if shopEntry.ResourceNum > discount {
		cost = shopEntry.ResourceNum - discount
	}
	cost = uint32(math.Max(0, float64(cost)))
	if cost > 0 && !client.Commander.HasEnoughResource(shopEntry.ResourceType, cost*count) {
		return &itemUsagePlan{result: 1, apply: func() error { return nil }}, nil
	}
	return &itemUsagePlan{
		result:   0,
		dropList: []*protobuf.DROPINFO{newDropInfo(consts.DROP_TYPE_SKIN, effectArgs[0], count)},
		apply: func() error {
			if cost > 0 {
				if err := client.Commander.ConsumeResource(shopEntry.ResourceType, cost*count); err != nil {
					return err
				}
			}
			for i := uint32(0); i < count; i++ {
				if err := client.Commander.GiveSkin(effectArgs[0]); err != nil {
					return err
				}
			}
			return nil
		},
	}, nil
}

func applyDrop(client *connection.Client, dropType uint32, dropID uint32, dropCount uint32) (bool, error) {
	switch dropType {
	case consts.DROP_TYPE_RESOURCE:
		return true, client.Commander.AddResource(dropID, dropCount)
	case consts.DROP_TYPE_ITEM:
		return true, client.Commander.AddItem(dropID, dropCount)
	case consts.DROP_TYPE_SHIP:
		for i := uint32(0); i < dropCount; i++ {
			if _, err := client.Commander.AddShip(dropID); err != nil {
				return true, err
			}
		}
		return true, nil
	case consts.DROP_TYPE_SKIN:
		for i := uint32(0); i < dropCount; i++ {
			if err := client.Commander.GiveSkin(dropID); err != nil {
				return true, err
			}
		}
		return true, nil
	case consts.DROP_TYPE_VITEM:
		return true, nil
	default:
		return false, nil
	}
}

func applyDropRestoreEntry(client *connection.Client, entry dropRestoreEntry, count uint32) error {
	amount := entry.ResourceNum * count
	switch entry.Type {
	case consts.DROP_TYPE_RESOURCE:
		return client.Commander.AddResource(entry.ResourceType, amount)
	case consts.DROP_TYPE_ITEM:
		return client.Commander.AddItem(entry.ResourceType, amount)
	case consts.DROP_TYPE_SHIP:
		for i := uint32(0); i < amount; i++ {
			if _, err := client.Commander.AddShip(entry.ResourceType); err != nil {
				return err
			}
		}
		return nil
	case consts.DROP_TYPE_SKIN:
		for i := uint32(0); i < amount; i++ {
			if err := client.Commander.GiveSkin(entry.ResourceType); err != nil {
				return err
			}
		}
		return nil
	default:
		logger.LogEvent("Item", "Use", fmt.Sprintf("unsupported trans drop type %d", entry.Type), logger.LOG_LEVEL_ERROR)
		return nil
	}
}

func loadItemUsageConfig(itemId uint32) (*itemUsageConfig, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, "sharecfgdata/item_data_statistics.json", fmt.Sprintf("%d", itemId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var config itemUsageConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func listDropRestoreEntries(dropID uint32) ([]dropRestoreEntry, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/drop_data_restore.json")
	if err != nil {
		return nil, err
	}
	result := make([]dropRestoreEntry, 0)
	for _, entry := range entries {
		var parsed dropRestoreEntry
		if err := json.Unmarshal(entry.Data, &parsed); err != nil {
			return nil, err
		}
		if parsed.DropID == dropID {
			result = append(result, parsed)
		}
	}
	return result, nil
}

func loadShopTemplate(shopID uint32) (*shopTemplateEntry, []uint32, error) {
	var entry orm.ConfigEntry
	result := orm.GormDB.Where("category = ? AND key = ?", "ShareCfg/shop_template.json", fmt.Sprintf("%d", shopID)).Limit(1).Find(&entry)
	if result.Error != nil {
		return nil, nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil, nil
	}
	var config shopTemplateEntry
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, nil, err
	}
	args, err := decodeEffectArgs(config.EffectArgs)
	if err != nil {
		return nil, nil, err
	}
	return &config, args, nil
}

func normalizeUsageArg(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		text = strings.TrimSpace(text)
		if text == "" {
			text = "[]"
		}
		if !json.Valid([]byte(text)) {
			return nil, fmt.Errorf("invalid usage_arg: %s", text)
		}
		return json.RawMessage([]byte(text)), nil
	}
	return raw, nil
}

func decodeUsageArg(raw json.RawMessage, out any) error {
	normalized, err := normalizeUsageArg(raw)
	if err != nil {
		return err
	}
	if len(normalized) == 0 {
		return nil
	}
	return json.Unmarshal(normalized, out)
}

func decodeEffectArgs(raw json.RawMessage) ([]uint32, error) {
	var args []uint32
	if err := decodeUsageArg(raw, &args); err != nil {
		return nil, err
	}
	return args, nil
}

func parseDisplayIcon(raw json.RawMessage) ([]displayIconEntry, error) {
	var entries [][]uint32
	if err := decodeUsageArg(raw, &entries); err != nil {
		return nil, err
	}
	result := make([]displayIconEntry, 0, len(entries))
	for _, entry := range entries {
		if len(entry) < 3 {
			continue
		}
		result = append(result, displayIconEntry{
			DropType: entry[0],
			DropID:   entry[1],
			Count:    entry[2],
		})
	}
	return result, nil
}

func filterDisplayIcons(entries []displayIconEntry, hasGold bool, hasOil bool) []displayIconEntry {
	if !hasGold && !hasOil {
		return entries
	}
	filtered := make([]displayIconEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.DropType == consts.DROP_TYPE_RESOURCE {
			if hasGold && entry.DropID == 1 {
				continue
			}
			if hasOil && entry.DropID == 2 {
				continue
			}
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func pickDropTemplateDisplayDrops(config *itemUsageConfig, count uint32, hasGold bool, hasOil bool) (map[string]*protobuf.DROPINFO, error) {
	randomDrops := make(map[string]*protobuf.DROPINFO)
	entries, err := parseDisplayIcon(config.IconList)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return randomDrops, nil
	}
	filtered := filterDisplayIcons(entries, hasGold, hasOil)
	if len(filtered) == 0 {
		return randomDrops, nil
	}
	for i := uint32(0); i < count; i++ {
		entry := filtered[randomIndex(len(filtered))]
		key := fmt.Sprintf("%d_%d", entry.DropType, entry.DropID)
		existing := randomDrops[key]
		if existing == nil {
			randomDrops[key] = newDropInfo(entry.DropType, entry.DropID, entry.Count)
			continue
		}
		existing.Number = proto.Uint32(existing.GetNumber() + entry.Count)
	}
	return randomDrops, nil
}

func applyDropList(client *connection.Client, drops map[string]*protobuf.DROPINFO) error {
	for _, drop := range drops {
		ok, err := applyDrop(client, drop.GetType(), drop.GetId(), drop.GetNumber())
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("unsupported drop type %d", drop.GetType())
		}
	}
	return nil
}

func dropMapToList(drops map[string]*protobuf.DROPINFO) []*protobuf.DROPINFO {
	list := make([]*protobuf.DROPINFO, 0, len(drops))
	for _, drop := range drops {
		list = append(list, drop)
	}
	return list
}

func parseSkinExchangeChoices(raw json.RawMessage) ([]uint32, error) {
	var entries []json.RawMessage
	if err := decodeUsageArg(raw, &entries); err != nil {
		return nil, err
	}
	choices := make([]uint32, 0)
	for _, entry := range entries {
		var list []uint32
		if err := decodeUsageArg(entry, &list); err == nil {
			choices = append(choices, list...)
		}
	}
	return choices, nil
}

func parseRandomSkinChoices(raw json.RawMessage) ([]uint32, error) {
	var entries []json.RawMessage
	if err := decodeUsageArg(raw, &entries); err != nil {
		return nil, err
	}
	if len(entries) < 3 {
		return nil, nil
	}
	var list []uint32
	if err := decodeUsageArg(entries[2], &list); err != nil {
		return nil, err
	}
	return list, nil
}

func parseDiscountArgs(raw json.RawMessage) ([]uint32, uint32, error) {
	var entries []json.RawMessage
	if err := decodeUsageArg(raw, &entries); err != nil {
		return nil, 0, err
	}
	allowed := []uint32{}
	if len(entries) > 0 {
		if err := decodeUsageArg(entries[0], &allowed); err != nil {
			var single uint32
			if err := decodeUsageArg(entries[0], &single); err == nil {
				allowed = []uint32{single}
			}
		}
	}
	var discount uint32
	if len(entries) > 1 {
		if err := decodeUsageArg(entries[1], &discount); err != nil {
			return nil, 0, err
		}
	}
	return allowed, discount, nil
}

func selectDropOption(options [][]uint32, arg []uint32) ([]uint32, bool) {
	if len(options) == 0 {
		return nil, false
	}
	if len(arg) >= 3 {
		for _, option := range options {
			if len(option) < 3 {
				continue
			}
			if option[0] == arg[0] && option[1] == arg[1] && option[2] == arg[2] {
				return option, true
			}
		}
		return nil, false
	}
	return options[0], true
}

func newDropInfo(dropType uint32, dropID uint32, count uint32) *protobuf.DROPINFO {
	return &protobuf.DROPINFO{
		Type:   proto.Uint32(dropType),
		Id:     proto.Uint32(dropID),
		Number: proto.Uint32(count),
	}
}

func randomIndex(max int) int {
	if max <= 1 {
		return 0
	}
	return int(time.Now().UnixNano() % int64(max))
}

func getCommanderItemCountFromDB(client *connection.Client, itemId uint32) (uint32, string, error) {
	var item orm.CommanderItem
	itemResult := orm.GormDB.Where("commander_id = ? AND item_id = ?", client.Commander.CommanderID, itemId).First(&item)
	if itemResult.Error == nil {
		return item.Count, "items", nil
	}
	if itemResult.Error != nil && !errors.Is(itemResult.Error, gorm.ErrRecordNotFound) {
		return 0, "", itemResult.Error
	}

	var misc orm.CommanderMiscItem
	miscResult := orm.GormDB.Where("commander_id = ? AND item_id = ?", client.Commander.CommanderID, itemId).First(&misc)
	if miscResult.Error == nil {
		return misc.Data, "misc", nil
	}
	if miscResult.Error != nil && !errors.Is(miscResult.Error, gorm.ErrRecordNotFound) {
		return 0, "", miscResult.Error
	}
	return 0, "none", nil
}

func consumeCommanderItemFromDB(client *connection.Client, itemId uint32, count uint32, source string) error {
	switch source {
	case "items":
		result := orm.GormDB.Model(&orm.CommanderItem{}).
			Where("commander_id = ? AND item_id = ? AND count >= ?", client.Commander.CommanderID, itemId, count).
			Update("count", gorm.Expr("count - ?", count))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("insufficient item")
		}
		if client.Commander.CommanderItemsMap != nil {
			if entry, ok := client.Commander.CommanderItemsMap[itemId]; ok {
				if entry.Count >= count {
					entry.Count -= count
				} else {
					entry.Count = 0
				}
			}
		}
		return nil
	case "misc":
		result := orm.GormDB.Model(&orm.CommanderMiscItem{}).
			Where("commander_id = ? AND item_id = ? AND data >= ?", client.Commander.CommanderID, itemId, count).
			Update("data", gorm.Expr("data - ?", count))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("insufficient item")
		}
		if client.Commander.MiscItemsMap != nil {
			if entry, ok := client.Commander.MiscItemsMap[itemId]; ok {
				if entry.Data >= count {
					entry.Data -= count
				} else {
					entry.Data = 0
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("item source not found")
	}
}
