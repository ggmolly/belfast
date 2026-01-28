package answer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupShoppingStreetCommander(t *testing.T, commanderID uint32) *orm.Commander {
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   commanderID,
		Name:        fmt.Sprintf("Shop Street Commander %d", commanderID),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	resource := orm.OwnedResource{
		CommanderID: commanderID,
		ResourceID:  1,
		Amount:      1000,
	}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{resource.ResourceID: &resource}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func seedShoppingStreetOffers(t *testing.T, offers []orm.ShopOffer) {
	for _, offer := range offers {
		if err := orm.GormDB.Create(&offer).Error; err != nil {
			t.Fatalf("failed to create shop offer: %v", err)
		}
	}
}

func cleanupShoppingStreetData(t *testing.T, commanderID uint32) {
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.ShoppingStreetGood{}).Error; err != nil {
		t.Fatalf("failed to cleanup street goods: %v", err)
	}
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.ShoppingStreetState{}).Error; err != nil {
		t.Fatalf("failed to cleanup street state: %v", err)
	}
	if err := orm.GormDB.Unscoped().Delete(&orm.Commander{}, commanderID).Error; err != nil {
		t.Fatalf("failed to cleanup commander: %v", err)
	}
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.OwnedResource{}).Error; err != nil {
		t.Fatalf("failed to cleanup resources: %v", err)
	}
}

func TestGetShopStreetCreatesState(t *testing.T) {
	commanderID := uint32(5001)
	cleanupShoppingStreetData(t, commanderID)
	client := &connection.Client{Commander: setupShoppingStreetCommander(t, commanderID)}
	defer cleanupShoppingStreetData(t, commanderID)

	offers := []orm.ShopOffer{
		{ID: 901001, Type: 1, ResourceID: 1, ResourceNumber: 0, Number: 0, Effects: orm.Int64List{}, Genre: "shopping_street", Discount: 0},
		{ID: 901002, Type: 1, ResourceID: 1, ResourceNumber: 0, Number: 0, Effects: orm.Int64List{}, Genre: "shopping_street", Discount: 20},
		{ID: 901003, Type: 1, ResourceID: 1, ResourceNumber: 0, Number: 0, Effects: orm.Int64List{}, Genre: "shopping_street", Discount: 0},
	}
	orm.GormDB.Where("genre = ?", "shopping_street").Delete(&orm.ShopOffer{})
	seedShoppingStreetOffers(t, offers)

	payload := &protobuf.CS_22101{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetShopStreet(&buf, client); err != nil {
		t.Fatalf("GetShopStreet failed: %v", err)
	}
	response := &protobuf.SC_22102{}
	decodeTestPacket(t, client, 22102, response)
	street := response.GetStreet()
	if street == nil {
		t.Fatalf("expected street payload")
	}
	if len(street.GetGoodsList()) != len(offers) {
		t.Fatalf("expected %d goods, got %d", len(offers), len(street.GetGoodsList()))
	}
	offerMap := map[uint32]orm.ShopOffer{}
	for _, offer := range offers {
		offerMap[offer.ID] = offer
	}
	for _, good := range street.GetGoodsList() {
		offer, ok := offerMap[good.GetGoodsId()]
		if !ok {
			t.Fatalf("unexpected goods id %d", good.GetGoodsId())
		}
		expectedDiscount := uint32(100)
		if offer.Discount > 0 {
			expectedDiscount = uint32(100 - offer.Discount)
		}
		if good.GetDiscount() != expectedDiscount {
			t.Fatalf("expected discount %d, got %d", expectedDiscount, good.GetDiscount())
		}
		if good.GetBuyCount() != 1 {
			t.Fatalf("expected buy_count 1, got %d", good.GetBuyCount())
		}
	}

	var state orm.ShoppingStreetState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		t.Fatalf("expected state row, got error: %v", err)
	}
	if state.Level != 1 {
		t.Fatalf("expected level 1, got %d", state.Level)
	}
	if state.NextFlashTime <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next flash time in the future")
	}
}

