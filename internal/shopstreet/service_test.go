package shopstreet

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
)

func TestSelectOffers(t *testing.T) {
	offers := []orm.ShopOffer{{ID: 1}, {ID: 2}, {ID: 3}}
	seed := int64(1)
	selected := selectOffers(offers, 2, &seed)
	if len(selected) != 2 {
		t.Fatalf("expected 2 offers, got %d", len(selected))
	}
}

func TestSelectOffersCountLarger(t *testing.T) {
	offers := []orm.ShopOffer{{ID: 1}, {ID: 2}}
	selected := selectOffers(offers, 5, nil)
	if len(selected) != 2 {
		t.Fatalf("expected all offers, got %d", len(selected))
	}
}

func TestBuildGoods(t *testing.T) {
	offers := []orm.ShopOffer{{ID: 1, Discount: 10}, {ID: 2, Discount: 0}}
	goods := buildGoods(10, offers, 2, nil)
	if len(goods) != 2 {
		t.Fatalf("expected 2 goods")
	}
	if goods[0].Discount != 90 {
		t.Fatalf("expected discount 90, got %d", goods[0].Discount)
	}
	if goods[1].Discount != 100 {
		t.Fatalf("expected discount 100, got %d", goods[1].Discount)
	}
}

func TestBuildGoodsOverride(t *testing.T) {
	offers := []orm.ShopOffer{{ID: 1, Discount: 10}}
	override := uint32(42)
	goods := buildGoods(10, offers, 2, &override)
	if goods[0].Discount != 42 {
		t.Fatalf("expected discount override 42, got %d", goods[0].Discount)
	}
}
