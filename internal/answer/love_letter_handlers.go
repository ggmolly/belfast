package answer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

const (
	loveLetterResultSuccess = uint32(0)
	loveLetterResultFailed  = uint32(1)

	loveLetterCharacterTemplateCategory = "ShareCfg/lover_character_template.json"
	loveLetterContentTemplateCategory   = "ShareCfg/lover_letter_content.json"
	loveLetterRewardTemplateCategory    = "ShareCfg/lover_reward.json"
	loveLetterLegacyTemplateCategory    = "ShareCfg/loveletter_2018_2021.json"

	loveLetterTextTemplateCategory = "ShareCfg/lover_letter_text.json"
)

type loveLetterCharacterConfig struct {
	ID            uint32   `json:"id"`
	ExpUp         uint32   `json:"exp_up"`
	ExpUpperLimit uint32   `json:"exp_upper_limit"`
	RelateGroupID []uint32 `json:"relate_group_id"`
}

type loveLetterContentConfig struct {
	ID        uint32   `json:"id"`
	ShipGroup uint32   `json:"ship_group"`
	Year      uint32   `json:"year"`
	LoveItem  []uint32 `json:"love_item"`
	Content   string   `json:"content"`
}

type loveLetterRewardConfig struct {
	ID         uint32     `json:"id"`
	TotalLevel uint32     `json:"total_level"`
	ShowReward [][]uint32 `json:"show_reward"`
}

type loveLetterLegacyConfig struct {
	ID          uint32 `json:"id"`
	ShipGroupID uint32 `json:"ship_group_id"`
	Year        uint32 `json:"year"`
}

type loveLetterConfigBundle struct {
	Characters        map[uint32]loveLetterCharacterConfig
	Contents          map[uint32]loveLetterContentConfig
	Rewards           map[uint32]loveLetterRewardConfig
	GroupLetterIDs    map[uint32][]uint32
	ItemGroupToYears  map[string]map[uint32]uint32
	LetterByGroupYear map[string]uint32
	LetterTextByID    map[uint32]string
}

type resolvedConvertedItem struct {
	Item           orm.LoveLetterConvertedItem
	CanonicalGroup uint32
	LetterID       uint32
}

type loveLetterSnapshot struct {
	ConvertedItems   []orm.LoveLetterConvertedItem
	RewardedIDs      []uint32
	Medals           []orm.LoveLetterMedalState
	Letters          []orm.LoveLetterLetterState
	ConvertedLetters []orm.LoveLetterLetterState
}

func parseLoveLetterUnlockLetterID(payload []byte) (uint32, error) {
	req := &protobuf.CS_12400{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return 0, err
	}
	if req.Id == nil {
		return 0, fmt.Errorf("missing required field id")
	}
	return req.GetId(), nil
}

func parseLoveLetterRewardIDs(payload []byte) ([]uint32, error) {
	req := &protobuf.CS_12402{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return nil, err
	}
	return append([]uint32{}, req.GetIdList()...), nil
}

func parseLoveLetterConvertedItems(payload []byte) ([]orm.LoveLetterConvertedItem, error) {
	req := &protobuf.CS_12404{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return nil, err
	}
	items := make([]orm.LoveLetterConvertedItem, 0, len(req.GetItemList()))
	for _, item := range req.GetItemList() {
		if item == nil || item.ItemId == nil || item.GroupId == nil || item.Year == nil {
			return nil, fmt.Errorf("missing required converted item fields")
		}
		items = append(items, orm.LoveLetterConvertedItem{
			ItemID:  item.GetItemId(),
			GroupID: item.GetGroupId(),
			Year:    item.GetYear(),
		})
	}
	return items, nil
}

func parseLoveLetterRequestType(payload []byte) (uint32, error) {
	req := &protobuf.CS_12406{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return 0, err
	}
	if req.Type == nil {
		return 0, fmt.Errorf("missing required field type")
	}
	return req.GetType(), nil
}

func parseLoveLetterLevelUpGroupID(payload []byte) (uint32, error) {
	req := &protobuf.CS_12408{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return 0, err
	}
	if req.GroupId == nil {
		return 0, fmt.Errorf("missing required field group_id")
	}
	return req.GetGroupId(), nil
}

func parseLoveLetterContentLetterID(payload []byte) (uint32, error) {
	req := &protobuf.CS_12410{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return 0, err
	}
	if req.LetterId == nil {
		return 0, fmt.Errorf("missing required field letter_id")
	}
	return req.GetLetterId(), nil
}

