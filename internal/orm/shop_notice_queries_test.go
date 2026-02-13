package orm

import "testing"

func TestShopNoticeQueries(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ShopOffer{})
	clearTable(t, &Notice{})
	clearTable(t, &Resource{})

	resource := Resource{ID: 1, Name: "Gold"}
	if err := CreateResourceRecord(&resource); err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	offer := ShopOffer{ID: 1, Effects: Int64List{}, Number: 1, ResourceNumber: 1, ResourceID: resource.ID, Type: 1, Genre: "daily", Discount: 0}
	if err := CreateShopOffer(&offer); err != nil {
		t.Fatalf("seed offer: %v", err)
	}
	notice := Notice{ID: 1, Version: "1", BtnTitle: "Btn", Title: "Title", TitleImage: "Img", TimeDesc: "Now", Content: "Body", TagType: 1, Icon: 1, Track: "T"}
	if err := notice.Create(); err != nil {
		t.Fatalf("seed notice: %v", err)
	}

	offers, err := ListShopOffers(ShopOfferQueryParams{Offset: 0, Limit: 10, Genre: "daily"})
	if err != nil || offers.Total != 1 {
		t.Fatalf("list offers: %v", err)
	}
	noGenre, err := ListShopOffers(ShopOfferQueryParams{Offset: 0, Limit: 10})
	if err != nil || noGenre.Total != 1 {
		t.Fatalf("list offers no genre: %v", err)
	}

	notices, err := ListNotices(NoticeQueryParams{Offset: 0, Limit: 10})
	if err != nil || notices.Total != 1 {
		t.Fatalf("list notices: %v", err)
	}
	active, err := ListActiveNotices()
	if err != nil || len(active) != 1 {
		t.Fatalf("list active notices: %v", err)
	}
}
