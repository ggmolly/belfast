package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var battleSessionKey uint32

func nextBattleSessionKey() uint32 {
	return atomic.AddUint32(&battleSessionKey, 1)
}

func BeginStage(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_40001
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 40002, err
	}
	key := nextBattleSessionKey()
	session := orm.BattleSession{
		CommanderID: client.Commander.CommanderID,
		System:      payload.GetSystem(),
		StageID:     payload.GetData(),
		Key:         key,
		ShipIDs:     orm.ToInt64List(payload.GetShipIdList()),
	}
	if err := orm.UpsertBattleSession(orm.GormDB, &session); err != nil {
		return 0, 40002, err
	}
	response := protobuf.SC_40002{
		Result:          proto.Uint32(0),
		Key:             proto.Uint32(key),
		DropPerformance: []*protobuf.DROPPERFORMANCE{},
	}
	return client.SendMessage(40002, &response)
}

func FinishStage(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_40003
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 40004, err
	}
	session, err := orm.GetBattleSession(orm.GormDB, client.Commander.CommanderID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 40004, err
	}
	shipIDs := []uint32{}
	if session != nil {
		shipIDs = orm.ToUint32List(session.ShipIDs)
	}
	shipExpList := make([]*protobuf.SHIP_EXP, 0, len(shipIDs))
	for _, shipID := range shipIDs {
		shipExpList = append(shipExpList, &protobuf.SHIP_EXP{
			ShipId: proto.Uint32(shipID),
			// todo: compute ship exp/intimacy/energy from battle statistics
			Exp:      proto.Uint32(0),
			Intimacy: proto.Uint32(0),
			Energy:   proto.Uint32(0),
		})
	}
	mvp := uint32(0)
	if len(shipIDs) > 0 {
		mvp = shipIDs[0]
	}
	dropList := []*protobuf.DROPINFO{}
	if session != nil {
		update, err := updateChapterStateAfterBattle(client.Commander.CommanderID, session.StageID)
		if err != nil {
			return 0, 40004, err
		}
		if update != nil {
			if err := updateChapterProgressAfterBattle(client.Commander.CommanderID, update, payload.GetScore()); err != nil {
				return 0, 40004, err
			}
		}
		if update != nil && update.defeated {
			drops, err := buildChapterAwardDrops(update.template)
			if err != nil {
				return 0, 40004, err
			}
			if len(drops) > 0 {
				if err := applyDropList(client, drops); err != nil {
					return 0, 40004, err
				}
				dropList = dropMapToList(drops)
			}
		}
	}
	if err := orm.DeleteBattleSession(orm.GormDB, client.Commander.CommanderID); err != nil {
		return 0, 40004, err
	}
	response := protobuf.SC_40004{
		Result: proto.Uint32(0),
		// todo: compute player exp from battle statistics
		DropInfo:      dropList,
		ExtraDropInfo: []*protobuf.DROPINFO{},
		PlayerExp:     proto.Uint32(0),
		ShipExpList:   shipExpList,
		Mvp:           proto.Uint32(mvp),
	}
	return client.SendMessage(40004, &response)
}

func buildChapterAwardDrops(template *chapterTemplate) (map[string]*protobuf.DROPINFO, error) {
	drops := make(map[string]*protobuf.DROPINFO)
	if template == nil || len(template.Awards) == 0 {
		return drops, nil
	}
	for _, entry := range template.Awards {
		if len(entry) < 2 {
			continue
		}
		dropType := entry[0]
		for i := 1; i < len(entry); i++ {
			dropID := entry[i]
			if dropID == 0 {
				continue
			}
			resolvedType, resolvedID, resolvedCount, err := resolveChapterAwardDrop(dropType, dropID)
			if err != nil {
				return nil, err
			}
			key := fmt.Sprintf("%d_%d", resolvedType, resolvedID)
			if existing, ok := drops[key]; ok {
				existing.Number = proto.Uint32(existing.GetNumber() + resolvedCount)
				continue
			}
			drops[key] = newDropInfo(resolvedType, resolvedID, resolvedCount)
		}
	}
	return drops, nil
}

func resolveChapterAwardDrop(dropType uint32, dropID uint32) (uint32, uint32, uint32, error) {
	if dropType != consts.DROP_TYPE_ITEM {
		return dropType, dropID, 1, nil
	}
	config, err := loadVirtualItemConfig(dropID)
	if err != nil {
		return 0, 0, 0, err
	}
	if config == nil || len(config.DisplayIcon) == 0 {
		return dropType, dropID, 1, nil
	}
	// TODO: honor loot odds/weights instead of uniform selection.
	entry := config.DisplayIcon[randomIndex(len(config.DisplayIcon))]
	if len(entry) < 2 {
		return dropType, dropID, 1, nil
	}
	count := uint32(1)
	if len(entry) > 2 && entry[2] > 0 {
		count = entry[2]
	}
	return entry[0], entry[1], count, nil
}

