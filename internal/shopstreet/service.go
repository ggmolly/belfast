package shopstreet

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	rngutil "github.com/ggmolly/belfast/internal/rng"
	"gorm.io/gorm"
)

const (
	DefaultGoodsCount     = 10
	DefaultRefreshSeconds = 24 * 60 * 60
)

type RefreshOptions struct {
	GoodsCount         *int
	NextFlashInSeconds *uint32
	SetFlashCount      *uint32
	Seed               *int64
	GoodsIDs           []uint32
	DiscountOverride   *uint32
	BuyCount           *uint32
}

func RefreshIfNeeded(commanderID uint32, now time.Time) (*orm.ShoppingStreetState, []orm.ShoppingStreetGood, error) {
	state, goods, err := EnsureState(commanderID, now)
	if err != nil {
		return nil, nil, err
	}
	if now.Unix() >= int64(state.NextFlashTime) || len(goods) == 0 {
		options := RefreshOptions{
			GoodsCount:         intPtr(DefaultGoodsCount),
			NextFlashInSeconds: uint32Ptr(DefaultRefreshSeconds),
			SetFlashCount:      uint32Ptr(0),
			BuyCount:           uint32Ptr(1),
		}
		return RefreshGoods(commanderID, now, options)
	}
	return state, goods, nil
}

func EnsureState(commanderID uint32, now time.Time) (*orm.ShoppingStreetState, []orm.ShoppingStreetGood, error) {
	var state orm.ShoppingStreetState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, err
		}
		state = orm.ShoppingStreetState{
			CommanderID:   commanderID,
			Level:         1,
			NextFlashTime: uint32(now.Unix()) + DefaultRefreshSeconds,
			LevelUpTime:   0,
			FlashCount:    0,
		}
		if err := orm.GormDB.Create(&state).Error; err != nil {
			return nil, nil, err
		}
		goods, err := refreshGoods(commanderID, now, RefreshOptions{
			GoodsCount:         intPtr(DefaultGoodsCount),
			NextFlashInSeconds: uint32Ptr(DefaultRefreshSeconds),
			SetFlashCount:      uint32Ptr(0),
			BuyCount:           uint32Ptr(1),
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

func RefreshGoods(commanderID uint32, now time.Time, options RefreshOptions) (*orm.ShoppingStreetState, []orm.ShoppingStreetGood, error) {
	state, err := loadOrCreateState(commanderID, now)
	if err != nil {
		return nil, nil, err
	}
	goods, err := refreshGoods(commanderID, now, options)
	if err != nil {
		return nil, nil, err
	}
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		return nil, nil, err
	}
	return state, goods, nil
}

func ReplaceGoods(commanderID uint32, goods []orm.ShoppingStreetGood) ([]orm.ShoppingStreetGood, error) {
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("commander_id = ?", commanderID).Delete(&orm.ShoppingStreetGood{}).Error; err != nil {
			return err
		}
		if len(goods) == 0 {
			return nil
		}
		return tx.Create(&goods).Error
	}); err != nil {
		return nil, err
	}
	return LoadGoods(commanderID)
}

func LoadGoods(commanderID uint32) ([]orm.ShoppingStreetGood, error) {
	var goods []orm.ShoppingStreetGood
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Find(&goods).Error; err != nil {
		return nil, err
	}
	return goods, nil
}

func ResolveOffers(ids []uint32) ([]orm.ShopOffer, []uint32, error) {
	if len(ids) == 0 {
		return nil, nil, nil
	}
	var offers []orm.ShopOffer
	if err := orm.GormDB.Where("id IN ? AND genre = ?", ids, "shopping_street").Find(&offers).Error; err != nil {
		return nil, nil, err
	}
	lookup := make(map[uint32]orm.ShopOffer, len(offers))
	for _, offer := range offers {
		lookup[offer.ID] = offer
	}
	invalid := make([]uint32, 0)
	ordered := make([]orm.ShopOffer, 0, len(ids))
	for _, id := range ids {
		offer, ok := lookup[id]
		if !ok {
			invalid = append(invalid, id)
			continue
		}
		ordered = append(ordered, offer)
	}
	return ordered, invalid, nil
}

