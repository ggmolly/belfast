package orm

import (
	"testing"
)

func TestDorm3dApartmentLifecycle(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Dorm3dApartment{})

	if _, err := GetDorm3dApartment(1); err == nil {
		t.Fatalf("expected error for missing apartment")
	}
	apartment, err := GetOrCreateDorm3dApartment(1)
	if err != nil {
		t.Fatalf("get or create apartment: %v", err)
	}
	if apartment.CommanderID != 1 {
		t.Fatalf("unexpected commander id")
	}
	if apartment.Gifts == nil || apartment.Ships == nil || apartment.Ins == nil {
		t.Fatalf("expected defaults initialized")
	}
}

func TestDorm3dInstagramUpdatesAndReplies(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Dorm3dApartment{})

	if err := UpdateDorm3dInstagramFlags(2, 10, []uint32{55}, Dorm3dInstagramOpRead, 100); err != nil {
		t.Fatalf("update instagram flags: %v", err)
	}
	if err := UpdateDorm3dInstagramFlags(2, 10, []uint32{55}, Dorm3dInstagramOpLike, 100); err != nil {
		t.Fatalf("update instagram like: %v", err)
	}
	if err := AddDorm3dInstagramReply(2, 10, 55, 7, 9, 100); err != nil {
		t.Fatalf("add instagram reply: %v", err)
	}
	apartment, err := GetDorm3dApartment(2)
	if err != nil {
		t.Fatalf("get apartment: %v", err)
	}
	if len(apartment.Ins) != 1 || len(apartment.Ins[0].FriendList) != 1 {
		t.Fatalf("expected ins entries")
	}
	entry := apartment.Ins[0].FriendList[0]
	if entry.ReadFlag != 1 || entry.GoodFlag != 1 {
		t.Fatalf("expected read and like flags set")
	}
	if len(entry.ReplyList) != 1 {
		t.Fatalf("expected reply list")
	}
}

func TestDorm3dEnsureDefaults(t *testing.T) {
	apartment := Dorm3dApartment{}
	apartment.EnsureDefaults()
	if apartment.Gifts == nil || apartment.Ships == nil || apartment.Ins == nil {
		t.Fatalf("expected defaults set")
	}
}

func TestDorm3dJSONScan(t *testing.T) {
	list := Dorm3dGiftList{{GiftID: 1}}
	value, err := list.Value()
	if err != nil {
		t.Fatalf("value: %v", err)
	}
	var decoded Dorm3dGiftList
	if err := decoded.Scan(value); err != nil {
		t.Fatalf("scan string: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("expected decoded list")
	}
	var decodedBytes Dorm3dGiftList
	if err := decodedBytes.Scan([]byte("[]")); err != nil {
		t.Fatalf("scan bytes: %v", err)
	}
	if err := decodedBytes.Scan(nil); err != nil {
		t.Fatalf("scan nil: %v", err)
	}
	if err := decodedBytes.Scan(123); err == nil {
		t.Fatalf("expected scan error for unsupported type")
	}

	giftShop := Dorm3dGiftShopList{{GiftID: 1, Count: 2}}
	value, err = giftShop.Value()
	if err != nil {
		t.Fatalf("gift shop value: %v", err)
	}
	var decodedGiftShop Dorm3dGiftShopList
	if err := decodedGiftShop.Scan(value); err != nil {
		t.Fatalf("gift shop scan: %v", err)
	}

	rooms := Dorm3dRoomList{{ID: 1}}
	value, err = rooms.Value()
	if err != nil {
		t.Fatalf("room value: %v", err)
	}
	var decodedRooms Dorm3dRoomList
	if err := decodedRooms.Scan(value); err != nil {
		t.Fatalf("room scan: %v", err)
	}

	ships := Dorm3dShipList{{ShipGroup: 1, Name: "X"}}
	value, err = ships.Value()
	if err != nil {
		t.Fatalf("ship value: %v", err)
	}
	var decodedShips Dorm3dShipList
	if err := decodedShips.Scan(value); err != nil {
		t.Fatalf("ship scan: %v", err)
	}

	ins := Dorm3dInsList{{ShipGroup: 1}}
	value, err = ins.Value()
	if err != nil {
		t.Fatalf("ins value: %v", err)
	}
	var decodedIns Dorm3dInsList
	if err := decodedIns.Scan(value); err != nil {
		t.Fatalf("ins scan: %v", err)
	}
}
