package answer_test

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func createDorm3dCommander(t *testing.T, commanderID uint32) *orm.Commander {
	commander := &orm.Commander{
		CommanderID: commanderID,
		AccountID:   commanderID,
		Name:        fmt.Sprintf("Dorm3d Tester %d", commanderID),
	}
	if err := orm.GormDB.Create(commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	return commander
}

func TestDorm3dApartmentDataUsesStoredData(t *testing.T) {
	commander := createDorm3dCommander(t, 9100)
	stored := orm.NewDorm3dApartment(commander.CommanderID)
	stored.DailyVigorMax = 120
	stored.Gifts = orm.Dorm3dGiftList{{GiftID: 101, Number: 5, UsedNumber: 2}}
	stored.GiftDaily = orm.Dorm3dGiftShopList{{GiftID: 202, Count: 3}}
	stored.Rooms = orm.Dorm3dRoomList{{
		ID:          7,
		Furnitures:  []orm.Dorm3dFurniture{{FurnitureID: 44, SlotID: 1}},
		Collections: []uint32{9},
		Ships:       []uint32{10},
	}}
	stored.Ships = orm.Dorm3dShipList{{
		ShipGroup:      301,
		FavorLv:        2,
		FavorExp:       123,
		RegularTrigger: []uint32{1},
		DailyFavor:     3,
		Dialogues:      []uint32{4},
		Skins:          []uint32{5},
		CurSkin:        6,
		Name:           "Dorm3d",
		NameCd:         7,
		VisitTime:      8,
		HiddenInfo:     []orm.Dorm3dSkinHiddenInfo{{SkinID: 6, HiddenParts: []uint32{1}}},
	}}
	stored.Ins = orm.Dorm3dInsList{{
		ShipGroup: 1,
		CareFlag:  1,
		CurBack:   2,
		CurCommId: 3,
		CommList: []orm.Dorm3dCommInfo{{
			ID:       11,
			Time:     100,
			ReadFlag: 1,
			ReplyList: []orm.Dorm3dKeyValue{{
				Key:   1,
				Value: 2,
			}},
		}},
		PhoneList: []orm.Dorm3dPhoneInfo{{ID: 12, Time: 200, ReadFlag: 0}},
		FriendList: []orm.Dorm3dFriendCircleInfo{{
			ID:       13,
			Time:     300,
			ReadFlag: 1,
			GoodFlag: 1,
			ExitTime: 400,
			ReplyList: []orm.Dorm3dReplyFriend{{
				Key:   3,
				Value: 4,
				Time:  300,
			}},
		}},
	}}
	if err := orm.GormDB.Create(&stored).Error; err != nil {
		t.Fatalf("failed to create dorm3d apartment: %v", err)
	}

	client := &connection.Client{Commander: commander}
	buffer := []byte{}
	if _, _, err := answer.Dorm3dApartmentData(&buffer, client); err != nil {
		t.Fatalf("Dorm3dApartmentData failed: %v", err)
	}
	response := &protobuf.SC_28000{}
	decodeTestPacket(t, client, 28000, response)
	if response.GetDailyVigorMax() != 120 {
		t.Fatalf("expected daily vigor max 120, got %d", response.GetDailyVigorMax())
	}
	if len(response.GetGifts()) != 1 {
		t.Fatalf("expected 1 gift, got %d", len(response.GetGifts()))
	}
	if response.GetGifts()[0].GetGiftId() != 101 {
		t.Fatalf("expected gift id 101, got %d", response.GetGifts()[0].GetGiftId())
	}
	if len(response.GetIns()) != 1 {
		t.Fatalf("expected 1 ins entry, got %d", len(response.GetIns()))
	}
	if len(response.GetIns()[0].GetFriendList()) != 1 {
		t.Fatalf("expected 1 friend list entry, got %d", len(response.GetIns()[0].GetFriendList()))
	}
}

func TestDorm3dInstagramOpsPersist(t *testing.T) {
	commander := createDorm3dCommander(t, 9101)
	stored := orm.NewDorm3dApartment(commander.CommanderID)
	stored.Ins = orm.Dorm3dInsList{{
		ShipGroup: 100,
		FriendList: []orm.Dorm3dFriendCircleInfo{{
			ID:        55,
			Time:      10,
			ReadFlag:  0,
			GoodFlag:  0,
			ExitTime:  0,
			ReplyList: []orm.Dorm3dReplyFriend{},
		}},
	}}
	if err := orm.GormDB.Create(&stored).Error; err != nil {
		t.Fatalf("failed to create dorm3d apartment: %v", err)
	}
	client := &connection.Client{Commander: commander}

	readPayload := &protobuf.CS_28026{
		ShipId: proto.Uint32(100),
		Type:   proto.Uint32(orm.Dorm3dInstagramOpRead),
		IdList: []uint32{55},
	}
	readBuf, err := proto.Marshal(readPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.Dorm3dInstagramOp(&readBuf, client); err != nil {
		t.Fatalf("Dorm3dInstagramOp failed: %v", err)
	}
	readResponse := &protobuf.SC_28027{}
	decodeTestPacket(t, client, 28027, readResponse)
	if readResponse.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", readResponse.GetResult())
	}

	discussPayload := &protobuf.CS_28028{
		ShipId: proto.Uint32(100),
		Type:   proto.Uint32(2),
		Id:     proto.Uint32(55),
		ChatId: proto.Uint32(9),
		Value:  proto.Uint32(3),
	}
	discussBuf, err := proto.Marshal(discussPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.Dorm3dInstagramDiscuss(&discussBuf, client); err != nil {
		t.Fatalf("Dorm3dInstagramDiscuss failed: %v", err)
	}
	discussResponse := &protobuf.SC_28029{}
	decodeTestPacket(t, client, 28029, discussResponse)
	if discussResponse.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", discussResponse.GetResult())
	}

	updated, err := orm.GetDorm3dApartment(commander.CommanderID)
	if err != nil {
		t.Fatalf("failed to load dorm3d apartment: %v", err)
	}
	if len(updated.Ins) != 1 {
		t.Fatalf("expected 1 ins entry, got %d", len(updated.Ins))
	}
	friend := updated.Ins[0].FriendList[0]
	if friend.ReadFlag != 1 {
		t.Fatalf("expected read flag 1, got %d", friend.ReadFlag)
	}
	if len(friend.ReplyList) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(friend.ReplyList))
	}
	if friend.ReplyList[0].Key != 9 || friend.ReplyList[0].Value != 3 {
		t.Fatalf("expected reply key 9 value 3, got %d/%d", friend.ReplyList[0].Key, friend.ReplyList[0].Value)
	}
	if friend.ReplyList[0].Time == 0 {
		t.Fatalf("expected reply time to be set")
	}
	if friend.Time == 0 {
		t.Fatalf("expected friend time to be set")
	}
}
