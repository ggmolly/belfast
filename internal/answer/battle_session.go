package answer

import (
	"errors"
	"sync/atomic"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var battleSessionKey uint32

func nextBattleSessionKey() uint32 {
	return atomic.AddUint32(&battleSessionKey, 1)
}

func BeginStage(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_40001
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 40002, err
	}
	key := nextBattleSessionKey()
	session := orm.BattleSession{
		CommanderID: client.Commander.CommanderID,
		System:      payload.GetSystem(),
		StageID:     payload.GetData(),
		Key:         key,
		ShipIDs:     orm.ToInt64List(payload.GetShipIdList()),
	}
	if err := orm.UpsertBattleSession(orm.GormDB, &session); err != nil {
		return 0, 40002, err
	}
	response := protobuf.SC_40002{
		Result:          proto.Uint32(0),
		Key:             proto.Uint32(key),
		DropPerformance: []*protobuf.DROPPERFORMANCE{},
	}
	return client.SendMessage(40002, &response)
}

func FinishStage(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_40003
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 40004, err
	}
	session, err := orm.GetBattleSession(orm.GormDB, client.Commander.CommanderID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 40004, err
	}
	shipIDs := []uint32{}
	if session != nil {
		shipIDs = orm.ToUint32List(session.ShipIDs)
	}
	shipExpList := make([]*protobuf.SHIP_EXP, 0, len(shipIDs))
	for _, shipID := range shipIDs {
		shipExpList = append(shipExpList, &protobuf.SHIP_EXP{
			ShipId: proto.Uint32(shipID),
			// todo: compute ship exp/intimacy/energy from battle statistics
			Exp:      proto.Uint32(0),
			Intimacy: proto.Uint32(0),
			Energy:   proto.Uint32(0),
		})
	}
	mvp := uint32(0)
	if len(shipIDs) > 0 {
		mvp = shipIDs[0]
	}
	if err := orm.DeleteBattleSession(orm.GormDB, client.Commander.CommanderID); err != nil {
		return 0, 40004, err
	}
	response := protobuf.SC_40004{
		Result: proto.Uint32(0),
		// todo: compute player exp from battle statistics
		PlayerExp:   proto.Uint32(0),
		ShipExpList: shipExpList,
		Mvp:         proto.Uint32(mvp),
	}
	return client.SendMessage(40004, &response)
}

func QuitBattle(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_40005
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 40006, err
	}
	if err := orm.DeleteBattleSession(orm.GormDB, client.Commander.CommanderID); err != nil {
		return 0, 40006, err
	}
	response := protobuf.SC_40006{Result: proto.Uint32(0)}
	return client.SendMessage(40006, &response)
}

func DailyQuickBattle(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_40007
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 40008, err
	}
	rewardCount := int(payload.GetCnt())
	rewardList := make([]*protobuf.QUICK_REWARD, 0, rewardCount)
	for i := 0; i < rewardCount; i++ {
		rewardList = append(rewardList, &protobuf.QUICK_REWARD{})
	}
	response := protobuf.SC_40008{
		Result:     proto.Uint32(0),
		RewardList: rewardList,
	}
	return client.SendMessage(40008, &response)
}
