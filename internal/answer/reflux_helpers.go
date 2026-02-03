package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

const (
	activityTypeReflux = 42

	returnSignTemplateCategory = "ShareCfg/return_sign_template.json"
	returnPtTemplateCategory   = "ShareCfg/return_pt_template.json"
	activityTemplateCategory   = "ShareCfg/activity_template.json"
)

var (
	errReturnSignTemplateMissing = errors.New("return sign template config missing")
	errReturnPtTemplateMissing   = errors.New("return pt template config missing")
)

type refluxEligibilityConfig struct {
	MinLevel       uint32
	MinOfflineDays uint32
	MaxOfflineDays uint32
}

type returnSignTemplate struct {
	ID           uint32       `json:"id"`
	Level        [][]uint32   `json:"level"`
	AwardDisplay [][][]uint32 `json:"award_display"`
}

type returnPtTemplate struct {
	ID           uint32     `json:"id"`
	Level        [][]uint32 `json:"level"`
	AwardDisplay [][]uint32 `json:"award_display"`
	PtRequire    uint32     `json:"pt_require"`
	VirtualItem  uint32     `json:"virtual_item"`
}

func loadRefluxEligibilityConfig() (refluxEligibilityConfig, bool, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, activityTemplateCategory)
	if err != nil {
		return refluxEligibilityConfig{}, false, err
	}
	for _, entry := range entries {
		var template activityTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return refluxEligibilityConfig{}, false, err
		}
		if template.Type != activityTypeReflux {
			continue
		}
		cfg, err := parseRefluxEligibilityConfig(template.ConfigData)
		if err != nil {
			return refluxEligibilityConfig{}, false, err
		}
		return cfg, true, nil
	}
	return refluxEligibilityConfig{}, false, nil
}

func parseRefluxEligibilityConfig(raw json.RawMessage) (refluxEligibilityConfig, error) {
	var values []any
	if err := json.Unmarshal(raw, &values); err != nil {
		return refluxEligibilityConfig{}, err
	}
	if len(values) < 2 {
		return refluxEligibilityConfig{}, errors.New("reflux config_data too short")
	}
	minLevel, ok := parseJSONUint(values[0])
	if !ok {
		return refluxEligibilityConfig{}, errors.New("reflux config_data invalid min level")
	}
	minOfflineDays, ok := parseJSONUint(values[1])
	if !ok {
		return refluxEligibilityConfig{}, errors.New("reflux config_data invalid min offline days")
	}
	var maxOfflineDays uint32
	if len(values) > 2 {
		if parsed, ok := parseJSONUint(values[2]); ok {
			maxOfflineDays = parsed
		}
	}
	return refluxEligibilityConfig{
		MinLevel:       minLevel,
		MinOfflineDays: minOfflineDays,
		MaxOfflineDays: maxOfflineDays,
	}, nil
}

func loadReturnSignTemplates() (map[uint32]returnSignTemplate, []uint32, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, returnSignTemplateCategory)
	if err != nil {
		return nil, nil, err
	}
	if len(entries) == 0 {
		return nil, nil, errReturnSignTemplateMissing
	}
	lookup := make(map[uint32]returnSignTemplate, len(entries))
	ids := make([]uint32, 0, len(entries))
	for _, entry := range entries {
		var template returnSignTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return nil, nil, err
		}
		lookup[template.ID] = template
		ids = append(ids, template.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return lookup, ids, nil
}

func loadReturnPtTemplates() (map[uint32]returnPtTemplate, []uint32, uint32, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, returnPtTemplateCategory)
	if err != nil {
		return nil, nil, 0, err
	}
	if len(entries) == 0 {
		return nil, nil, 0, errReturnPtTemplateMissing
	}
	lookup := make(map[uint32]returnPtTemplate, len(entries))
	ids := make([]uint32, 0, len(entries))
	var ptItemID uint32
	for _, entry := range entries {
		var template returnPtTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return nil, nil, 0, err
		}
		lookup[template.ID] = template
		ids = append(ids, template.ID)
		if ptItemID == 0 {
			ptItemID = template.VirtualItem
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return lookup, ids, ptItemID, nil
}

func selectLevelIndex(level uint32, ranges [][]uint32) (int, error) {
	for i, entry := range ranges {
		if len(entry) < 2 {
			continue
		}
		if entry[0] <= level && level <= entry[1] {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no level range for %d", level)
}

func buildAwardDrops(display [][]uint32) ([]*protobuf.DROPINFO, error) {
	drops := make([]*protobuf.DROPINFO, 0, len(display))
	for _, entry := range display {
		if len(entry) < 3 {
			return nil, errors.New("award display entry missing fields")
		}
		drops = append(drops, newDropInfo(entry[0], entry[1], entry[2]))
	}
	return drops, nil
}

func ensureCommanderLoaded(client *connection.Client, scope string) error {
	if client.Commander.CommanderItemsMap != nil && client.Commander.MiscItemsMap != nil && client.Commander.OwnedResourcesMap != nil && client.Commander.OwnedShipsMap != nil {
		return nil
	}
	logger.LogEvent(scope, "Load", "commander maps missing, reloading commander", logger.LOG_LEVEL_INFO)
	if err := client.Commander.Load(); err != nil {
		logger.LogEvent(scope, "Load", "commander load failed", logger.LOG_LEVEL_ERROR)
		return err
	}
	return nil
}

func isSameDay(a uint32, b uint32) bool {
	if a == 0 || b == 0 {
		return false
	}
	dayA := time.Unix(int64(a), 0).UTC()
	dayB := time.Unix(int64(b), 0).UTC()
	return dayA.Year() == dayB.Year() && dayA.Month() == dayB.Month() && dayA.Day() == dayB.Day()
}

func isRefluxExpired(returnTime uint32, signDays uint32, now uint32) bool {
	if returnTime == 0 || signDays == 0 {
		return false
	}
	return returnTime+signDays*86400 <= now
}

func isRefluxEligible(client *connection.Client, cfg refluxEligibilityConfig, now time.Time) bool {
	if cfg.MinLevel > 0 && uint32(client.Commander.Level) < cfg.MinLevel {
		return false
	}
	if client.PreviousLoginAt.IsZero() {
		return false
	}
	if now.Before(client.PreviousLoginAt) {
		return false
	}
	offlineSeconds := now.Unix() - client.PreviousLoginAt.Unix()
	if offlineSeconds < 0 {
		return false
	}
	offlineDays := uint32(offlineSeconds / 86400)
	if offlineDays < cfg.MinOfflineDays {
		return false
	}
	if cfg.MaxOfflineDays > 0 && offlineDays > cfg.MaxOfflineDays {
		return false
	}
	return true
}
