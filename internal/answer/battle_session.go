package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync/atomic"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var battleSessionKey uint32

const (
	battleSystemScenario = 1
	battleSystemRoutine  = 2
	battleSystemDuel     = 3
	battleSystemSub      = 11
	battleSystemWorld    = 51
)

const (
	rankScoreS = 4
)

var commanderXpTableA = []uint32{0, 0, 20, 27, 35, 42, 49, 57, 65, 72}
var commanderXpTableS = []uint32{0, 0, 24, 32, 42, 51, 59, 69, 78, 87}

const (
	expBonusAmazonRate  = 0.18
	expBonusYuubariRate = 0.15
	expBonusHoushouRate = 0.15
	expBonusLangleyRate = 0.15
	expBonusArgusRate   = 0.10
	expBonusNurnRate    = 0.10
)

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
	if client.Commander.OwnedShipsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return 0, 40004, err
		}
	}
	shipIDs := []uint32{}
	if session != nil {
		shipIDs = orm.ToUint32List(session.ShipIDs)
	}
	statsByShip := buildStatisticsMap(payload.GetStatistics())
	mvp := resolveMvpShip(shipIDs, statsByShip)
	shipExpList := make([]*protobuf.SHIP_EXP, 0, len(shipIDs))
	baseExpedition, err := loadExpeditionConfig(payload.GetData())
	if err != nil {
		return 0, 40004, err
	}
	baseShipExp := uint32(0)
	if baseExpedition != nil {
		baseShipExp = baseExpedition.Exp
	}
	applyMorale := battleUsesMorale(payload.GetSystem())
	isRankS := payload.GetScore() >= rankScoreS
	shipExpGains := make(map[uint32]uint32)
	shipEnergyUpdates := make(map[uint32]uint32)
	shipIntimacyUpdates := make(map[uint32]uint32)
	if len(shipIDs) > 0 {
		fleetShips := make([]*orm.OwnedShip, 0, len(shipIDs))
		for _, shipID := range shipIDs {
			if owned, ok := client.Commander.OwnedShipsMap[shipID]; ok {
				fleetShips = append(fleetShips, owned)
			}
		}
		for _, shipID := range shipIDs {
			owned, ok := client.Commander.OwnedShipsMap[shipID]
			if !ok {
				shipExpList = append(shipExpList, &protobuf.SHIP_EXP{
					ShipId:   proto.Uint32(shipID),
					Exp:      proto.Uint32(0),
					Intimacy: proto.Uint32(10000),
					Energy:   proto.Uint32(0),
				})
				continue
			}
			shipStat := statsByShip[shipID]
			shipExpGain := computeShipExpGain(baseShipExp, owned, fleetShips, shipIDs[0], mvp, shipStat, isRankS, applyMorale)
			shipExpGains[shipID] = shipExpGain
			energyDelta := computeMoraleDelta(owned, shipStat, applyMorale)
			newEnergy := clampUint32(int(owned.Energy)+energyDelta, 0, 150)
			shipEnergyUpdates[shipID] = newEnergy
			intimacyDelta := computeIntimacyDelta(newEnergy, applyMorale)
			newIntimacy := clampUint32(int(owned.Intimacy)+intimacyDelta, 0, int(^uint32(0)))
			shipIntimacyUpdates[shipID] = newIntimacy
			energyLoss := uint32(0)
			if energyDelta < 0 {
				energyLoss = uint32(-energyDelta)
			}
			intimacyValue := uint32(10000 + intimacyDelta)
			shipExpList = append(shipExpList, &protobuf.SHIP_EXP{
				ShipId:   proto.Uint32(shipID),
				Exp:      proto.Uint32(shipExpGain),
				Intimacy: proto.Uint32(intimacyValue),
				Energy:   proto.Uint32(energyLoss),
			})
		}
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
	if err := applyBattleShipUpdates(client, shipExpGains, shipEnergyUpdates, shipIntimacyUpdates); err != nil {
		return 0, 40004, err
	}
	playerExp := uint32(0)
	if payload.GetSystem() == battleSystemScenario || payload.GetSystem() == battleSystemRoutine || payload.GetSystem() == battleSystemSub {
		playerExp = computeCommanderExpGain(len(shipIDs), isRankS)
		playerExp = applyCommanderExpReduction(playerExp, client.Commander.Level, baseExpedition, payload.GetSystem())
	}
	if err := applyCommanderExpGain(client, playerExp); err != nil {
		return 0, 40004, err
	}
	if err := orm.DeleteBattleSession(orm.GormDB, client.Commander.CommanderID); err != nil {
		return 0, 40004, err
	}
	response := protobuf.SC_40004{
		Result:        proto.Uint32(0),
		DropInfo:      dropList,
		ExtraDropInfo: []*protobuf.DROPINFO{},
		PlayerExp:     proto.Uint32(playerExp),
		ShipExpList:   shipExpList,
		Mvp:           proto.Uint32(mvp),
	}
	return client.SendMessage(40004, &response)
}