func TestGetShopStreetRefreshUpdatesState(t *testing.T) {
	commanderID := uint32(5002)
	cleanupShoppingStreetData(t, commanderID)
	client := &connection.Client{Commander: setupShoppingStreetCommander(t, commanderID)}
	defer cleanupShoppingStreetData(t, commanderID)

	orm.GormDB.Where("genre = ?", "shopping_street").Delete(&orm.ShopOffer{})
	seedShoppingStreetOffers(t, []orm.ShopOffer{
		{ID: 902001, Type: 1, ResourceID: 1, ResourceNumber: 0, Number: 0, Effects: orm.Int64List{}, Genre: "shopping_street", Discount: 0},
		{ID: 902002, Type: 1, ResourceID: 1, ResourceNumber: 0, Number: 0, Effects: orm.Int64List{}, Genre: "shopping_street", Discount: 0},
	})
	state := orm.ShoppingStreetState{
		CommanderID:   commanderID,
		Level:         1,
		NextFlashTime: uint32(time.Now().Add(-time.Hour).Unix()),
		LevelUpTime:   0,
		FlashCount:    0,
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("failed to create state: %v", err)
	}
	if err := orm.GormDB.Create(&orm.ShoppingStreetGood{
		CommanderID: commanderID,
		GoodsID:     902001,
		Discount:    100,
		BuyCount:    0,
	}).Error; err != nil {
		t.Fatalf("failed to create goods: %v", err)
	}

	payload := &protobuf.CS_22101{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetShopStreet(&buf, client); err != nil {
		t.Fatalf("GetShopStreet failed: %v", err)
	}
	response := &protobuf.SC_22102{}
	decodeTestPacket(t, client, 22102, response)
	if len(response.GetStreet().GetGoodsList()) != 2 {
		t.Fatalf("expected refreshed goods list, got %d", len(response.GetStreet().GetGoodsList()))
	}
	var updated orm.ShoppingStreetState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&updated).Error; err != nil {
		t.Fatalf("failed to fetch state: %v", err)
	}
	if updated.NextFlashTime <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next flash time in the future")
	}
}

func TestShoppingCommandDecrementsStreetBuyCount(t *testing.T) {
	commanderID := uint32(5003)
	cleanupShoppingStreetData(t, commanderID)
	client := &connection.Client{Commander: setupShoppingStreetCommander(t, commanderID)}
	defer cleanupShoppingStreetData(t, commanderID)

	offer := orm.ShopOffer{
		ID:             903001,
		Type:           1,
		ResourceID:     1,
		ResourceNumber: 0,
		Number:         0,
		Effects:        orm.Int64List{},
		Genre:          "shopping_street",
		Discount:       0,
	}
	if err := orm.GormDB.Create(&offer).Error; err != nil {
		t.Fatalf("failed to create offer: %v", err)
	}
	if err := orm.GormDB.Create(&orm.ShoppingStreetGood{
		CommanderID: commanderID,
		GoodsID:     offer.ID,
		Discount:    100,
		BuyCount:    1,
	}).Error; err != nil {
		t.Fatalf("failed to create goods: %v", err)
	}

	payload := &protobuf.CS_16001{Id: proto.Uint32(offer.ID), Number: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ShoppingCommandAnswer(&buf, client); err != nil {
		t.Fatalf("ShoppingCommandAnswer failed: %v", err)
	}
	var updated orm.ShoppingStreetGood
	if err := orm.GormDB.Where("commander_id = ? AND goods_id = ?", commanderID, offer.ID).First(&updated).Error; err != nil {
		t.Fatalf("failed to fetch good: %v", err)
	}
	if updated.BuyCount != 0 {
		t.Fatalf("expected buy_count 0, got %d", updated.BuyCount)
	}
}