func loadVirtualItemConfig(itemID uint32) (*virtualItemConfig, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, "sharecfgdata/item_virtual_data_statistics.json", fmt.Sprintf("%d", itemID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var config virtualItemConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

type chapterBattleUpdate struct {
	current            *protobuf.CURRENTCHAPTERINFO
	template           *chapterTemplate
	defeatedAttachment uint32
	defeated           bool
	expeditionID       uint32
}

type virtualItemConfig struct {
	ID          uint32     `json:"id"`
	Type        uint32     `json:"type"`
	VirtualType uint32     `json:"virtual_type"`
	DisplayIcon [][]uint32 `json:"display_icon"`
}

func updateChapterStateAfterBattle(commanderID uint32, expeditionID uint32) (*chapterBattleUpdate, error) {
	if expeditionID == 0 {
		return nil, nil
	}
	state, err := orm.GetChapterState(orm.GormDB, commanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		return nil, err
	}
	update := &chapterBattleUpdate{current: &current, expeditionID: expeditionID}
	loopFlag := current.GetLoopFlag()
	template, err := loadChapterTemplate(current.GetId(), loopFlag)
	if err != nil {
		return nil, err
	}
	update.template = template
	changed, attachment := disableChapterCellByExpedition(&current, expeditionID)
	update.defeated = changed
	update.defeatedAttachment = attachment
	if !changed {
		return update, nil
	}
	stateBytes, err := proto.Marshal(&current)
	if err != nil {
		return nil, err
	}
	state.State = stateBytes
	state.ChapterID = current.GetId()
	if err := orm.UpsertChapterState(orm.GormDB, state); err != nil {
		return nil, err
	}
	update.current = &current
	return update, nil
}

func disableChapterCellByExpedition(current *protobuf.CURRENTCHAPTERINFO, expeditionID uint32) (bool, uint32) {
	for _, cell := range current.GetCellList() {
		if cell.GetItemId() != expeditionID {
			continue
		}
		if !isEnemyAttachment(cell.GetItemType()) {
			continue
		}
		if cell.GetItemFlag() == chapterCellDisabled {
			continue
		}
		cell.ItemFlag = proto.Uint32(chapterCellDisabled)
		return true, cell.GetItemType()
	}
	return false, 0
}

func isEnemyAttachment(attachment uint32) bool {
	switch attachment {
	case chapterAttachBoss,
		chapterAttachElite,
		chapterAttachAmbush,
		chapterAttachEnemy,
		chapterAttachTorpedoEnemy,
		chapterAttachChampion,
		chapterAttachBombEnemy:
		return true
	default:
		return false
	}
}

func isEnemyCountAttachment(attachment uint32) bool {
	switch attachment {
	case chapterAttachEnemy,
		chapterAttachElite,
		chapterAttachChampion:
		return true
	default:
		return false
	}
}

func updateChapterProgressAfterBattle(commanderID uint32, update *chapterBattleUpdate, score uint32) error {
	if update.template == nil || update.current == nil {
		return nil
	}
	if !update.defeated {
		return nil
	}
	progress, err := orm.GetChapterProgress(orm.GormDB, commanderID, update.current.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			progress = &orm.ChapterProgress{CommanderID: commanderID, ChapterID: update.current.GetId()}
		} else {
			return err
		}
	}
	chapterCleared := isChapterCleared(update.current)
	bossDefeated := update.defeatedAttachment == chapterAttachBoss || containsUint32(update.template.BossExpeditionID, update.expeditionID)
	if bossDefeated {
		newProgress := minUint32(progress.Progress+update.template.ProgressBoss, 100)
		if newProgress == 100 {
			progress.PassCount++
		}
		progress.Progress = newProgress
		progress.DefeatCount++
		progress.TodayDefeatCount++
	}
	applyStarSlot(update.template.StarRequire1, update.template.Num1, chapterCleared, bossDefeated, update.defeatedAttachment, update.defeated, score, update.current.GetInitShipCount(), &progress.KillBossCount)
	applyStarSlot(update.template.StarRequire2, update.template.Num2, chapterCleared, bossDefeated, update.defeatedAttachment, update.defeated, score, update.current.GetInitShipCount(), &progress.KillEnemyCount)
	applyStarSlot(update.template.StarRequire3, update.template.Num3, chapterCleared, bossDefeated, update.defeatedAttachment, update.defeated, score, update.current.GetInitShipCount(), &progress.TakeBoxCount)
	return orm.UpsertChapterProgress(orm.GormDB, progress)
}

func applyStarSlot(starType uint32, config uint32, chapterCleared bool, bossDefeated bool, attachment uint32, defeated bool, score uint32, initShipCount uint32, count *uint32) {
	if config == 0 || count == nil {
		return
	}
	if *count >= config {
		return
	}
	switch starType {
	case 1:
		if bossDefeated && defeated {
			*count = *count + 1
		}
	case 2:
		if defeated && isEnemyCountAttachment(attachment) {
			*count = *count + 1
		}
	case 3:
		if chapterCleared {
			*count = *count + 1
		}
	case 4:
		if bossDefeated && initShipCount <= config {
			*count = *count + 1
		}
	case 6:
		if bossDefeated && score == 4 {
			*count = *count + 1
		}
	}
}

func isChapterCleared(current *protobuf.CURRENTCHAPTERINFO) bool {
	for _, cell := range current.GetCellList() {
		if !isEnemyAttachment(cell.GetItemType()) {
			continue
		}
		if cell.GetItemFlag() != chapterCellDisabled {
			return false
		}
	}
	return true
}

func minUint32(value uint32, max uint32) uint32 {
	if value > max {
		return max
	}
	return value
}

func QuitBattle(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_40005
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 40006, err
	}
	if err := orm.DeleteBattleSession(orm.GormDB, client.Commander.CommanderID); err != nil {
		return 0, 40006, err
	}
	response := protobuf.SC_40006{Result: proto.Uint32(0)}
	return client.SendMessage(40006, &response)
}

func DailyQuickBattle(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_40007
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 40008, err
	}
	rewardCount := int(payload.GetCnt())
	rewardList := make([]*protobuf.QUICK_REWARD, 0, rewardCount)
	for i := 0; i < rewardCount; i++ {
		rewardList = append(rewardList, &protobuf.QUICK_REWARD{})
	}
	response := protobuf.SC_40008{
		Result:     proto.Uint32(0),
		RewardList: rewardList,
	}
	return client.SendMessage(40008, &response)
}
