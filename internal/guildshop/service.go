package guildshop

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	rngutil "github.com/ggmolly/belfast/internal/rng"
	"gorm.io/gorm"
)

const (
	guildStoreConfigCategory = "ShareCfg/guild_store.json"
	guildSetConfigCategory   = "ShareCfg/guildset.json"
)

type StoreEntry struct {
	ID                 uint32 `json:"id"`
	Weight             uint32 `json:"weight"`
	GoodsPurchaseLimit uint32 `json:"goods_purchase_limit"`
}

type SetEntry struct {
	Key      string   `json:"key"`
	KeyValue uint32   `json:"key_value"`
	KeyArgs  []uint32 `json:"key_args"`
}

type Config struct {
	StoreEntries []StoreEntry
	GoodsCount   uint32
	ResetCost    uint32
}

func LoadConfig() (*Config, error) {
	storeEntries, err := orm.ListConfigEntries(orm.GormDB, guildStoreConfigCategory)
	if err != nil {
		return nil, err
	}
	entries := make([]StoreEntry, 0, len(storeEntries))
	for _, entry := range storeEntries {
		var store StoreEntry
		if err := json.Unmarshal(entry.Data, &store); err != nil {
			return nil, err
		}
		if store.ID == 0 {
			continue
		}
		entries = append(entries, store)
	}
	goodsCount, err := getGuildSetValue("store_goods_quantity")
	if err != nil {
		return nil, err
	}
	resetCost, err := getGuildSetValue("store_reset_cost")
	if err != nil {
		return nil, err
	}
	if goodsCount == 0 {
		goodsCount = 10
	}
	return &Config{
		StoreEntries: entries,
		GoodsCount:   goodsCount,
		ResetCost:    resetCost,
	}, nil
}

func EnsureState(commanderID uint32, now time.Time, config *Config) (*orm.GuildShopState, []orm.GuildShopGood, error) {
	var state orm.GuildShopState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, err
		}
		state = orm.GuildShopState{
			CommanderID:     commanderID,
			RefreshCount:    0,
			NextRefreshTime: nextDailyReset(now),
		}
		if err := orm.GormDB.Create(&state).Error; err != nil {
			return nil, nil, err
		}
		goods, err := RefreshGoods(commanderID, now, config, RefreshOptions{
			RefreshCount:    0,
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

func RefreshIfNeeded(commanderID uint32, now time.Time, config *Config) (*orm.GuildShopState, []orm.GuildShopGood, error) {
	state, goods, err := EnsureState(commanderID, now, config)
	if err != nil {
		return nil, nil, err
	}
	if now.Unix() >= int64(state.NextRefreshTime) || len(goods) == 0 {
		goods, err = RefreshGoods(commanderID, now, config, RefreshOptions{
			RefreshCount:    0,
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

type RefreshOptions struct {
	RefreshCount    uint32
	NextRefreshTime uint32
}

func RefreshGoods(commanderID uint32, now time.Time, config *Config, options RefreshOptions) ([]orm.GuildShopGood, error) {
	goods := buildGoods(commanderID, config)
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("commander_id = ?", commanderID).Delete(&orm.GuildShopGood{}).Error; err != nil {
			return err
		}
		if len(goods) > 0 {
			if err := tx.Create(&goods).Error; err != nil {
				return err
			}
		}
		return tx.Model(&orm.GuildShopState{}).
			Where("commander_id = ?", commanderID).
			Updates(map[string]interface{}{
				"refresh_count":     options.RefreshCount,
				"next_refresh_time": options.NextRefreshTime,
			}).Error
	}); err != nil {
		return nil, err
	}
	return goods, nil
}

func LoadGoods(commanderID uint32) ([]orm.GuildShopGood, error) {
	var goods []orm.GuildShopGood
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Find(&goods).Error; err != nil {
		return nil, err
	}
	return goods, nil
}

func buildGoods(commanderID uint32, config *Config) []orm.GuildShopGood {
	if config == nil {
		return nil
	}
	entries := selectGoods(config.StoreEntries, int(config.GoodsCount))
	goods := make([]orm.GuildShopGood, 0, len(entries))
	for i, entry := range entries {
		count := entry.GoodsPurchaseLimit
		if count == 0 {
			count = 1
		}
		goods = append(goods, orm.GuildShopGood{
			CommanderID: commanderID,
			Index:       uint32(i + 1),
			GoodsID:     entry.ID,
			Count:       count,
		})
	}
	return goods
}

func selectGoods(entries []StoreEntry, count int) []StoreEntry {
	if count <= 0 || len(entries) == 0 {
		return nil
	}
	if len(entries) <= count {
		return entries
	}
	pool := make([]StoreEntry, len(entries))
	copy(pool, entries)
	rng := rngutil.NewLockedRand()
	selected := make([]StoreEntry, 0, count)
	for len(selected) < count && len(pool) > 0 {
		total := uint32(0)
		for _, entry := range pool {
			weight := entry.Weight
			if weight == 0 {
				weight = 1
			}
			total += weight
		}
		roll := rng.Uint32N(total)
		idx := 0
		for i, entry := range pool {
			weight := entry.Weight
			if weight == 0 {
				weight = 1
			}
			if roll < weight {
				idx = i
				break
			}
			roll -= weight
		}
		selected = append(selected, pool[idx])
		pool = append(pool[:idx], pool[idx+1:]...)
	}
	return selected
}

func getGuildSetValue(key string) (uint32, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, guildSetConfigCategory, key)
	if err != nil {
		return 0, err
	}
	var setEntry SetEntry
	if err := json.Unmarshal(entry.Data, &setEntry); err != nil {
		return 0, err
	}
	return setEntry.KeyValue, nil
}

func nextDailyReset(now time.Time) uint32 {
	utc := now.UTC()
	next := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return uint32(next.Unix())
}
