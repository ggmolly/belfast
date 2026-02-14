package orm

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

func GetArenaShopState(commanderID uint32) (*ArenaShopState, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetArenaShopState(ctx, int64(commanderID))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &ArenaShopState{
		CommanderID:     uint32(row.CommanderID),
		FlashCount:      uint32(row.FlashCount),
		LastRefreshTime: uint32(row.LastRefreshTime),
		NextFlashTime:   uint32(row.NextFlashTime),
	}, nil
}

func CreateArenaShopState(state ArenaShopState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateArenaShopState(ctx, gen.CreateArenaShopStateParams{
		CommanderID:     int64(state.CommanderID),
		FlashCount:      int64(state.FlashCount),
		LastRefreshTime: int64(state.LastRefreshTime),
		NextFlashTime:   int64(state.NextFlashTime),
	})
}

func UpdateArenaShopState(state ArenaShopState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.UpdateArenaShopState(ctx, gen.UpdateArenaShopStateParams{
		CommanderID:     int64(state.CommanderID),
		FlashCount:      int64(state.FlashCount),
		LastRefreshTime: int64(state.LastRefreshTime),
		NextFlashTime:   int64(state.NextFlashTime),
	})
}

func DeleteArenaShopState(commanderID uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM arena_shop_states WHERE commander_id = $1`, int64(commanderID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func GetShoppingStreetState(commanderID uint32) (*ShoppingStreetState, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetShoppingStreetState(ctx, int64(commanderID))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &ShoppingStreetState{
		CommanderID:   uint32(row.CommanderID),
		Level:         uint32(row.Level),
		NextFlashTime: uint32(row.NextFlashTime),
		LevelUpTime:   uint32(row.LevelUpTime),
		FlashCount:    uint32(row.FlashCount),
	}, nil
}

func CreateShoppingStreetState(state ShoppingStreetState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateShoppingStreetState(ctx, gen.CreateShoppingStreetStateParams{
		CommanderID:   int64(state.CommanderID),
		Level:         int64(state.Level),
		NextFlashTime: int64(state.NextFlashTime),
		LevelUpTime:   int64(state.LevelUpTime),
		FlashCount:    int64(state.FlashCount),
	})
}

func UpdateShoppingStreetState(state ShoppingStreetState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE shopping_street_states
SET level = $2,
	next_flash_time = $3,
	level_up_time = $4,
	flash_count = $5
WHERE commander_id = $1
`,
		int64(state.CommanderID),
		int64(state.Level),
		int64(state.NextFlashTime),
		int64(state.LevelUpTime),
		int64(state.FlashCount),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ClearShoppingStreetState(commanderID uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM shopping_street_goods WHERE commander_id = $1`, int64(commanderID)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM shopping_street_states WHERE commander_id = $1`, int64(commanderID)); err != nil {
			return err
		}
		return nil
	})
}

