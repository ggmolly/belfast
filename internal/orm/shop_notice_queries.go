package orm

import "gorm.io/gorm"

func ListShopOffers(db *gorm.DB, params ShopOfferQueryParams) (ShopOfferListResult, error) {
	query := db.Model(&ShopOffer{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ShopOfferListResult{}, err
	}

	var offers []ShopOffer
	if err := query.
		Order("id asc").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&offers).Error; err != nil {
		return ShopOfferListResult{}, err
	}

	return ShopOfferListResult{Offers: offers, Total: total}, nil
}

func ListNotices(db *gorm.DB, params NoticeQueryParams) (NoticeListResult, error) {
	query := db.Model(&Notice{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return NoticeListResult{}, err
	}

	var notices []Notice
	if err := query.
		Order("id desc").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&notices).Error; err != nil {
		return NoticeListResult{}, err
	}

	return NoticeListResult{Notices: notices, Total: total}, nil
}

func ListActiveNotices(db *gorm.DB) ([]Notice, error) {
	var notices []Notice
	if err := db.Order("id desc").Limit(10).Find(&notices).Error; err != nil {
		return nil, err
	}
	return notices, nil
}
