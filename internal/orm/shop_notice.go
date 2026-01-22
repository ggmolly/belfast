package orm

type ShopOfferQueryParams struct {
	Offset int
	Limit  int
	Genre  string
}

type ShopOfferListResult struct {
	Offers []ShopOffer
	Total  int64
}

type NoticeQueryParams struct {
	Offset int
	Limit  int
}

type NoticeListResult struct {
	Notices []Notice
	Total   int64
}
