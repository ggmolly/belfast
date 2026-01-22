package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func Dorm3dApartmentData(buffer *[]byte, client *connection.Client) (int, int, error) {
	if client.Commander == nil {
		return 0, 28000, errors.New("missing commander")
	}
	apartment, err := orm.GetOrCreateDorm3dApartment(client.Commander.CommanderID)
	if err != nil {
		return 0, 28000, err
	}
	response := protobuf.SC_28000{
		Gifts:              buildDorm3dGifts(apartment.Gifts),
		Ships:              buildDorm3dShips(apartment.Ships),
		GiftDaily:          buildDorm3dGiftShop(apartment.GiftDaily),
		GiftPermanent:      buildDorm3dGiftShop(apartment.GiftPermanent),
		FurnitureDaily:     buildDorm3dGiftShop(apartment.FurnitureDaily),
		FurniturePermanent: buildDorm3dGiftShop(apartment.FurniturePermanent),
		DailyVigorMax:      proto.Uint32(apartment.DailyVigorMax),
		Rooms:              buildDorm3dRooms(apartment.Rooms),
		Ins:                buildDorm3dIns(apartment.Ins),
	}
	return client.SendMessage(28000, &response)
}

func buildDorm3dGifts(gifts orm.Dorm3dGiftList) []*protobuf.APARTMENT_GIFT {
	results := make([]*protobuf.APARTMENT_GIFT, 0, len(gifts))
	for _, gift := range gifts {
		results = append(results, &protobuf.APARTMENT_GIFT{
			GiftId:     proto.Uint32(gift.GiftID),
			Number:     proto.Uint32(gift.Number),
			UsedNumber: proto.Uint32(gift.UsedNumber),
		})
	}
	return results
}

func buildDorm3dGiftShop(entries orm.Dorm3dGiftShopList) []*protobuf.APARTMENT_GIFT_SHOP {
	results := make([]*protobuf.APARTMENT_GIFT_SHOP, 0, len(entries))
	for _, entry := range entries {
		results = append(results, &protobuf.APARTMENT_GIFT_SHOP{
			GiftId: proto.Uint32(entry.GiftID),
			Count:  proto.Uint32(entry.Count),
		})
	}
	return results
}

func buildDorm3dRooms(rooms orm.Dorm3dRoomList) []*protobuf.APARTMENT_ROOM {
	results := make([]*protobuf.APARTMENT_ROOM, 0, len(rooms))
	for _, room := range rooms {
		furnitures := make([]*protobuf.APARTMENT_FURNITURE, 0, len(room.Furnitures))
		for _, furniture := range room.Furnitures {
			furnitures = append(furnitures, &protobuf.APARTMENT_FURNITURE{
				FurnitureId: proto.Uint32(furniture.FurnitureID),
				SlotId:      proto.Uint32(furniture.SlotID),
			})
		}
		results = append(results, &protobuf.APARTMENT_ROOM{
			Id:          proto.Uint32(room.ID),
			Furnitures:  furnitures,
			Collections: room.Collections,
			Ships:       room.Ships,
		})
	}
	return results
}

func buildDorm3dShips(ships orm.Dorm3dShipList) []*protobuf.APARTMENT_SHIP {
	results := make([]*protobuf.APARTMENT_SHIP, 0, len(ships))
	for _, ship := range ships {
		hiddenInfo := make([]*protobuf.SKIN_HIDDEN_INFO, 0, len(ship.HiddenInfo))
		for _, hidden := range ship.HiddenInfo {
			hiddenInfo = append(hiddenInfo, &protobuf.SKIN_HIDDEN_INFO{
				SkinId:      proto.Uint32(hidden.SkinID),
				HiddenParts: hidden.HiddenParts,
			})
		}
		results = append(results, &protobuf.APARTMENT_SHIP{
			ShipGroup:      proto.Uint32(ship.ShipGroup),
			FavorLv:        proto.Uint32(ship.FavorLv),
			FavorExp:       proto.Uint32(ship.FavorExp),
			RegularTrigger: ship.RegularTrigger,
			DailyFavor:     proto.Uint32(ship.DailyFavor),
			Dialogues:      ship.Dialogues,
			Skins:          ship.Skins,
			CurSkin:        proto.Uint32(ship.CurSkin),
			Name:           proto.String(ship.Name),
			NameCd:         proto.Uint32(ship.NameCd),
			VisitTime:      proto.Uint32(ship.VisitTime),
			HiddenInfo:     hiddenInfo,
		})
	}
	return results
}

func buildDorm3dIns(entries orm.Dorm3dInsList) []*protobuf.APARTMENT_INS {
	results := make([]*protobuf.APARTMENT_INS, 0, len(entries))
	for _, entry := range entries {
		commList := make([]*protobuf.COMM_INFO, 0, len(entry.CommList))
		for _, comm := range entry.CommList {
			replies := make([]*protobuf.KEYVALUE_P28, 0, len(comm.ReplyList))
			for _, reply := range comm.ReplyList {
				replies = append(replies, &protobuf.KEYVALUE_P28{
					Key:   proto.Uint32(reply.Key),
					Value: proto.Uint32(reply.Value),
				})
			}
			commList = append(commList, &protobuf.COMM_INFO{
				Id:        proto.Uint32(comm.ID),
				Time:      proto.Uint32(comm.Time),
				ReadFlag:  proto.Uint32(comm.ReadFlag),
				ReplyList: replies,
			})
		}
		phoneList := make([]*protobuf.PHONE_INFO, 0, len(entry.PhoneList))
		for _, phone := range entry.PhoneList {
			phoneList = append(phoneList, &protobuf.PHONE_INFO{
				Id:       proto.Uint32(phone.ID),
				Time:     proto.Uint32(phone.Time),
				ReadFlag: proto.Uint32(phone.ReadFlag),
			})
		}
		friendList := make([]*protobuf.FRIEND_CIRCLE_INFO, 0, len(entry.FriendList))
		for _, friend := range entry.FriendList {
			replies := make([]*protobuf.REPLY_FRIEND, 0, len(friend.ReplyList))
			for _, reply := range friend.ReplyList {
				replies = append(replies, &protobuf.REPLY_FRIEND{
					Key:   proto.Uint32(reply.Key),
					Value: proto.Uint32(reply.Value),
					Time:  proto.Uint32(reply.Time),
				})
			}
			friendList = append(friendList, &protobuf.FRIEND_CIRCLE_INFO{
				Id:        proto.Uint32(friend.ID),
				Time:      proto.Uint32(friend.Time),
				ReadFlag:  proto.Uint32(friend.ReadFlag),
				GoodFlag:  proto.Uint32(friend.GoodFlag),
				ReplyList: replies,
				ExitTime:  proto.Uint32(friend.ExitTime),
			})
		}
		results = append(results, &protobuf.APARTMENT_INS{
			ShipGroup:  proto.Uint32(entry.ShipGroup),
			CareFlag:   proto.Uint32(entry.CareFlag),
			CurBack:    proto.Uint32(entry.CurBack),
			CurCommId:  proto.Uint32(entry.CurCommId),
			CommList:   commList,
			PhoneList:  phoneList,
			FriendList: friendList,
		})
	}
	return results
}
