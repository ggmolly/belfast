package orm

import (
	"context"
	"encoding/json"

	"github.com/ggmolly/belfast/internal/db"
)

func ListShopOffers(params ShopOfferQueryParams) (ShopOfferListResult, error) {
	ctx := context.Background()

	var total int64
	if params.Genre != "" {
		if err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM shop_offers
WHERE genre = $1
`, params.Genre).Scan(&total); err != nil {
			return ShopOfferListResult{}, err
		}
	} else {
		if err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM shop_offers
`).Scan(&total); err != nil {
			return ShopOfferListResult{}, err
		}
	}

	var (
		rows anyRows
		err  error
	)
	if params.Genre != "" {
		rows, err = db.DefaultStore.Pool.Query(ctx, `
SELECT id, effects, effect_args, number, resource_number, resource_id, type, genre, discount
FROM shop_offers
WHERE genre = $1
ORDER BY id ASC
OFFSET $2
LIMIT $3
`, params.Genre, int64(params.Offset), int64(params.Limit))
	} else {
		rows, err = db.DefaultStore.Pool.Query(ctx, `
SELECT id, effects, effect_args, number, resource_number, resource_id, type, genre, discount
FROM shop_offers
ORDER BY id ASC
OFFSET $1
LIMIT $2
`, int64(params.Offset), int64(params.Limit))
	}
	if err != nil {
		return ShopOfferListResult{}, err
	}
	defer rows.Close()

	offers := make([]ShopOffer, 0)
	for rows.Next() {
		offer, err := scanShopOffer(rows)
		if err != nil {
			return ShopOfferListResult{}, err
		}
		offers = append(offers, offer)
	}
	if err := rows.Err(); err != nil {
		return ShopOfferListResult{}, err
	}

	return ShopOfferListResult{Offers: offers, Total: total}, nil
}

func GetShopOffer(offerID uint32) (*ShopOffer, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, effects, effect_args, number, resource_number, resource_id, type, genre, discount
FROM shop_offers
WHERE id = $1
`, int64(offerID))
	offer, err := scanShopOffer(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func CreateShopOffer(offer *ShopOffer) error {
	ctx := context.Background()
	effectsPayload, err := json.Marshal(offer.Effects)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO shop_offers (id, effects, effect_args, number, resource_number, resource_id, type, genre, discount)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`,
		int64(offer.ID),
		effectsPayload,
		offer.EffectArgs,
		offer.Number,
		offer.ResourceNumber,
		int64(offer.ResourceID),
		int64(offer.Type),
		offer.Genre,
		offer.Discount,
	)
	return err
}

func UpdateShopOffer(offer *ShopOffer) error {
	ctx := context.Background()
	effectsPayload, err := json.Marshal(offer.Effects)
	if err != nil {
		return err
	}
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE shop_offers
SET effects = $2,
	effect_args = $3,
	number = $4,
	resource_number = $5,
	resource_id = $6,
	type = $7,
	genre = $8,
	discount = $9
WHERE id = $1
`,
		int64(offer.ID),
		effectsPayload,
		offer.EffectArgs,
		offer.Number,
		offer.ResourceNumber,
		int64(offer.ResourceID),
		int64(offer.Type),
		offer.Genre,
		offer.Discount,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteShopOffer(offerID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM shop_offers WHERE id = $1`, int64(offerID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListNotices(params NoticeQueryParams) (NoticeListResult, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notices`).Scan(&total); err != nil {
		return NoticeListResult{}, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, version, btn_title, title, title_image, time_desc, content, tag_type, icon, track
FROM notices
ORDER BY id DESC
OFFSET $1
LIMIT $2
`, int64(params.Offset), int64(params.Limit))
	if err != nil {
		return NoticeListResult{}, err
	}
	defer rows.Close()

	notices := make([]Notice, 0)
	for rows.Next() {
		notice, err := scanNotice(rows)
		if err != nil {
			return NoticeListResult{}, err
		}
		notices = append(notices, notice)
	}
	if err := rows.Err(); err != nil {
		return NoticeListResult{}, err
	}

	return NoticeListResult{Notices: notices, Total: total}, nil
}

func ListActiveNotices() ([]Notice, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, version, btn_title, title, title_image, time_desc, content, tag_type, icon, track
FROM notices
ORDER BY id DESC
LIMIT 10
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notices := make([]Notice, 0)
	for rows.Next() {
		notice, err := scanNotice(rows)
		if err != nil {
			return nil, err
		}
		notices = append(notices, notice)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notices, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

type anyRows interface {
	rowScanner
	Close()
	Err() error
	Next() bool
}

func scanShopOffer(scanner rowScanner) (ShopOffer, error) {
	var (
		offer          ShopOffer
		id             int64
		resourceID     int64
		typeID         int64
		effectsPayload []byte
	)
	if err := scanner.Scan(
		&id,
		&effectsPayload,
		&offer.EffectArgs,
		&offer.Number,
		&offer.ResourceNumber,
		&resourceID,
		&typeID,
		&offer.Genre,
		&offer.Discount,
	); err != nil {
		return ShopOffer{}, err
	}
	if err := json.Unmarshal(effectsPayload, &offer.Effects); err != nil {
		return ShopOffer{}, err
	}
	offer.ID = uint32(id)
	offer.ResourceID = uint32(resourceID)
	offer.Type = uint32(typeID)
	return offer, nil
}

func scanNotice(scanner rowScanner) (Notice, error) {
	var (
		notice  Notice
		id      int64
		tagType int64
		icon    int64
	)
	if err := scanner.Scan(
		&id,
		&notice.Version,
		&notice.BtnTitle,
		&notice.Title,
		&notice.TitleImage,
		&notice.TimeDesc,
		&notice.Content,
		&tagType,
		&icon,
		&notice.Track,
	); err != nil {
		return Notice{}, err
	}
	notice.ID = int(id)
	notice.TagType = int(tagType)
	notice.Icon = int(icon)
	return notice, nil
}
