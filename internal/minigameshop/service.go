package minigameshop

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"gorm.io/gorm"
)

const gameRoomShopCategory = "ShareCfg/gameroom_shop_template.json"

type shopEntry struct {
	ID                 uint32     `json:"id"`
	GoodsPurchaseLimit uint32     `json:"goods_purchase_limit"`
	Time               [][][3]int `json:"time"`
	Order              uint32     `json:"order"`
}

type Config struct {
	Goods []shopEntry
}

type RefreshOptions struct {
	NextRefreshTime uint32
}

func LoadConfig(now time.Time) (*Config, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, gameRoomShopCategory)
	if err != nil {
		return nil, err
	}
	goods := make([]shopEntry, 0, len(entries))
	for _, entry := range entries {
		var configEntry shopEntry
		if err := json.Unmarshal(entry.Data, &configEntry); err != nil {
			return nil, err
		}
		if configEntry.ID == 0 {
			continue
		}
		if !isWithinTime(now, configEntry.Time) {
			continue
		}
		goods = append(goods, configEntry)
	}
	sort.Slice(goods, func(i, j int) bool {
		if goods[i].Order != goods[j].Order {
			return goods[i].Order < goods[j].Order
		}
		return goods[i].ID < goods[j].ID
	})
	return &Config{Goods: goods}, nil
}

func EnsureState(commanderID uint32, now time.Time, config *Config) (*orm.MiniGameShopState, []orm.MiniGameShopGood, error) {
	var state orm.MiniGameShopState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, err
		}
		state = orm.MiniGameShopState{
			CommanderID:     commanderID,
			NextRefreshTime: nextDailyReset(now),
		}
		if err := orm.GormDB.Create(&state).Error; err != nil {
			return nil, nil, err
		}
		goods, err := RefreshGoods(commanderID, config, RefreshOptions{
			NextRefreshTime: state.NextRefreshTime,
		})
		if err != nil {
			return nil, nil, err
		}
		return &state, goods, nil
	}
	goods, err := LoadGoods(commanderID)
	if err != nil {
		return nil, nil, err
	}
	return &state, goods, nil
}

func RefreshIfNeeded(commanderID uint32, now time.Time, config *Config) (*orm.MiniGameShopState, []orm.MiniGameShopGood, error) {
	state, goods, err := EnsureState(commanderID, now, config)
	if err != nil {
		return nil, nil, err
	}
	if now.Unix() >= int64(state.NextRefreshTime) || len(goods) == 0 {
		goods, err = RefreshGoods(commanderID, config, RefreshOptions{
			NextRefreshTime: nextDailyReset(now),
		})
		if err != nil {
			return nil, nil, err
		}
		if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
			return nil, nil, err
		}
	}
	return state, goods, nil
}

func RefreshGoods(commanderID uint32, config *Config, options RefreshOptions) ([]orm.MiniGameShopGood, error) {
	goods := buildGoods(commanderID, config)
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("commander_id = ?", commanderID).Delete(&orm.MiniGameShopGood{}).Error; err != nil {
			return err
		}
		if len(goods) > 0 {
			if err := tx.Create(&goods).Error; err != nil {
				return err
			}
		}
		return tx.Model(&orm.MiniGameShopState{}).
			Where("commander_id = ?", commanderID).
			Updates(map[string]interface{}{
				"next_refresh_time": options.NextRefreshTime,
			}).Error
	}); err != nil {
		return nil, err
	}
	return goods, nil
}

func LoadGoods(commanderID uint32) ([]orm.MiniGameShopGood, error) {
	var goods []orm.MiniGameShopGood
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Find(&goods).Error; err != nil {
		return nil, err
	}
	return goods, nil
}

func buildGoods(commanderID uint32, config *Config) []orm.MiniGameShopGood {
	if config == nil {
		return nil
	}
	goods := make([]orm.MiniGameShopGood, 0, len(config.Goods))
	for _, entry := range config.Goods {
		count := entry.GoodsPurchaseLimit
		if count == 0 {
			count = 1
		}
		goods = append(goods, orm.MiniGameShopGood{
			CommanderID: commanderID,
			GoodsID:     entry.ID,
			Count:       count,
		})
	}
	return goods
}

func isWithinTime(now time.Time, ranges [][][3]int) bool {
	if len(ranges) == 0 {
		return true
	}
	current := now.UTC()
	for _, window := range ranges {
		if len(window) != 2 {
			continue
		}
		start := timeFromConfig(current.Location(), window[0])
		end := timeFromConfig(current.Location(), window[1])
		if !start.IsZero() && !end.IsZero() {
			if !current.Before(start) && !current.After(end) {
				return true
			}
		}
	}
	return false
}

func timeFromConfig(loc *time.Location, parts [3]int) time.Time {
	if parts[0] == 0 && parts[1] == 0 && parts[2] == 0 {
		return time.Time{}
	}
	return time.Date(parts[0], time.Month(parts[1]), parts[2], 0, 0, 0, 0, loc)
}

func nextDailyReset(now time.Time) uint32 {
	utc := now.UTC()
	next := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return uint32(next.Unix())
}