type expeditionConfig struct {
	ID    uint32 `json:"id"`
	Exp   uint32 `json:"exp"`
	Level uint32 `json:"level"`
}

type shipLevelConfig struct {
	Level uint32 `json:"level"`
	Exp   uint32 `json:"exp"`
	ExpUR uint32 `json:"exp_ur"`
}

type userLevelConfig struct {
	Level uint32 `json:"level"`
	Exp   uint32 `json:"exp"`
}

func loadExpeditionConfig(expeditionID uint32) (*expeditionConfig, error) {
	if expeditionID == 0 {
		return nil, nil
	}
	entry, err := orm.GetConfigEntry(orm.GormDB, "sharecfgdata/expedition_data_template.json", fmt.Sprintf("%d", expeditionID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var config expeditionConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func loadShipLevelConfig(level uint32) (*shipLevelConfig, error) {
	if level == 0 {
		return nil, nil
	}
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/ship_level.json", fmt.Sprintf("%d", level))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var config shipLevelConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func loadUserLevelConfig(level uint32) (*userLevelConfig, error) {
	if level == 0 {
		return nil, nil
	}
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/user_level.json", fmt.Sprintf("%d", level))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var config userLevelConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func buildStatisticsMap(stats []*protobuf.STATISTICSINFO) map[uint32]*protobuf.STATISTICSINFO {
	result := make(map[uint32]*protobuf.STATISTICSINFO)
	for _, entry := range stats {
		if entry == nil {
			continue
		}
		shipID := entry.GetShipId()
		if shipID == 0 {
			continue
		}
		result[shipID] = entry
	}
	return result
}

func resolveMvpShip(shipIDs []uint32, stats map[uint32]*protobuf.STATISTICSINFO) uint32 {
	if len(shipIDs) == 0 {
		return 0
	}
	mvp := shipIDs[0]
	maxDamage := uint32(0)
	for _, shipID := range shipIDs {
		entry, ok := stats[shipID]
		if !ok {
			continue
		}
		damage := entry.GetDamageCaused()
		if damage > maxDamage {
			maxDamage = damage
			mvp = shipID
		}
	}
	return mvp
}

func battleUsesMorale(system uint32) bool {
	switch system {
	case battleSystemDuel, battleSystemWorld:
		return false
	default:
		return true
	}
}

func computeShipExpGain(baseExp uint32, owned *orm.OwnedShip, fleetShips []*orm.OwnedShip, flagshipID uint32, mvp uint32, stats *protobuf.STATISTICSINFO, isRankS bool, applyMorale bool) uint32 {
	if baseExp == 0 || owned == nil {
		return 0
	}
	if owned.Level >= owned.MaxLevel && owned.MaxLevel < 100 {
		return 0
	}
	multiplier := 1.0
	if isRankS {
		multiplier *= 1.2
	}
	if owned.ID == flagshipID {
		multiplier *= 1.5
	}
	if owned.ID == mvp {
		multiplier *= 2.0
	}
	if applyMorale {
		if owned.Energy >= 120 {
			multiplier *= 1.2
		} else if owned.Energy == 0 {
			multiplier *= 0.5
		}
	}
	baseGain := uint32(math.Floor(float64(baseExp)*multiplier + 1e-6))
	bonusRate := computeExpSkillBonusRate(owned, fleetShips)
	bonusGain := uint32(math.Floor(float64(baseExp) * bonusRate))
	return baseGain + bonusGain
}

func computeMoraleDelta(owned *orm.OwnedShip, stats *protobuf.STATISTICSINFO, applyMorale bool) int {
	if owned == nil {
		return 0
	}
	if !applyMorale {
		return 0
	}
	if stats != nil && stats.GetHpRest() == 0 {
		return -12
	}
	return -2
}

func computeIntimacyDelta(energy uint32, applyMorale bool) int {
	if !applyMorale {
		return 100
	}
	if energy == 0 {
		return -100
	}
	if energy <= 30 {
		return 0
	}
	return 100
}

func computeExpSkillBonusRate(target *orm.OwnedShip, fleetShips []*orm.OwnedShip) float64 {
	if target == nil {
		return 0
	}
	bonusRate := 0.0
	if target.Ship.Type == 1 {
		bonusRate += expSkillBonusAmazon(fleetShips)
	}
	if target.Ship.Type == 2 || target.Ship.Type == 3 {
		bonusRate += expSkillBonusYuubari(fleetShips)
	}
	if target.Ship.Type == 6 || target.Ship.Type == 7 {
		bonusRate += expSkillBonusCarrier(fleetShips)
	}
	if target.Ship.Type == 6 {
		bonusRate += expSkillBonusArgus(fleetShips)
	}
	if target.Ship.Type == 8 || target.Ship.Type == 17 {
		bonusRate += expSkillBonusNurnberg(fleetShips)
	}
	return bonusRate
}

func expSkillBonusAmazon(fleetShips []*orm.OwnedShip) float64 {
	return expBonusIfPresent(fleetShips, "HMS Amazon", expBonusAmazonRate, false)
}

func expSkillBonusYuubari(fleetShips []*orm.OwnedShip) float64 {
	return expBonusIfPresent(fleetShips, "IJN Yūbari", expBonusYuubariRate, false)
}

func expSkillBonusCarrier(fleetShips []*orm.OwnedShip) float64 {
	bonus := expBonusIfPresent(fleetShips, "IJN Hōshō", expBonusHoushouRate, false)
	bonus += expBonusIfPresent(fleetShips, "USS Langley", expBonusLangleyRate, false)
	return bonus
}

func expSkillBonusArgus(fleetShips []*orm.OwnedShip) float64 {
	return expBonusIfPresent(fleetShips, "HMS Argus", expBonusArgusRate, true)
}

func expSkillBonusNurnberg(fleetShips []*orm.OwnedShip) float64 {
	return expBonusIfPresent(fleetShips, "KMS Nürnberg", expBonusNurnRate, false)
}

func expBonusIfPresent(fleetShips []*orm.OwnedShip, englishName string, rate float64, stackable bool) float64 {
	if rate == 0 {
		return 0
	}
	count := 0
	for _, ship := range fleetShips {
		if ship == nil {
			continue
		}
		if ship.Ship.EnglishName == englishName {
			count++
			if !stackable {
				break
			}
		}
	}
	if count == 0 {
		return 0
	}
	if stackable {
		return float64(count) * rate
	}
	return rate
}

func computeCommanderExpGain(fleetSize int, isRankS bool) uint32 {
	if fleetSize < 0 {
		return 0
	}
	if fleetSize > 9 {
		fleetSize = 9
	}
	if isRankS {
		return commanderXpTableS[fleetSize]
	}
	return commanderXpTableA[fleetSize]
}

func applyCommanderExpReduction(exp uint32, commanderLevel int, expedition *expeditionConfig, system uint32) uint32 {
	if exp == 0 || expedition == nil {
		return exp
	}
	levelGap := commanderLevel - int(expedition.Level)
	if levelGap <= 0 {
		return exp
	}
	thresholdHalf := 21
	thresholdSevere := 41
	if system == battleSystemWorld {
		thresholdHalf = 31
		thresholdSevere = 61
	}
	if levelGap >= thresholdSevere {
		return exp / 10
	}
	if levelGap >= thresholdHalf {
		return exp / 2
	}
	return exp
}

func applyCommanderExpGain(client *connection.Client, exp uint32) error {
	if exp == 0 {
		return nil
	}
	if client.Commander.Level >= 200 {
		return nil
	}
	remaining := exp
	for remaining > 0 && client.Commander.Level < 200 {
		config, err := loadUserLevelConfig(uint32(client.Commander.Level))
		if err != nil {
			return err
		}
		if config == nil || config.Exp == 0 {
			client.Commander.Exp += int(remaining)
			return orm.GormDB.Save(client.Commander).Error
		}
		needed := int(config.Exp) - client.Commander.Exp
		if needed <= 0 {
			client.Commander.Level++
			client.Commander.Exp = 0
			continue
		}
		if int(remaining) < needed {
			client.Commander.Exp += int(remaining)
			remaining = 0
			break
		}
		remaining -= uint32(needed)
		client.Commander.Level++
		client.Commander.Exp = 0
	}
	if client.Commander.Level >= 200 {
		client.Commander.Level = 200
		client.Commander.Exp = 0
	}
	return orm.GormDB.Save(client.Commander).Error
}

func applyBattleShipUpdates(client *connection.Client, expGains map[uint32]uint32, energyUpdates map[uint32]uint32, intimacyUpdates map[uint32]uint32) error {
	if len(expGains) == 0 && len(energyUpdates) == 0 && len(intimacyUpdates) == 0 {
		return nil
	}
	for shipID, gain := range expGains {
		owned, ok := client.Commander.OwnedShipsMap[shipID]
		if !ok {
			continue
		}
		if gain > 0 {
			if owned.Level >= owned.MaxLevel {
				if owned.MaxLevel >= 100 {
					owned.SurplusExp = addSurplusExp(owned.SurplusExp, gain)
					gain = 0
				}
			} else {
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
			}
		}
		if energy, ok := energyUpdates[shipID]; ok {
			owned.Energy = energy
		}
		if intimacy, ok := intimacyUpdates[shipID]; ok {
			owned.Intimacy = intimacy
		}
		if err := orm.GormDB.Save(owned).Error; err != nil {
			return err
		}
	}
	return nil
}

func addSurplusExp(current uint32, gain uint32) uint32 {
	const surplusCap = 3000000
	if current >= surplusCap {
		return current
	}
	newValue := current + gain
	if newValue > surplusCap {
		return surplusCap
	}
	return newValue
}

func clampUint32(value int, min int, max int) uint32 {
	if value < min {
		return uint32(min)
	}
	if value > max {
		return uint32(max)
	}
	return uint32(value)
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