func buildLoveLetterUnlockResponse(result uint32) *protobuf.SC_12401 {
	return &protobuf.SC_12401{Result: proto.Uint32(result)}
}

func buildLoveLetterClaimRewardsResponse(result uint32, drops []*protobuf.DROPINFO) *protobuf.SC_12403 {
	return &protobuf.SC_12403{Result: proto.Uint32(result), DropList: drops}
}

func buildLoveLetterRealizeGiftResponse(result uint32) *protobuf.SC_12405 {
	return &protobuf.SC_12405{Result: proto.Uint32(result)}
}

func buildLoveLetterGetAllDataResponse(snapshot loveLetterSnapshot) *protobuf.SC_12407 {
	converted := make([]*protobuf.PT_OLD_LOVER_ITEM, 0, len(snapshot.ConvertedItems))
	for _, item := range snapshot.ConvertedItems {
		converted = append(converted, &protobuf.PT_OLD_LOVER_ITEM{
			ItemId:  proto.Uint32(item.ItemID),
			GroupId: proto.Uint32(item.GroupID),
			Year:    proto.Uint32(item.Year),
		})
	}
	medals := make([]*protobuf.PT_LOVE_LETTER_MEDAL, 0, len(snapshot.Medals))
	for _, medal := range snapshot.Medals {
		medals = append(medals, &protobuf.PT_LOVE_LETTER_MEDAL{
			GroupId: proto.Uint32(medal.GroupID),
			Exp:     proto.Uint32(medal.Exp),
			Level:   proto.Uint32(medal.Level),
		})
	}
	letters := make([]*protobuf.PT_SHIP_LOVE_LETTER, 0, len(snapshot.Letters))
	for _, letter := range snapshot.Letters {
		letters = append(letters, &protobuf.PT_SHIP_LOVE_LETTER{
			GroupId:      proto.Uint32(letter.GroupID),
			LetterIdList: append([]uint32{}, letter.LetterIDList...),
		})
	}
	convertedLetters := make([]*protobuf.PT_SHIP_LOVE_LETTER, 0, len(snapshot.ConvertedLetters))
	for _, letter := range snapshot.ConvertedLetters {
		convertedLetters = append(convertedLetters, &protobuf.PT_SHIP_LOVE_LETTER{
			GroupId:      proto.Uint32(letter.GroupID),
			LetterIdList: append([]uint32{}, letter.LetterIDList...),
		})
	}
	return &protobuf.SC_12407{
		ConvertedList:       converted,
		RewardedList:        append([]uint32{}, snapshot.RewardedIDs...),
		MedalList:           medals,
		LetterList:          letters,
		ConvertedLetterList: convertedLetters,
	}
}

func buildLoveLetterLevelUpResponse(result uint32) *protobuf.SC_12409 {
	return &protobuf.SC_12409{Ret: proto.Uint32(result)}
}

func buildLoveLetterContentResponse(content string) *protobuf.SC_12411 {
	return &protobuf.SC_12411{Content: proto.String(content)}
}

func LoveLetterGetAllData(buffer *[]byte, client *connection.Client) (int, int, error) {
	if _, err := parseLoveLetterRequestType(*buffer); err != nil {
		return 0, 12407, err
	}
	bundle, err := loadLoveLetterConfigBundle()
	if err != nil {
		return connection.SendProtoMessage(12407, client, buildLoveLetterGetAllDataResponse(loveLetterSnapshot{}))
	}
	state, err := orm.GetOrCreateCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		return connection.SendProtoMessage(12407, client, buildLoveLetterGetAllDataResponse(loveLetterSnapshot{}))
	}
	snapshot := buildLoveLetterSnapshot(state, bundle)
	return connection.SendProtoMessage(12407, client, buildLoveLetterGetAllDataResponse(snapshot))
}

