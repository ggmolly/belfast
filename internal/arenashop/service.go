package arenashop

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const arenaShopConfigCategory = "ShareCfg/arena_data_shop.json"

type shopTemplate struct {
	CommodityList1      [][]uint32 `json:"commodity_list_1"`
	CommodityList2      [][]uint32 `json:"commodity_list_2"`
	CommodityList3      [][]uint32 `json:"commodity_list_3"`
	CommodityList4      [][]uint32 `json:"commodity_list_4"`
	CommodityList5      [][]uint32 `json:"commodity_list_5"`
	CommodityListCommon [][]uint32 `json:"commodity_list_common"`
	RefreshPrice        []uint32   `json:"refresh_price"`
}

type Config struct {
	Template shopTemplate
}

func LoadConfig() (*Config, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, arenaShopConfigCategory)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return &Config{}, nil
	}
	var template shopTemplate
	if err := json.Unmarshal(entries[0].Data, &template); err != nil {
		return nil, err
	}
	return &Config{Template: template}, nil
}

func EnsureState(commanderID uint32, now time.Time) (*orm.ArenaShopState, error) {
	var state orm.ArenaShopState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		state = orm.ArenaShopState{
			CommanderID:     commanderID,
			FlashCount:      0,
			LastRefreshTime: uint32(now.Unix()),
			NextFlashTime:   nextDailyReset(now),
		}
		if err := orm.GormDB.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	return &state, nil
}

func RefreshIfNeeded(commanderID uint32, now time.Time) (*orm.ArenaShopState, error) {
	state, err := EnsureState(commanderID, now)
	if err != nil {
		return nil, err
	}
	if now.Unix() >= int64(state.NextFlashTime) {
		state.FlashCount = 0
		state.LastRefreshTime = uint32(now.Unix())
		state.NextFlashTime = nextDailyReset(now)
		if err := orm.GormDB.Save(state).Error; err != nil {
			return nil, err
		}
	}
	return state, nil
}

func RefreshShop(commanderID uint32, now time.Time, config *Config) (*orm.ArenaShopState, []*protobuf.ARENASHOP, uint32, error) {
	state, err := EnsureState(commanderID, now)
	if err != nil {
		return nil, nil, 0, err
	}
	if config == nil {
		return state, nil, 0, nil
	}
	refreshCount := int(state.FlashCount + 1)
	if refreshCount > len(config.Template.RefreshPrice) {
		return state, nil, 0, nil
	}
	cost := config.Template.RefreshPrice[refreshCount-1]
	state.FlashCount++
	state.LastRefreshTime = uint32(now.Unix())
	state.NextFlashTime = nextDailyReset(now)
	if err := orm.GormDB.Save(state).Error; err != nil {
		return nil, nil, 0, err
	}
	list := BuildShopList(state.FlashCount, config)
	return state, list, cost, nil
}

func BuildShopList(flashCount uint32, config *Config) []*protobuf.ARENASHOP {
	if config == nil {
		return nil
	}
	template := config.Template
	var tier [][]uint32
	switch flashCount {
	case 0:
		tier = template.CommodityList1
	case 1:
		tier = template.CommodityList2
	case 2:
		tier = template.CommodityList3
	case 3:
		tier = template.CommodityList4
	case 4:
		tier = template.CommodityList5
	default:
		tier = nil
	}
	entries := append([][]uint32{}, tier...)
	if len(template.CommodityListCommon) > 0 {
		entries = append(entries, template.CommodityListCommon...)
	}
	return buildArenaShop(entries)
}

func buildArenaShop(entries [][]uint32) []*protobuf.ARENASHOP {
	if len(entries) == 0 {
		return nil
	}
	list := make([]*protobuf.ARENASHOP, 0, len(entries))
	for _, entry := range entries {
		if len(entry) < 2 {
			continue
		}
		list = append(list, &protobuf.ARENASHOP{
			ShopId: proto.Uint32(entry[0]),
			Count:  proto.Uint32(entry[1]),
		})
	}
	return list
}

func nextDailyReset(now time.Time) uint32 {
	utc := now.UTC()
	next := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return uint32(next.Unix())
}
