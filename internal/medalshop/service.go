package medalshop

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"gorm.io/gorm"
)

const (
	monthShopConfigCategory = "ShareCfg/month_shop_template.json"
	shopTemplateCategory    = "ShareCfg/shop_template.json"
)

type monthShopTemplate struct {
	HonorMedalShopGoods []uint32 `json:"honormedal_shop_goods"`
}

type shopTemplateEntry struct {
	ID                 uint32 `json:"id"`
	GoodsPurchaseLimit uint32 `json:"goods_purchase_limit"`
}

type Config struct {
	GoodsIDs      []uint32
	PurchaseLimit map[uint32]uint32
}

type RefreshOptions struct {
	NextRefreshTime uint32
}

func LoadConfig() (*Config, error) {
	monthEntries, err := orm.ListConfigEntries(orm.GormDB, monthShopConfigCategory)
	if err != nil {
		return nil, err
	}
	var template monthShopTemplate
	if len(monthEntries) > 0 {
		if err := json.Unmarshal(monthEntries[0].Data, &template); err != nil {
			return nil, err
		}
	}
	purchaseLimit := map[uint32]uint32{}
	shopEntries, err := orm.ListConfigEntries(orm.GormDB, shopTemplateCategory)
	if err != nil {
		return nil, err
	}
	for _, entry := range shopEntries {
		var shopItem shopTemplateEntry
		if err := json.Unmarshal(entry.Data, &shopItem); err != nil {
			return nil, err
		}
		if shopItem.ID == 0 {
			continue
		}
		purchaseLimit[shopItem.ID] = shopItem.GoodsPurchaseLimit
	}
	return &Config{
		GoodsIDs:      template.HonorMedalShopGoods,
		PurchaseLimit: purchaseLimit,
	}, nil
}

func EnsureState(commanderID uint32, now time.Time, config *Config) (*orm.MedalShopState, []orm.MedalShopGood, error) {
	var state orm.MedalShopState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, err
		}
		state = orm.MedalShopState{
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

func RefreshIfNeeded(commanderID uint32, now time.Time, config *Config) (*orm.MedalShopState, []orm.MedalShopGood, error) {
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

func RefreshGoods(commanderID uint32, config *Config, options RefreshOptions) ([]orm.MedalShopGood, error) {
	goods := buildGoods(commanderID, config)
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("commander_id = ?", commanderID).Delete(&orm.MedalShopGood{}).Error; err != nil {
			return err
		}
		if len(goods) > 0 {
			if err := tx.Create(&goods).Error; err != nil {
				return err
			}
		}
		return tx.Model(&orm.MedalShopState{}).
			Where("commander_id = ?", commanderID).
			Updates(map[string]interface{}{
				"next_refresh_time": options.NextRefreshTime,
			}).Error
	}); err != nil {
		return nil, err
	}
	return goods, nil
}

func LoadGoods(commanderID uint32) ([]orm.MedalShopGood, error) {
	var goods []orm.MedalShopGood
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Find(&goods).Error; err != nil {
		return nil, err
	}
	return goods, nil
}

func buildGoods(commanderID uint32, config *Config) []orm.MedalShopGood {
	if config == nil || len(config.GoodsIDs) == 0 {
		return nil
	}
	goods := make([]orm.MedalShopGood, 0, len(config.GoodsIDs))
	for i, id := range config.GoodsIDs {
		count := config.PurchaseLimit[id]
		if count == 0 {
			count = 1
		}
		goods = append(goods, orm.MedalShopGood{
			CommanderID: commanderID,
			Index:       uint32(i + 1),
			GoodsID:     id,
			Count:       count,
		})
	}
	return goods
}

func nextDailyReset(now time.Time) uint32 {
	utc := now.UTC()
	next := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return uint32(next.Unix())
}

func NextDailyReset(now time.Time) uint32 {
	return nextDailyReset(now)
}
