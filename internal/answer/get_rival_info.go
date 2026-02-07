package answer

import (
	"errors"
	"sort"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func GetRivalInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_18104
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 18105, err
	}

	requestedID := payload.GetId()
	commander, err := orm.LoadCommanderWithDetails(requestedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return client.SendMessage(18105, &protobuf.SC_18105{Info: rivalInfoSentinel()})
		}
		return 0, 18105, err
	}

	info := buildRivalTargetInfo(&commander)
	return client.SendMessage(18105, &protobuf.SC_18105{Info: info})
}

func rivalInfoSentinel() *protobuf.TARGETINFO {
	return &protobuf.TARGETINFO{
		Id:    proto.Uint32(0),
		Level: proto.Uint32(0),
		Name:  proto.String(""),
		Score: proto.Uint32(0),
		Rank:  proto.Uint32(0),
		// Repeated fields can be empty/nil; client only checks `info.id == 0`.
		VanguardShipList: nil,
		MainShipList:     nil,
		Display:          nil,
	}
}

func buildRivalTargetInfo(commander *orm.Commander) *protobuf.TARGETINFO {
	display := &protobuf.DISPLAYINFO{
		Icon:          proto.Uint32(commander.DisplayIconID),
		Skin:          proto.Uint32(commander.DisplaySkinID),
		IconFrame:     proto.Uint32(commander.SelectedIconFrameID),
		ChatFrame:     proto.Uint32(commander.SelectedChatFrameID),
		IconTheme:     proto.Uint32(commander.DisplayIconThemeID),
		MarryFlag:     proto.Uint32(0),
		TransformFlag: proto.Uint32(0),
	}

	secretaries := commander.GetSecretaries()
	if display.GetIcon() == 0 && len(secretaries) > 0 {
		display.Icon = proto.Uint32(secretaries[0].ShipID)
	}
	if display.GetSkin() == 0 && len(secretaries) > 0 {
		display.Skin = proto.Uint32(secretaries[0].SkinID)
	}

	vanguardShips, mainShips := buildRivalDefenseShipLists(commander)

	return &protobuf.TARGETINFO{
		Id:               proto.Uint32(commander.CommanderID),
		Level:            proto.Uint32(uint32(commander.Level)),
		Name:             proto.String(commander.Name),
		Score:            proto.Uint32(0),
		Rank:             proto.Uint32(0),
		VanguardShipList: vanguardShips,
		MainShipList:     mainShips,
		Display:          display,
	}
}

func buildRivalDefenseShipLists(commander *orm.Commander) ([]*protobuf.SHIPINFO, []*protobuf.SHIPINFO) {
	shipsByID := make(map[uint32]*orm.OwnedShip, len(commander.Ships))
	for i := range commander.Ships {
		ship := &commander.Ships[i]
		shipsByID[ship.ID] = ship
	}

	fleetIDs := commanderFleetShipIDs(commander)
	preferredIDs := fleetIDs
	if len(preferredIDs) == 0 {
		preferredIDs = stableOwnedShipIDs(commander)
	}

	candidates := make([]*orm.OwnedShip, 0, 6)
	appendCandidates := func(ids []uint32) {
		for _, id := range ids {
			ship, ok := shipsByID[id]
			if !ok {
				continue
			}
			candidates = append(candidates, ship)
			if len(candidates) >= 6 {
				break
			}
		}
	}
	appendCandidates(preferredIDs)
	if len(candidates) == 0 && len(fleetIDs) > 0 {
		// Fleet references can get stale (ships retired/deleted); fall back to stable owned ships.
		appendCandidates(stableOwnedShipIDs(commander))
	}

	vanguard := make([]*protobuf.SHIPINFO, 0, 3)
	main := make([]*protobuf.SHIPINFO, 0, 3)
	used := make(map[uint32]bool, 6)

	for _, ship := range candidates {
		if len(vanguard) < 3 && isVanguardShipType(uint32(ship.Ship.Type)) {
			vanguard = append(vanguard, orm.ToProtoOwnedShip(*ship, nil, nil))
			used[ship.ID] = true
			continue
		}
		if len(main) < 3 {
			main = append(main, orm.ToProtoOwnedShip(*ship, nil, nil))
			used[ship.ID] = true
		}
	}

	// Fill remaining slots (avoid empty lists when we have ships).
	for _, ship := range candidates {
		if used[ship.ID] {
			continue
		}
		if len(vanguard) < 3 {
			vanguard = append(vanguard, orm.ToProtoOwnedShip(*ship, nil, nil))
			used[ship.ID] = true
			continue
		}
		if len(main) < 3 {
			main = append(main, orm.ToProtoOwnedShip(*ship, nil, nil))
			used[ship.ID] = true
		}
		if len(vanguard) >= 3 && len(main) >= 3 {
			break
		}
	}

	return vanguard, main
}

func commanderFleetShipIDs(commander *orm.Commander) []uint32 {
	for i := range commander.Fleets {
		fleet := commander.Fleets[i]
		if fleet.GameID != 1 {
			continue
		}
		result := make([]uint32, 0, len(fleet.ShipList))
		for _, id := range fleet.ShipList {
			result = append(result, uint32(id))
		}
		return result
	}
	return nil
}

func stableOwnedShipIDs(commander *orm.Commander) []uint32 {
	ships := make([]*orm.OwnedShip, 0, len(commander.Ships))
	for i := range commander.Ships {
		ships = append(ships, &commander.Ships[i])
	}
	sort.Slice(ships, func(i, j int) bool {
		return ships[i].ID < ships[j].ID
	})
	result := make([]uint32, 0, len(ships))
	for _, ship := range ships {
		result = append(result, ship.ID)
	}
	return result
}

func isVanguardShipType(shipType uint32) bool {
	switch shipType {
	case 1, 2, 3, 18, 19: // DD/CL/CA/CB/AE
		return true
	case 8, 17: // submarines/sub carriers (kept in vanguard list for now)
		return true
	default:
		return false
	}
}