func LoveLetterUnlock(buffer *[]byte, client *connection.Client) (int, int, error) {
	letterID, err := parseLoveLetterUnlockLetterID(*buffer)
	if err != nil {
		return 0, 12401, err
	}
	bundle, err := loadLoveLetterConfigBundle()
	if err != nil {
		return connection.SendProtoMessage(12401, client, buildLoveLetterUnlockResponse(loveLetterResultFailed))
	}
	letterConfig, ok := bundle.Contents[letterID]
	if !ok {
		return connection.SendProtoMessage(12401, client, buildLoveLetterUnlockResponse(loveLetterResultFailed))
	}
	state, err := orm.GetOrCreateCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		return connection.SendProtoMessage(12401, client, buildLoveLetterUnlockResponse(loveLetterResultFailed))
	}
	medalMap := medalsToMap(state.Medals)
	manualSet := letterStatesToSet(state.ManualLetters)
	giftSet := convertedLetterSet(resolveConvertedItemsLenient(state.ConvertedItems, bundle))
	merged := mergeLetterSets(manualSet, giftSet)
	if merged[letterConfig.ShipGroup] != nil {
		if _, exists := merged[letterConfig.ShipGroup][letterID]; exists {
			return connection.SendProtoMessage(12401, client, buildLoveLetterUnlockResponse(loveLetterResultFailed))
		}
	}
	medal, ok := medalMap[letterConfig.ShipGroup]
	if !ok {
		medal = &orm.LoveLetterMedalState{GroupID: letterConfig.ShipGroup}
		medalMap[letterConfig.ShipGroup] = medal
	}
	groupLetters := bundle.GroupLetterIDs[letterConfig.ShipGroup]
	index := 0
	for i := range groupLetters {
		if groupLetters[i] == letterID {
			index = i + 1
			break
		}
	}
	if index == 0 || uint32(index) > medal.Level {
		return connection.SendProtoMessage(12401, client, buildLoveLetterUnlockResponse(loveLetterResultFailed))
	}
	if manualSet[letterConfig.ShipGroup] == nil {
		manualSet[letterConfig.ShipGroup] = make(map[uint32]struct{})
	}
	manualSet[letterConfig.ShipGroup][letterID] = struct{}{}
	state.ManualLetters = letterSetToStates(manualSet)
	state.Medals = medalMapToList(medalMap)
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		return connection.SendProtoMessage(12401, client, buildLoveLetterUnlockResponse(loveLetterResultFailed))
	}
	return connection.SendProtoMessage(12401, client, buildLoveLetterUnlockResponse(loveLetterResultSuccess))
}

func LoveLetterClaimRewards(buffer *[]byte, client *connection.Client) (int, int, error) {
	rewardIDs, err := parseLoveLetterRewardIDs(*buffer)
	if err != nil {
		return 0, 12403, err
	}
	if len(rewardIDs) == 0 {
		return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
	}
	if err := ensureCommanderLoaded(client, "LoveLetter/Rewards"); err != nil {
		return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
	}
	bundle, err := loadLoveLetterConfigBundle()
	if err != nil {
		return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
	}
	state, err := orm.GetOrCreateCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
	}
	levelAll := totalDisplayLevel(state.Medals, bundle)
	rewardedSet := make(map[uint32]struct{}, len(state.RewardedIDs))
	for _, rewardID := range state.RewardedIDs {
		rewardedSet[rewardID] = struct{}{}
	}
	requestSet := make(map[uint32]struct{}, len(rewardIDs))
	drops := make(map[string]*protobuf.DROPINFO)
	for _, rewardID := range rewardIDs {
		if _, seen := requestSet[rewardID]; seen {
			return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
		}
		requestSet[rewardID] = struct{}{}
		if _, claimed := rewardedSet[rewardID]; claimed {
			return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
		}
		rewardConfig, ok := bundle.Rewards[rewardID]
		if !ok || levelAll < rewardConfig.TotalLevel {
			return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
		}
		for _, drop := range rewardConfig.ShowReward {
			if len(drop) < 3 {
				return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
			}
			accumulateDrop(drops, drop[0], drop[1], drop[2])
		}
	}
	ctx := context.Background()
	err = db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if err := applyLoveLetterDropsTx(ctx, tx, client, drops); err != nil {
			return err
		}
		state.RewardedIDs = append(state.RewardedIDs, rewardIDs...)
		sort.Slice(state.RewardedIDs, func(i int, j int) bool {
			return state.RewardedIDs[i] < state.RewardedIDs[j]
		})
		return orm.SaveCommanderLoveLetterStateTx(ctx, tx, state)
	})
	if err != nil {
		return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultFailed, []*protobuf.DROPINFO{}))
	}
	return connection.SendProtoMessage(12403, client, buildLoveLetterClaimRewardsResponse(loveLetterResultSuccess, dropMapToSortedList(drops)))
}

