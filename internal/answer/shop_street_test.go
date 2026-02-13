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
	name := fmt.Sprintf("Shop Street Commander %d", commanderID)
	if err := orm.CreateCommanderRoot(commanderID, commanderID, name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	resource := orm.OwnedResource{
		CommanderID: commanderID,
		ResourceID:  1,
		Amount:      1000,
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(resource.CommanderID), int64(resource.ResourceID), int64(resource.Amount))
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{resource.ResourceID: &resource}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func seedShoppingStreetOffers(t *testing.T, offers []orm.ShopOffer) {
	for _, offer := range offers {
		execAnswerExternalTestSQLT(t, "INSERT INTO shop_offers (id, type, resource_id, resource_number, number, effects, genre, discount) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)", int64(offer.ID), int64(offer.Type), int64(offer.ResourceID), int64(offer.ResourceNumber), int64(offer.Number), `[]`, offer.Genre, int64(offer.Discount))
	}
}

func cleanupShoppingStreetData(t *testing.T, commanderID uint32) {
	execAnswerExternalTestSQLT(t, "DELETE FROM shopping_street_goods WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM shopping_street_states WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_resources WHERE commander_id = $1", int64(commanderID))
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
	execAnswerExternalTestSQLT(t, "DELETE FROM shop_offers WHERE genre = $1", "shopping_street")
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

	state, err := orm.GetShoppingStreetState(commanderID)
	if err != nil {
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

	execAnswerExternalTestSQLT(t, "DELETE FROM shop_offers WHERE genre = $1", "shopping_street")
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
	execAnswerExternalTestSQLT(t, "INSERT INTO shopping_street_states (commander_id, level, next_flash_time, level_up_time, flash_count) VALUES ($1, $2, $3, $4, $5)", int64(state.CommanderID), int64(state.Level), int64(state.NextFlashTime), int64(state.LevelUpTime), int64(state.FlashCount))
	execAnswerExternalTestSQLT(t, "INSERT INTO shopping_street_goods (commander_id, goods_id, discount, buy_count) VALUES ($1, $2, $3, $4)", int64(commanderID), int64(902001), int64(100), int64(0))

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
	updated, err := orm.GetShoppingStreetState(commanderID)
	if err != nil {
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
	execAnswerExternalTestSQLT(t, "INSERT INTO shop_offers (id, type, resource_id, resource_number, number, effects, genre, discount) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)", int64(offer.ID), int64(offer.Type), int64(offer.ResourceID), int64(offer.ResourceNumber), int64(offer.Number), `[]`, offer.Genre, int64(offer.Discount))
	execAnswerExternalTestSQLT(t, "INSERT INTO shopping_street_goods (commander_id, goods_id, discount, buy_count) VALUES ($1, $2, $3, $4)", int64(commanderID), int64(offer.ID), int64(100), int64(1))

	payload := &protobuf.CS_16001{Id: proto.Uint32(offer.ID), Number: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ShoppingCommandAnswer(&buf, client); err != nil {
		t.Fatalf("ShoppingCommandAnswer failed: %v", err)
	}
	updated, err := orm.GetShoppingStreetGood(commanderID, offer.ID)
	if err != nil {
		t.Fatalf("failed to fetch good: %v", err)
	}
	if updated.BuyCount != 0 {
		t.Fatalf("expected buy_count 0, got %d", updated.BuyCount)
	}
}
