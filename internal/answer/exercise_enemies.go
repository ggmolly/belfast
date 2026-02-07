package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func ExerciseEnemies(buffer *[]byte, client *connection.Client) (int, int, error) {
	vanguardIDs, mainIDs, err := loadExerciseFleetIDs(client.Commander)
	if err != nil {
		return 0, 18002, err
	}
	response := protobuf.SC_18002{
		Score:               proto.Uint32(0),
		Rank:                proto.Uint32(0),
		FightCount:          proto.Uint32(0),
		FightCountResetTime: proto.Uint32(0),
		FlashTargetCount:    proto.Uint32(0),
		VanguardShipIdList:  vanguardIDs,
		MainShipIdList:      mainIDs,
	}
	return client.SendMessage(18002, &response)
}

func loadExerciseFleetIDs(commander *orm.Commander) ([]uint32, []uint32, error) {
	// Prefer persisted exercise fleet state.
	stored, err := orm.GetExerciseFleet(orm.GormDB, commander.CommanderID)
	if err == nil {
		v := orm.ToUint32List(stored.VanguardShipIDs)
		m := orm.ToUint32List(stored.MainShipIDs)
		if len(v) > 0 && len(m) > 0 && ownsAllShips(commander, v) && ownsAllShips(commander, m) {
			return v, m, nil
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	// Fallback to commander fleet 1.
	fleet1, ok := commander.FleetsMap[1]
	if !ok {
		return []uint32{}, []uint32{}, nil
	}
	shipIDs := orm.ToUint32List(fleet1.ShipList)
	if len(shipIDs) == 0 {
		return []uint32{}, []uint32{}, nil
	}
	vanguard := shipIDs
	if len(vanguard) > 3 {
		vanguard = vanguard[:3]
	}
	main := []uint32{}
	if len(shipIDs) > 3 {
		main = shipIDs[3:]
		if len(main) > 3 {
			main = main[:3]
		}
	}
	return vanguard, main, nil
}

func ownsAllShips(commander *orm.Commander, shipIDs []uint32) bool {
	for _, shipID := range shipIDs {
		if _, ok := commander.OwnedShipsMap[shipID]; !ok {
			return false
		}
	}
	return true
}