func LoveLetterRealizeGift(buffer *[]byte, client *connection.Client) (int, int, error) {
	convertedItems, err := parseLoveLetterConvertedItems(*buffer)
	if err != nil {
		return 0, 12405, err
	}
	bundle, err := loadLoveLetterConfigBundle()
	if err != nil {
		return connection.SendProtoMessage(12405, client, buildLoveLetterRealizeGiftResponse(loveLetterResultFailed))
	}
	resolvedNew, err := resolveConvertedItemsStrict(convertedItems, bundle)
	if err != nil {
		return connection.SendProtoMessage(12405, client, buildLoveLetterRealizeGiftResponse(loveLetterResultFailed))
	}
	state, err := orm.GetOrCreateCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		return connection.SendProtoMessage(12405, client, buildLoveLetterRealizeGiftResponse(loveLetterResultFailed))
	}
	resolvedOld := resolveConvertedItemsLenient(state.ConvertedItems, bundle)
	oldCounts := convertedCountByGroup(resolvedOld)
	newCounts := convertedCountByGroup(resolvedNew)
	medalMap := medalsToMap(state.Medals)
	for groupID, newCount := range newCounts {
		oldCount := oldCounts[groupID]
		if newCount == oldCount {
			continue
		}
		characterConfig, ok := bundle.Characters[groupID]
		if !ok || characterConfig.ExpUp == 0 {
			return connection.SendProtoMessage(12405, client, buildLoveLetterRealizeGiftResponse(loveLetterResultFailed))
		}
		medal, exists := medalMap[groupID]
		if !exists {
			medal = &orm.LoveLetterMedalState{GroupID: groupID}
			medalMap[groupID] = medal
		}
		if newCount > oldCount {
			delta := newCount - oldCount
			medal.Exp += delta * characterConfig.ExpUp
			medal.Level += delta
			continue
		}
		delta := oldCount - newCount
		expDelta := delta * characterConfig.ExpUp
		if medal.Exp > expDelta {
			medal.Exp -= expDelta
		} else {
			medal.Exp = 0
		}
		if medal.Level > delta {
			medal.Level -= delta
		} else {
			medal.Level = 0
		}
	}
	for groupID, oldCount := range oldCounts {
		if _, stillPresent := newCounts[groupID]; stillPresent {
			continue
		}
		characterConfig, ok := bundle.Characters[groupID]
		if !ok || characterConfig.ExpUp == 0 {
			return connection.SendProtoMessage(12405, client, buildLoveLetterRealizeGiftResponse(loveLetterResultFailed))
		}
		medal, exists := medalMap[groupID]
		if !exists {
			continue
		}
		expDelta := oldCount * characterConfig.ExpUp
		if medal.Exp > expDelta {
			medal.Exp -= expDelta
		} else {
			medal.Exp = 0
		}
		if medal.Level > oldCount {
			medal.Level -= oldCount
		} else {
			medal.Level = 0
		}
	}
	state.ConvertedItems = convertedItems
	state.Medals = medalMapToList(medalMap)
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		return connection.SendProtoMessage(12405, client, buildLoveLetterRealizeGiftResponse(loveLetterResultFailed))
	}
	return connection.SendProtoMessage(12405, client, buildLoveLetterRealizeGiftResponse(loveLetterResultSuccess))
}

func LoveLetterLevelUp(buffer *[]byte, client *connection.Client) (int, int, error) {
	groupID, err := parseLoveLetterLevelUpGroupID(*buffer)
	if err != nil {
		return 0, 12409, err
	}
	bundle, err := loadLoveLetterConfigBundle()
	if err != nil {
		return connection.SendProtoMessage(12409, client, buildLoveLetterLevelUpResponse(loveLetterResultFailed))
	}
	characterConfig, ok := bundle.Characters[groupID]
	if !ok || characterConfig.ExpUp == 0 {
		return connection.SendProtoMessage(12409, client, buildLoveLetterLevelUpResponse(loveLetterResultFailed))
	}
	state, err := orm.GetOrCreateCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		return connection.SendProtoMessage(12409, client, buildLoveLetterLevelUpResponse(loveLetterResultFailed))
	}
	medalMap := medalsToMap(state.Medals)
	medal, exists := medalMap[groupID]
	if !exists {
		medal = &orm.LoveLetterMedalState{GroupID: groupID}
		medalMap[groupID] = medal
	}
	threshold := (medal.Level + 1) * characterConfig.ExpUp
	if medal.Exp < threshold {
		return connection.SendProtoMessage(12409, client, buildLoveLetterLevelUpResponse(loveLetterResultFailed))
	}
	medal.Level = medal.Exp / characterConfig.ExpUp
	state.Medals = medalMapToList(medalMap)
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		return connection.SendProtoMessage(12409, client, buildLoveLetterLevelUpResponse(loveLetterResultFailed))
	}
	return connection.SendProtoMessage(12409, client, buildLoveLetterLevelUpResponse(loveLetterResultSuccess))
}

