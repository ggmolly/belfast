package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
	"github.com/jackc/pgx/v5"
)

type MonthShopPurchase struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	GoodsID     uint32    `gorm:"primaryKey;autoIncrement:false"`
	Month       uint32    `gorm:"primaryKey;autoIncrement:false"`
	BuyCount    uint32    `gorm:"not_null"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func ListMonthShopPurchaseCounts(commanderID uint32, month uint32) (map[uint32]uint32, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListMonthShopPurchasesByCommanderAndMonth(ctx, gen.ListMonthShopPurchasesByCommanderAndMonthParams{CommanderID: int64(commanderID), Month: int64(month)})
	if err != nil {
		return nil, err
	}
	counts := make(map[uint32]uint32, len(rows))
	for _, r := range rows {
		counts[uint32(r.GoodsID)] = uint32(r.BuyCount)
	}
	return counts, nil
}

func GetMonthShopPurchaseCountTx(ctx context.Context, tx pgx.Tx, commanderID uint32, goodsID uint32, month uint32) (uint32, error) {
	qtx := db.DefaultStore.Queries.WithTx(tx)
	row, err := qtx.GetMonthShopPurchase(ctx, gen.GetMonthShopPurchaseParams{CommanderID: int64(commanderID), GoodsID: int64(goodsID), Month: int64(month)})
	err = db.MapNotFound(err)
	if db.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return uint32(row.BuyCount), nil
}

func IncrementMonthShopPurchaseTx(ctx context.Context, tx pgx.Tx, commanderID uint32, goodsID uint32, month uint32, delta uint32) error {
	qtx := db.DefaultStore.Queries.WithTx(tx)
	return qtx.IncrementMonthShopPurchase(ctx, gen.IncrementMonthShopPurchaseParams{CommanderID: int64(commanderID), GoodsID: int64(goodsID), Month: int64(month), BuyCount: int64(delta)})
}
