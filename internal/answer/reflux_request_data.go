package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func buildRefluxInactiveResponse() protobuf.SC_11752 {
	return protobuf.SC_11752{
		Active:          proto.Uint32(0),
		ReturnLv:        proto.Uint32(0),
		ReturnTime:      proto.Uint32(0),
		ShipNumber:      proto.Uint32(0),
		LastOfflineTime: proto.Uint32(0),
		Pt:              proto.Uint32(0),
		SignCnt:         proto.Uint32(0),
		SignLastTime:    proto.Uint32(0),
		PtStage:         proto.Uint32(0),
	}
}

func RefluxRequestData(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11751
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11752, err
	}
	nowTime := time.Now().UTC()
	now := uint32(nowTime.Unix())
	_, signIDs, err := loadReturnSignTemplates()
	if err != nil {
		if errors.Is(err, errReturnSignTemplateMissing) {
			response := buildRefluxInactiveResponse()
			return client.SendMessage(11752, &response)
		}
		return 0, 11752, err
	}
	_, _, ptItemID, err := loadReturnPtTemplates()
	if err != nil {
		if errors.Is(err, errReturnPtTemplateMissing) {
			response := buildRefluxInactiveResponse()
			return client.SendMessage(11752, &response)
		}
		return 0, 11752, err
	}
	state, err := orm.GetOrCreateRefluxState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11752, err
	}
	if state.Active == 1 && isRefluxExpired(state.ReturnTime, uint32(len(signIDs)), now) {
		state.Active = 0
		if err := orm.SaveRefluxState(orm.GormDB, state); err != nil {
			return 0, 11752, err
		}
	}
	if state.Active == 0 {
		cfg, ok, err := loadRefluxEligibilityConfig()
		if err != nil {
			return 0, 11752, err
		}
		if ok && isRefluxEligible(client, cfg, nowTime) {
			if err := ensureCommanderLoaded(client, "Reflux"); err != nil {
				return 0, 11752, err
			}
			state.Active = 1
			state.ReturnLv = uint32(client.Commander.Level)
			state.ReturnTime = now
			state.ShipNumber = uint32(len(client.Commander.OwnedShipsMap))
			if !client.PreviousLoginAt.IsZero() {
				state.LastOfflineTime = uint32(client.PreviousLoginAt.Unix())
			} else {
				state.LastOfflineTime = 0
			}
			state.Pt = 0
			state.SignCnt = 0
			state.SignLastTime = 0
			state.PtStage = 0
		}
	}
	if state.Active == 1 && ptItemID != 0 {
		ptCount, _, err := getCommanderItemCountFromDB(client, ptItemID)
		if err != nil {
			return 0, 11752, err
		}
		state.Pt = ptCount
	}
	if err := orm.SaveRefluxState(orm.GormDB, state); err != nil {
		return 0, 11752, err
	}
	response := protobuf.SC_11752{}
	if state.Active == 1 {
		response.Active = proto.Uint32(1)
		response.ReturnLv = proto.Uint32(state.ReturnLv)
		response.ReturnTime = proto.Uint32(state.ReturnTime)
		response.ShipNumber = proto.Uint32(state.ShipNumber)
		response.LastOfflineTime = proto.Uint32(state.LastOfflineTime)
		response.Pt = proto.Uint32(state.Pt)
		response.SignCnt = proto.Uint32(state.SignCnt)
		response.SignLastTime = proto.Uint32(state.SignLastTime)
		response.PtStage = proto.Uint32(state.PtStage)
	} else {
		response = buildRefluxInactiveResponse()
	}
	return client.SendMessage(11752, &response)
}