func LoveLetterGetContent(buffer *[]byte, client *connection.Client) (int, int, error) {
	letterID, err := parseLoveLetterContentLetterID(*buffer)
	if err != nil {
		return 0, 12411, err
	}
	bundle, err := loadLoveLetterConfigBundle()
	if err != nil {
		return connection.SendProtoMessage(12411, client, buildLoveLetterContentResponse(""))
	}
	state, err := orm.GetOrCreateCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		return connection.SendProtoMessage(12411, client, buildLoveLetterContentResponse(""))
	}
	if content, ok := state.LetterContents[letterID]; ok {
		return connection.SendProtoMessage(12411, client, buildLoveLetterContentResponse(content))
	}
	if content, ok := bundle.LetterTextByID[letterID]; ok {
		return connection.SendProtoMessage(12411, client, buildLoveLetterContentResponse(content))
	}
	return connection.SendProtoMessage(12411, client, buildLoveLetterContentResponse(""))
}

func buildLoveLetterSnapshot(state *orm.CommanderLoveLetterState, bundle *loveLetterConfigBundle) loveLetterSnapshot {
	resolved := resolveConvertedItemsLenient(state.ConvertedItems, bundle)
	manualSet := letterStatesToSet(state.ManualLetters)
	giftSet := convertedLetterSet(resolved)
	merged := mergeLetterSets(manualSet, giftSet)
	return loveLetterSnapshot{
		ConvertedItems:   append([]orm.LoveLetterConvertedItem{}, state.ConvertedItems...),
		RewardedIDs:      append([]uint32{}, state.RewardedIDs...),
		Medals:           append([]orm.LoveLetterMedalState{}, state.Medals...),
		Letters:          letterSetToStates(merged),
		ConvertedLetters: letterSetToStates(giftSet),
	}
}

func loadLoveLetterConfigBundle() (*loveLetterConfigBundle, error) {
	characters, err := loadLoveLetterCharacterConfigs()
	if err != nil {
		return nil, err
	}
	contents, err := loadLoveLetterContentConfigs()
	if err != nil {
		return nil, err
	}
	rewards, err := loadLoveLetterRewardConfigs()
	if err != nil {
		return nil, err
	}
	legacyMappings, err := loadLoveLetterLegacyConfigs()
	if err != nil {
		return nil, err
	}
	bundle := &loveLetterConfigBundle{
		Characters:        characters,
		Contents:          contents,
		Rewards:           rewards,
		GroupLetterIDs:    make(map[uint32][]uint32),
		ItemGroupToYears:  make(map[string]map[uint32]uint32),
		LetterByGroupYear: make(map[string]uint32),
		LetterTextByID:    make(map[uint32]string),
	}
	for _, contentConfig := range contents {
		bundle.GroupLetterIDs[contentConfig.ShipGroup] = append(bundle.GroupLetterIDs[contentConfig.ShipGroup], contentConfig.ID)
		bundle.LetterByGroupYear[groupYearKey(contentConfig.ShipGroup, contentConfig.Year)] = contentConfig.ID
		if contentConfig.Content != "" {
			bundle.LetterTextByID[contentConfig.ID] = contentConfig.Content
		}
		groups := []uint32{contentConfig.ShipGroup}
		if characterConfig, ok := characters[contentConfig.ShipGroup]; ok {
			groups = append(groups, characterConfig.RelateGroupID...)
		}
		for _, itemID := range contentConfig.LoveItem {
			for _, groupID := range groups {
				key := itemGroupKey(itemID, groupID)
				if bundle.ItemGroupToYears[key] == nil {
					bundle.ItemGroupToYears[key] = make(map[uint32]uint32)
				}
				bundle.ItemGroupToYears[key][contentConfig.Year] = contentConfig.ShipGroup
			}
		}
	}
	for groupID := range bundle.GroupLetterIDs {
		ids := bundle.GroupLetterIDs[groupID]
		sort.Slice(ids, func(i int, j int) bool {
			return ids[i] < ids[j]
		})
		bundle.GroupLetterIDs[groupID] = ids
	}
	for _, legacyConfig := range legacyMappings {
		key := itemGroupKey(legacyConfig.ID, legacyConfig.ShipGroupID)
		if bundle.ItemGroupToYears[key] == nil {
			bundle.ItemGroupToYears[key] = make(map[uint32]uint32)
		}
		if _, exists := bundle.ItemGroupToYears[key][legacyConfig.Year]; !exists {
			bundle.ItemGroupToYears[key][legacyConfig.Year] = legacyConfig.ShipGroupID
		}
	}
	if textEntries, err := orm.ListConfigEntries(loveLetterTextTemplateCategory); err == nil {
		for _, entry := range textEntries {
			if !isJSONMap(entry.Data) {
				continue
			}
			var payload map[string]any
			if err := json.Unmarshal(entry.Data, &payload); err != nil {
				return nil, err
			}
			if textValue, ok := payload["content"].(string); ok && textValue != "" {
				letterID, parseErr := strconv.ParseUint(entry.Key, 10, 32)
				if parseErr != nil {
					continue
				}
				bundle.LetterTextByID[uint32(letterID)] = textValue
			}
		}
	}
	return bundle, nil
}