func LoadShoppingStreetGoods(commanderID uint32) ([]ShoppingStreetGood, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListShoppingStreetGoodsByCommander(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	goods := make([]ShoppingStreetGood, 0, len(rows))
	for _, row := range rows {
		goods = append(goods, ShoppingStreetGood{
			CommanderID: uint32(row.CommanderID),
			GoodsID:     uint32(row.GoodsID),
			Discount:    uint32(row.Discount),
			BuyCount:    uint32(row.BuyCount),
		})
	}
	return goods, nil
}

func GetShoppingStreetGood(commanderID uint32, goodsID uint32) (*ShoppingStreetGood, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	var good ShoppingStreetGood
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, goods_id, discount, buy_count
FROM shopping_street_goods
WHERE commander_id = $1 AND goods_id = $2
`, int64(commanderID), int64(goodsID)).Scan(&good.CommanderID, &good.GoodsID, &good.Discount, &good.BuyCount)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &good, nil
}

func CreateShoppingStreetGood(good ShoppingStreetGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateShoppingStreetGood(ctx, gen.CreateShoppingStreetGoodParams{
		CommanderID: int64(good.CommanderID),
		GoodsID:     int64(good.GoodsID),
		Discount:    int64(good.Discount),
		BuyCount:    int64(good.BuyCount),
	})
}

func UpdateShoppingStreetGood(good ShoppingStreetGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE shopping_street_goods
SET discount = $3,
	buy_count = $4
WHERE commander_id = $1
	AND goods_id = $2
`, int64(good.CommanderID), int64(good.GoodsID), int64(good.Discount), int64(good.BuyCount))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteShoppingStreetGood(commanderID uint32, goodsID uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM shopping_street_goods
WHERE commander_id = $1
	AND goods_id = $2
`, int64(commanderID), int64(goodsID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ReplaceShoppingStreetGoods(commanderID uint32, goods []ShoppingStreetGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.DeleteShoppingStreetGoodsByCommander(ctx, int64(commanderID)); err != nil {
			return err
		}
		for _, good := range goods {
			err := q.CreateShoppingStreetGood(ctx, gen.CreateShoppingStreetGoodParams{
				CommanderID: int64(good.CommanderID),
				GoodsID:     int64(good.GoodsID),
				Discount:    int64(good.Discount),
				BuyCount:    int64(good.BuyCount),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func RefreshShoppingStreetGoods(commanderID uint32, goods []ShoppingStreetGood, nextFlashTime uint32, flashCount uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.DeleteShoppingStreetGoodsByCommander(ctx, int64(commanderID)); err != nil {
			return err
		}
		for _, good := range goods {
			err := q.CreateShoppingStreetGood(ctx, gen.CreateShoppingStreetGoodParams{
				CommanderID: int64(good.CommanderID),
				GoodsID:     int64(good.GoodsID),
				Discount:    int64(good.Discount),
				BuyCount:    int64(good.BuyCount),
			})
			if err != nil {
				return err
			}
		}
		return q.UpdateShoppingStreetStateRefresh(ctx, gen.UpdateShoppingStreetStateRefreshParams{
			CommanderID:   int64(commanderID),
			NextFlashTime: int64(nextFlashTime),
			FlashCount:    int64(flashCount),
		})
	})
}

func ListShopOffersByGenre(genre string) ([]ShopOffer, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListShopOffersByGenre(ctx, genre)
	if err != nil {
		return nil, err
	}
	offers := make([]ShopOffer, 0, len(rows))
	for _, row := range rows {
		offer, convErr := convertShopOfferRow(row)
		if convErr != nil {
			return nil, convErr
		}
		offers = append(offers, offer)
	}
	return offers, nil
}

func ListShopOffersByIDsAndGenre(ids []uint32, genre string) ([]ShopOffer, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	if len(ids) == 0 {
		return nil, nil
	}
	lookup := make([]int64, len(ids))
	for i, id := range ids {
		lookup[i] = int64(id)
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListShopOffersByIDsAndGenre(ctx, gen.ListShopOffersByIDsAndGenreParams{
		Genre:   genre,
		Column2: lookup,
	})
	if err != nil {
		return nil, err
	}
	offers := make([]ShopOffer, 0, len(rows))
	for _, row := range rows {
		offer, convErr := convertShopOfferRow(row)
		if convErr != nil {
			return nil, convErr
		}
		offers = append(offers, offer)
	}
	return offers, nil
}

func ListShopOffersByIDs(ids []uint32) ([]ShopOffer, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	if len(ids) == 0 {
		return nil, nil
	}
	lookup := make([]int64, len(ids))
	for i, id := range ids {
		lookup[i] = int64(id)
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, effects, effect_args, number, resource_number, resource_id, type, genre, discount
FROM shop_offers
WHERE id = ANY($1::bigint[])
`, lookup)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offers := make([]ShopOffer, 0)
	for rows.Next() {
		var (
			offer     ShopOffer
			effects   []byte
			effectArg []byte
			number    int64
			resNumber int64
			resID     int64
			typ       int64
			discount  int64
		)
		if err := rows.Scan(&offer.ID, &effects, &effectArg, &number, &resNumber, &resID, &typ, &offer.Genre, &discount); err != nil {
			return nil, err
		}
		offer.Effects = nil
		if len(effects) > 0 {
			if err := json.Unmarshal(effects, &offer.Effects); err != nil {
				return nil, err
			}
		}
		offer.EffectArgs = effectArg
		offer.Number = int(number)
		offer.ResourceNumber = int(resNumber)
		offer.ResourceID = uint32(resID)
		offer.Type = uint32(typ)
		offer.Discount = int(discount)
		offers = append(offers, offer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return offers, nil
}

func convertShopOfferRow(row gen.ShopOffer) (ShopOffer, error) {
	offer := ShopOffer{
		ID:             uint32(row.ID),
		EffectArgs:     row.EffectArgs,
		Number:         int(row.Number),
		ResourceNumber: int(row.ResourceNumber),
		ResourceID:     uint32(row.ResourceID),
		Type:           uint32(row.Type),
		Genre:          row.Genre,
		Discount:       int(row.Discount),
	}
	if len(row.Effects) == 0 {
		return offer, nil
	}
	if err := json.Unmarshal(row.Effects, &offer.Effects); err != nil {
		return ShopOffer{}, err
	}
	return offer, nil
}

func GetMedalShopState(commanderID uint32) (*MedalShopState, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetMedalShopState(ctx, int64(commanderID))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &MedalShopState{CommanderID: uint32(row.CommanderID), NextRefreshTime: uint32(row.NextRefreshTime)}, nil
}

func CreateMedalShopState(state MedalShopState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateMedalShopState(ctx, gen.CreateMedalShopStateParams{
		CommanderID:     int64(state.CommanderID),
		NextRefreshTime: int64(state.NextRefreshTime),
	})
}

func UpdateMedalShopState(state MedalShopState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.UpdateMedalShopState(ctx, gen.UpdateMedalShopStateParams{
		CommanderID:     int64(state.CommanderID),
		NextRefreshTime: int64(state.NextRefreshTime),
	})
}

func LoadMedalShopGoods(commanderID uint32) ([]MedalShopGood, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListMedalShopGoodsByCommander(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	goods := make([]MedalShopGood, 0, len(rows))
	for _, row := range rows {
		goods = append(goods, MedalShopGood{
			CommanderID: uint32(row.CommanderID),
			Index:       uint32(row.Index),
			GoodsID:     uint32(row.GoodsID),
			Count:       uint32(row.Count),
		})
	}
	return goods, nil
}

func CreateMedalShopGood(good MedalShopGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateMedalShopGood(ctx, gen.CreateMedalShopGoodParams{
		CommanderID: int64(good.CommanderID),
		Index:       int64(good.Index),
		GoodsID:     int64(good.GoodsID),
		Count:       int64(good.Count),
	})
}

func UpdateMedalShopGood(good MedalShopGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE medal_shop_goods
SET goods_id = $3,
	count = $4
WHERE commander_id = $1
	AND index = $2
`, int64(good.CommanderID), int64(good.Index), int64(good.GoodsID), int64(good.Count))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteMedalShopGood(commanderID uint32, index uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM medal_shop_goods
WHERE commander_id = $1
	AND index = $2
`, int64(commanderID), int64(index))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func RefreshMedalShopGoods(commanderID uint32, goods []MedalShopGood, nextRefreshTime uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.DeleteMedalShopGoodsByCommander(ctx, int64(commanderID)); err != nil {
			return err
		}
		for _, good := range goods {
			err := q.CreateMedalShopGood(ctx, gen.CreateMedalShopGoodParams{
				CommanderID: int64(good.CommanderID),
				Index:       int64(good.Index),
				GoodsID:     int64(good.GoodsID),
				Count:       int64(good.Count),
			})
			if err != nil {
				return err
			}
		}
		return q.UpdateMedalShopState(ctx, gen.UpdateMedalShopStateParams{CommanderID: int64(commanderID), NextRefreshTime: int64(nextRefreshTime)})
	})
}

func GetMiniGameShopState(commanderID uint32) (*MiniGameShopState, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetMiniGameShopState(ctx, int64(commanderID))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &MiniGameShopState{CommanderID: uint32(row.CommanderID), NextRefreshTime: uint32(row.NextRefreshTime)}, nil
}

func CreateMiniGameShopState(state MiniGameShopState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateMiniGameShopState(ctx, gen.CreateMiniGameShopStateParams{
		CommanderID:     int64(state.CommanderID),
		NextRefreshTime: int64(state.NextRefreshTime),
	})
}

func UpdateMiniGameShopState(state MiniGameShopState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.UpdateMiniGameShopState(ctx, gen.UpdateMiniGameShopStateParams{
		CommanderID:     int64(state.CommanderID),
		NextRefreshTime: int64(state.NextRefreshTime),
	})
}

func LoadMiniGameShopGoods(commanderID uint32) ([]MiniGameShopGood, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListMiniGameShopGoodsByCommander(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	goods := make([]MiniGameShopGood, 0, len(rows))
	for _, row := range rows {
		goods = append(goods, MiniGameShopGood{
			CommanderID: uint32(row.CommanderID),
			GoodsID:     uint32(row.GoodsID),
			Count:       uint32(row.Count),
		})
	}
	return goods, nil
}

func CreateMiniGameShopGood(good MiniGameShopGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateMiniGameShopGood(ctx, gen.CreateMiniGameShopGoodParams{
		CommanderID: int64(good.CommanderID),
		GoodsID:     int64(good.GoodsID),
		Count:       int64(good.Count),
	})
}

func UpdateMiniGameShopGood(good MiniGameShopGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE mini_game_shop_goods
SET count = $3
WHERE commander_id = $1
	AND goods_id = $2
`, int64(good.CommanderID), int64(good.GoodsID), int64(good.Count))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteMiniGameShopGood(commanderID uint32, goodsID uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM mini_game_shop_goods
WHERE commander_id = $1
	AND goods_id = $2
`, int64(commanderID), int64(goodsID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func RefreshMiniGameShopGoods(commanderID uint32, goods []MiniGameShopGood, nextRefreshTime uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.DeleteMiniGameShopGoodsByCommander(ctx, int64(commanderID)); err != nil {
			return err
		}
		for _, good := range goods {
			err := q.CreateMiniGameShopGood(ctx, gen.CreateMiniGameShopGoodParams{
				CommanderID: int64(good.CommanderID),
				GoodsID:     int64(good.GoodsID),
				Count:       int64(good.Count),
			})
			if err != nil {
				return err
			}
		}
		return q.UpdateMiniGameShopState(ctx, gen.UpdateMiniGameShopStateParams{CommanderID: int64(commanderID), NextRefreshTime: int64(nextRefreshTime)})
	})
}

func GetGuildShopState(commanderID uint32) (*GuildShopState, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetGuildShopState(ctx, int64(commanderID))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &GuildShopState{CommanderID: uint32(row.CommanderID), RefreshCount: uint32(row.RefreshCount), NextRefreshTime: uint32(row.NextRefreshTime)}, nil
}

func CreateGuildShopState(state GuildShopState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateGuildShopState(ctx, gen.CreateGuildShopStateParams{
		CommanderID:     int64(state.CommanderID),
		RefreshCount:    int64(state.RefreshCount),
		NextRefreshTime: int64(state.NextRefreshTime),
	})
}

func UpdateGuildShopState(state GuildShopState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.UpdateGuildShopState(ctx, gen.UpdateGuildShopStateParams{
		CommanderID:     int64(state.CommanderID),
		RefreshCount:    int64(state.RefreshCount),
		NextRefreshTime: int64(state.NextRefreshTime),
	})
}

func LoadGuildShopGoods(commanderID uint32) ([]GuildShopGood, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListGuildShopGoodsByCommander(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	goods := make([]GuildShopGood, 0, len(rows))
	for _, row := range rows {
		goods = append(goods, GuildShopGood{
			CommanderID: uint32(row.CommanderID),
			Index:       uint32(row.Index),
			GoodsID:     uint32(row.GoodsID),
			Count:       uint32(row.Count),
		})
	}
	return goods, nil
}

func CreateGuildShopGood(good GuildShopGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateGuildShopGood(ctx, gen.CreateGuildShopGoodParams{
		CommanderID: int64(good.CommanderID),
		Index:       int64(good.Index),
		GoodsID:     int64(good.GoodsID),
		Count:       int64(good.Count),
	})
}

func UpdateGuildShopGood(good GuildShopGood) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE guild_shop_goods
SET goods_id = $3,
	count = $4
WHERE commander_id = $1
	AND index = $2
`, int64(good.CommanderID), int64(good.Index), int64(good.GoodsID), int64(good.Count))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteGuildShopGood(commanderID uint32, index uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM guild_shop_goods
WHERE commander_id = $1
	AND index = $2
`, int64(commanderID), int64(index))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func RefreshGuildShopGoods(commanderID uint32, goods []GuildShopGood, refreshCount uint32, nextRefreshTime uint32) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.DeleteGuildShopGoodsByCommander(ctx, int64(commanderID)); err != nil {
			return err
		}
		for _, good := range goods {
			err := q.CreateGuildShopGood(ctx, gen.CreateGuildShopGoodParams{
				CommanderID: int64(good.CommanderID),
				Index:       int64(good.Index),
				GoodsID:     int64(good.GoodsID),
				Count:       int64(good.Count),
			})
			if err != nil {
				return err
			}
		}
		return q.UpdateGuildShopState(ctx, gen.UpdateGuildShopStateParams{
			CommanderID:     int64(commanderID),
			RefreshCount:    int64(refreshCount),
			NextRefreshTime: int64(nextRefreshTime),
		})
	})
}

func UpdateAccountWebAuthnUserHandle(accountID string, handle []byte, updatedAt int64) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.UpdateAccountWebAuthnUserHandle(ctx, gen.UpdateAccountWebAuthnUserHandleParams{
		ID:                 accountID,
		WebAuthnUserHandle: handle,
		UpdatedAt:          pgTimestamptz(time.Unix(updatedAt, 0).UTC()),
	})
}

func InsertDebugPacket(packetSize int, packetID int, payload []byte) error {
	if db.DefaultStore == nil {
		return nil
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateDebugPacket(ctx, gen.CreateDebugPacketParams{
		PacketSize: int64(packetSize),
		PacketID:   int64(packetID),
		Data:       payload,
	})
}