func loadOrCreateState(commanderID uint32, now time.Time) (*orm.ShoppingStreetState, error) {
	var state orm.ShoppingStreetState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		state = orm.ShoppingStreetState{
			CommanderID:   commanderID,
			Level:         1,
			NextFlashTime: uint32(now.Unix()) + DefaultRefreshSeconds,
			LevelUpTime:   0,
			FlashCount:    0,
		}
		if err := orm.GormDB.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	return &state, nil
}

func refreshGoods(commanderID uint32, now time.Time, options RefreshOptions) ([]orm.ShoppingStreetGood, error) {
	goodsCount := DefaultGoodsCount
	if options.GoodsCount != nil {
		goodsCount = *options.GoodsCount
	}
	nextFlash := uint32(DefaultRefreshSeconds)
	if options.NextFlashInSeconds != nil {
		nextFlash = *options.NextFlashInSeconds
	}
	flashCount := uint32(0)
	if options.SetFlashCount != nil {
		flashCount = *options.SetFlashCount
	}
	buyCount := uint32(1)
	if options.BuyCount != nil {
		buyCount = *options.BuyCount
	}
	var offers []orm.ShopOffer
	if len(options.GoodsIDs) > 0 {
		resolved, _, err := ResolveOffers(options.GoodsIDs)
		if err != nil {
			return nil, err
		}
		offers = resolved
	} else {
		var err error
		offers, err = getShoppingStreetOffers()
		if err != nil {
			return nil, err
		}
		offers = selectOffers(offers, goodsCount, options.Seed)
	}
	goods := buildGoods(commanderID, offers, buyCount, options.DiscountOverride)
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("commander_id = ?", commanderID).Delete(&orm.ShoppingStreetGood{}).Error; err != nil {
			return err
		}
		if len(goods) > 0 {
			if err := tx.Create(&goods).Error; err != nil {
				return err
			}
		}
		return tx.Model(&orm.ShoppingStreetState{}).
			Where("commander_id = ?", commanderID).
			Updates(map[string]interface{}{
				"next_flash_time": uint32(now.Unix()) + nextFlash,
				"flash_count":     flashCount,
			}).Error
	}); err != nil {
		return nil, err
	}
	return goods, nil
}

func getShoppingStreetOffers() ([]orm.ShopOffer, error) {
	var offers []orm.ShopOffer
	if err := orm.GormDB.Where("genre = ?", "shopping_street").Find(&offers).Error; err != nil {
		return nil, err
	}
	return offers, nil
}

func selectOffers(offers []orm.ShopOffer, count int, seed *int64) []orm.ShopOffer {
	if len(offers) <= count {
		return offers
	}
	shuffled := make([]orm.ShopOffer, len(offers))
	copy(shuffled, offers)
	rng := rngutil.NewLockedRand()
	if seed != nil {
		rng = rngutil.NewLockedRandFromSeed(uint64(*seed))
	}
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:count]
}

func buildGoods(commanderID uint32, offers []orm.ShopOffer, buyCount uint32, discountOverride *uint32) []orm.ShoppingStreetGood {
	goods := make([]orm.ShoppingStreetGood, 0, len(offers))
	for _, offer := range offers {
		discount := uint32(100)
		if discountOverride != nil {
			discount = *discountOverride
		} else if offer.Discount > 0 {
			discount = uint32(100 - offer.Discount)
		}
		goods = append(goods, orm.ShoppingStreetGood{
			CommanderID: commanderID,
			GoodsID:     offer.ID,
			Discount:    discount,
			BuyCount:    buyCount,
		})
	}
	return goods
}

func intPtr(value int) *int {
	return &value
}

func uint32Ptr(value uint32) *uint32 {
	return &value
}