func loadLoveLetterCharacterConfigs() (map[uint32]loveLetterCharacterConfig, error) {
	entries, err := orm.ListConfigEntries(loveLetterCharacterTemplateCategory)
	if err != nil {
		return nil, err
	}
	configs := make(map[uint32]loveLetterCharacterConfig)
	for _, entry := range entries {
		if !isJSONMap(entry.Data) {
			continue
		}
		var config loveLetterCharacterConfig
		if err := json.Unmarshal(entry.Data, &config); err != nil {
			return nil, err
		}
		if config.ID == 0 {
			continue
		}
		configs[config.ID] = config
	}
	return configs, nil
}

func loadLoveLetterContentConfigs() (map[uint32]loveLetterContentConfig, error) {
	entries, err := orm.ListConfigEntries(loveLetterContentTemplateCategory)
	if err != nil {
		return nil, err
	}
	configs := make(map[uint32]loveLetterContentConfig)
	for _, entry := range entries {
		if !isJSONMap(entry.Data) {
			continue
		}
		var config loveLetterContentConfig
		if err := json.Unmarshal(entry.Data, &config); err != nil {
			return nil, err
		}
		if config.ID == 0 || config.ShipGroup == 0 || config.Year == 0 {
			continue
		}
		configs[config.ID] = config
	}
	return configs, nil
}

func loadLoveLetterRewardConfigs() (map[uint32]loveLetterRewardConfig, error) {
	entries, err := orm.ListConfigEntries(loveLetterRewardTemplateCategory)
	if err != nil {
		return nil, err
	}
	configs := make(map[uint32]loveLetterRewardConfig)
	for _, entry := range entries {
		if !isJSONMap(entry.Data) {
			continue
		}
		var config loveLetterRewardConfig
		if err := json.Unmarshal(entry.Data, &config); err != nil {
			return nil, err
		}
		if config.ID == 0 {
			continue
		}
		configs[config.ID] = config
	}
	return configs, nil
}

func loadLoveLetterLegacyConfigs() (map[uint32]loveLetterLegacyConfig, error) {
	entries, err := orm.ListConfigEntries(loveLetterLegacyTemplateCategory)
	if err != nil {
		return nil, err
	}
	configs := make(map[uint32]loveLetterLegacyConfig)
	for _, entry := range entries {
		if !isJSONMap(entry.Data) {
			continue
		}
		var config loveLetterLegacyConfig
		if err := json.Unmarshal(entry.Data, &config); err != nil {
			return nil, err
		}
		if config.ID == 0 || config.ShipGroupID == 0 || config.Year == 0 {
			continue
		}
		configs[config.ID] = config
	}
	return configs, nil
}

