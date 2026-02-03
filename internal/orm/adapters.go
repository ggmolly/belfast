package orm

import (
	"time"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type BuildInfoPayload struct {
	Build      *Build
	PoolID     uint32
	BuildTime  uint32
	FinishTime time.Time
}

func ToProtoBuildInfo(payload BuildInfoPayload) *protobuf.BUILDINFO {
	finishTime := payload.FinishTime
	poolID := payload.PoolID
	if payload.Build != nil {
		finishTime = payload.Build.FinishesAt
		if payload.Build.PoolID != 0 {
			poolID = payload.Build.PoolID
		}
	}
	return &protobuf.BUILDINFO{
		Time:       proto.Uint32(payload.BuildTime),
		FinishTime: proto.Uint32(uint32(finishTime.Unix())),
		BuildId:    proto.Uint32(poolID),
	}
}

func ToProtoOwnedShip(ship OwnedShip, randomFlags []uint32) *protobuf.SHIPINFO {
	equipInfo := buildEquipInfoList(ship)
	return &protobuf.SHIPINFO{
		Id:                  proto.Uint32(ship.ID),
		TemplateId:          proto.Uint32(ship.ShipID),
		Level:               proto.Uint32(ship.Level),
		Exp:                 proto.Uint32(ship.Exp),
		EquipInfoList:       equipInfo,
		Energy:              proto.Uint32(ship.Energy),
		State:               &protobuf.SHIPSTATE{State: proto.Uint32(0)},
		IsLocked:            proto.Uint32(boolToUint32(ship.IsLocked)),
		Intimacy:            proto.Uint32(ship.Intimacy),
		Proficiency:         proto.Uint32(boolToUint32(ship.Proficiency)),
		CreateTime:          proto.Uint32(uint32(ship.CreateTime.Unix())),
		SkinId:              proto.Uint32(ship.SkinID),
		Propose:             proto.Uint32(boolToUint32(ship.Propose)),
		Name:                proto.String(ship.CustomName),
		ChangeNameTimestamp: proto.Uint32(uint32(ship.ChangeNameTimestamp.Unix())),
		MaxLevel:            proto.Uint32(ship.MaxLevel),
		CommonFlag:          proto.Uint32(boolToUint32(ship.CommonFlag)),
		ActivityNpc:         proto.Uint32(ship.ActivityNPC),
		MetaRepairList:      nil,
		Spweapon:            nil,
		CharRandomFlag:      randomFlags,
	}
}

func buildEquipInfoList(ship OwnedShip) []*protobuf.EQUIPSKIN_INFO {
	slotCount := uint32(3)
	if config, err := GetShipEquipConfig(ship.ShipID); err == nil {
		slotCount = config.SlotCount()
	}
	result := make([]*protobuf.EQUIPSKIN_INFO, slotCount)
	index := make(map[uint32]OwnedShipEquipment, len(ship.Equipments))
	for _, entry := range ship.Equipments {
		index[entry.Pos] = entry
	}
	for pos := uint32(1); pos <= slotCount; pos++ {
		entry, ok := index[pos]
		equipID := uint32(0)
		skinID := uint32(0)
		if ok {
			equipID = entry.EquipID
			skinID = entry.SkinID
		}
		result[pos-1] = &protobuf.EQUIPSKIN_INFO{
			Id:     proto.Uint32(equipID),
			SkinId: proto.Uint32(skinID),
		}
	}
	return result
}

func boolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func ToProtoOwnedShipList(ships []OwnedShip, randomFlags map[uint32][]uint32) []*protobuf.SHIPINFO {
	result := make([]*protobuf.SHIPINFO, len(ships))
	for i, ship := range ships {
		result[i] = ToProtoOwnedShip(ship, randomFlags[ship.ID])
	}
	return result
}

func ToProtoDropInfoList(attachments []MailAttachment) []*protobuf.DROPINFO {
	result := make([]*protobuf.DROPINFO, len(attachments))
	for i, attachment := range attachments {
		a := attachment
		result[i] = &protobuf.DROPINFO{
			Type:   &a.Type,
			Id:     &a.ItemID,
			Number: &a.Quantity,
		}
	}
	return result
}

func ToProtoCompensationDropInfoList(attachments []CompensationAttachment) []*protobuf.DROPINFO {
	result := make([]*protobuf.DROPINFO, len(attachments))
	for i, attachment := range attachments {
		a := attachment
		result[i] = &protobuf.DROPINFO{
			Type:   &a.Type,
			Id:     &a.ItemID,
			Number: &a.Quantity,
		}
	}
	return result
}

func ToProtoProposeResponse(success bool) *protobuf.SC_12033 {
	result := proto.Uint32(0)
	if !success {
		result = proto.Uint32(1)
	}
	return &protobuf.SC_12033{
		Result: result,
		Time:   proto.Uint32(uint32(time.Now().Unix())),
	}
}
