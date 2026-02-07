package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MonthShopPurchase struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	GoodsID     uint32    `gorm:"primaryKey;autoIncrement:false"`
	Month       uint32    `gorm:"primaryKey;autoIncrement:false"`
	BuyCount    uint32    `gorm:"not_null"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func ListMonthShopPurchaseCounts(commanderID uint32, month uint32) (map[uint32]uint32, error) {
	var entries []MonthShopPurchase
	if err := GormDB.Where("commander_id = ? AND month = ?", commanderID, month).Find(&entries).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint32]uint32, len(entries))
	for _, entry := range entries {
		counts[entry.GoodsID] = entry.BuyCount
	}
	return counts, nil
}

func GetMonthShopPurchaseCountTx(tx *gorm.DB, commanderID uint32, goodsID uint32, month uint32) (uint32, error) {
	var entry MonthShopPurchase
	result := tx.Where("commander_id = ? AND goods_id = ? AND month = ?", commanderID, goodsID, month).Limit(1).Find(&entry)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, nil
	}
	return entry.BuyCount, nil
}

func IncrementMonthShopPurchaseTx(tx *gorm.DB, commanderID uint32, goodsID uint32, month uint32, delta uint32) error {
	entry := MonthShopPurchase{CommanderID: commanderID, GoodsID: goodsID, Month: month, BuyCount: delta}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "goods_id"}, {Name: "month"}},
		DoUpdates: clause.Assignments(map[string]any{
			"buy_count":  gorm.Expr("buy_count + ?", delta),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&entry).Error
}