func resolveConvertedItemsStrict(items []orm.LoveLetterConvertedItem, bundle *loveLetterConfigBundle) ([]resolvedConvertedItem, error) {
	resolved := make([]resolvedConvertedItem, 0, len(items))
	for _, item := range items {
		canonicalGroup, letterID, ok := resolveConvertedItem(item, bundle)
		if !ok {
			return nil, fmt.Errorf("invalid converted item: %d/%d/%d", item.ItemID, item.GroupID, item.Year)
		}
		resolved = append(resolved, resolvedConvertedItem{
			Item:           item,
			CanonicalGroup: canonicalGroup,
			LetterID:       letterID,
		})
	}
	return resolved, nil
}

func resolveConvertedItemsLenient(items []orm.LoveLetterConvertedItem, bundle *loveLetterConfigBundle) []resolvedConvertedItem {
	resolved := make([]resolvedConvertedItem, 0, len(items))
	for _, item := range items {
		canonicalGroup, letterID, ok := resolveConvertedItem(item, bundle)
		if !ok {
			continue
		}
		resolved = append(resolved, resolvedConvertedItem{
			Item:           item,
			CanonicalGroup: canonicalGroup,
			LetterID:       letterID,
		})
	}
	return resolved
}

func resolveConvertedItem(item orm.LoveLetterConvertedItem, bundle *loveLetterConfigBundle) (uint32, uint32, bool) {
	if item.ItemID == 0 || item.GroupID == 0 || item.Year == 0 {
		return 0, 0, false
	}
	yearMap := bundle.ItemGroupToYears[itemGroupKey(item.ItemID, item.GroupID)]
	if yearMap == nil {
		return 0, 0, false
	}
	canonicalGroup, ok := yearMap[item.Year]
	if !ok {
		return 0, 0, false
	}
	letterID := bundle.LetterByGroupYear[groupYearKey(canonicalGroup, item.Year)]
	if letterID == 0 {
		return 0, 0, false
	}
	return canonicalGroup, letterID, true
}

func convertedCountByGroup(items []resolvedConvertedItem) map[uint32]uint32 {
	counts := make(map[uint32]uint32)
	for _, item := range items {
		counts[item.CanonicalGroup]++
	}
	return counts
}

func convertedLetterSet(items []resolvedConvertedItem) map[uint32]map[uint32]struct{} {
	letters := make(map[uint32]map[uint32]struct{})
	for _, item := range items {
		if letters[item.CanonicalGroup] == nil {
			letters[item.CanonicalGroup] = make(map[uint32]struct{})
		}
		letters[item.CanonicalGroup][item.LetterID] = struct{}{}
	}
	return letters
}

func medalsToMap(medals []orm.LoveLetterMedalState) map[uint32]*orm.LoveLetterMedalState {
	medalMap := make(map[uint32]*orm.LoveLetterMedalState, len(medals))
	for i := range medals {
		medal := medals[i]
		copyMedal := medal
		medalMap[medal.GroupID] = &copyMedal
	}
	return medalMap
}

func medalMapToList(medals map[uint32]*orm.LoveLetterMedalState) []orm.LoveLetterMedalState {
	groupIDs := make([]uint32, 0, len(medals))
	for groupID := range medals {
		groupIDs = append(groupIDs, groupID)
	}
	sort.Slice(groupIDs, func(i int, j int) bool {
		return groupIDs[i] < groupIDs[j]
	})
	list := make([]orm.LoveLetterMedalState, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		list = append(list, *medals[groupID])
	}
	return list
}

func letterStatesToSet(states []orm.LoveLetterLetterState) map[uint32]map[uint32]struct{} {
	set := make(map[uint32]map[uint32]struct{})
	for _, state := range states {
		if set[state.GroupID] == nil {
			set[state.GroupID] = make(map[uint32]struct{})
		}
		for _, letterID := range state.LetterIDList {
			set[state.GroupID][letterID] = struct{}{}
		}
	}
	return set
}

func letterSetToStates(set map[uint32]map[uint32]struct{}) []orm.LoveLetterLetterState {
	groups := make([]uint32, 0, len(set))
	for groupID := range set {
		groups = append(groups, groupID)
	}
	sort.Slice(groups, func(i int, j int) bool {
		return groups[i] < groups[j]
	})
	states := make([]orm.LoveLetterLetterState, 0, len(groups))
	for _, groupID := range groups {
		letters := make([]uint32, 0, len(set[groupID]))
		for letterID := range set[groupID] {
			letters = append(letters, letterID)
		}
		sort.Slice(letters, func(i int, j int) bool {
			return letters[i] < letters[j]
		})
		states = append(states, orm.LoveLetterLetterState{GroupID: groupID, LetterIDList: letters})
	}
	return states
}

