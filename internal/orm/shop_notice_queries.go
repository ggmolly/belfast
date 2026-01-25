package orm

import "gorm.io/gorm"

func ListShopOffers(db *gorm.DB, params ShopOfferQueryParams) (ShopOfferListResult, error) {
	query := db.Model(&ShopOffer{})
	if params.Genre != "" {
		query = query.Where("genre = ?", params.Genre)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ShopOfferListResult{}, err
	}

	var offers []ShopOffer
	query = query.Order("id asc")
	query = ApplyPagination(query, params.Offset, params.Limit)
	if err := query.Find(&offers).Error; err != nil {
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
	query = query.Order("id desc")
	query = ApplyPagination(query, params.Offset, params.Limit)
	if err := query.Find(&notices).Error; err != nil {
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