func mergeLetterSets(
	a map[uint32]map[uint32]struct{},
	b map[uint32]map[uint32]struct{},
) map[uint32]map[uint32]struct{} {
	merged := make(map[uint32]map[uint32]struct{}, len(a)+len(b))
	for groupID, letters := range a {
		merged[groupID] = make(map[uint32]struct{}, len(letters))
		for letterID := range letters {
			merged[groupID][letterID] = struct{}{}
		}
	}
	for groupID, letters := range b {
		if merged[groupID] == nil {
			merged[groupID] = make(map[uint32]struct{}, len(letters))
		}
		for letterID := range letters {
			merged[groupID][letterID] = struct{}{}
		}
	}
	return merged
}

func totalDisplayLevel(medals []orm.LoveLetterMedalState, bundle *loveLetterConfigBundle) uint32 {
	var total uint32
	for _, medal := range medals {
		displayLevel := medal.Level
		characterConfig, ok := bundle.Characters[medal.GroupID]
		if ok && characterConfig.ExpUp > 0 && characterConfig.ExpUpperLimit > 0 {
			maxLevel := characterConfig.ExpUpperLimit / characterConfig.ExpUp
			if displayLevel > maxLevel {
				displayLevel = maxLevel
			}
		}
		total += displayLevel
	}
	return total
}

func applyLoveLetterDropsTx(ctx context.Context, tx pgx.Tx, client *connection.Client, drops map[string]*protobuf.DROPINFO) error {
	for _, drop := range drops {
		dropType := drop.GetType()
		dropID := drop.GetId()
		dropCount := drop.GetNumber()
		switch dropType {
		case consts.DROP_TYPE_RESOURCE:
			if err := client.Commander.AddResourceTx(ctx, tx, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_ITEM, consts.DROP_TYPE_LOVE_LETTER:
			if err := client.Commander.AddItemTx(ctx, tx, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_EQUIP:
			if err := addOwnedEquipmentPGXTx(ctx, tx, client.Commander, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_SHIP:
			for i := uint32(0); i < dropCount; i++ {
				if _, err := client.Commander.AddShipTx(ctx, tx, dropID); err != nil {
					return err
				}
			}
		case consts.DROP_TYPE_FURNITURE:
			now := uint32(time.Now().Unix())
			if err := orm.AddCommanderFurnitureTx(ctx, tx, client.Commander.CommanderID, dropID, dropCount, now); err != nil {
				return err
			}
		case consts.DROP_TYPE_SKIN:
			for i := uint32(0); i < dropCount; i++ {
				if err := client.Commander.GiveSkinTx(ctx, tx, dropID); err != nil {
					return err
				}
			}
		case consts.DROP_TYPE_VITEM:
			continue
		default:
			return fmt.Errorf("unsupported reward drop type %d", dropType)
		}
	}
	return nil
}

func accumulateDrop(drops map[string]*protobuf.DROPINFO, dropType uint32, dropID uint32, count uint32) {
	key := fmt.Sprintf("%d_%d", dropType, dropID)
	entry := drops[key]
	if entry == nil {
		drops[key] = &protobuf.DROPINFO{
			Type:   proto.Uint32(dropType),
			Id:     proto.Uint32(dropID),
			Number: proto.Uint32(count),
		}
		return
	}
	entry.Number = proto.Uint32(entry.GetNumber() + count)
}

func dropMapToSortedList(drops map[string]*protobuf.DROPINFO) []*protobuf.DROPINFO {
	list := make([]*protobuf.DROPINFO, 0, len(drops))
	for _, drop := range drops {
		list = append(list, drop)
	}
	sort.Slice(list, func(i int, j int) bool {
		if list[i].GetType() == list[j].GetType() {
			return list[i].GetId() < list[j].GetId()
		}
		return list[i].GetType() < list[j].GetType()
	})
	return list
}

func itemGroupKey(itemID uint32, groupID uint32) string {
	return fmt.Sprintf("%d_%d", itemID, groupID)
}

func groupYearKey(groupID uint32, year uint32) string {
	return fmt.Sprintf("%d_%d", groupID, year)
}

func isJSONMap(data json.RawMessage) bool {
	trimmed := bytes.TrimSpace(data)
	return len(trimmed) > 0 && trimmed[0] == '{'
}
