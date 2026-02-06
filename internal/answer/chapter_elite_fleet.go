package answer

import (
	"errors"
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const eliteFleetStateField protowire.Number = 1001

// RemoveEliteTargetShip handles CS_13111.
//
// The game uses this packet to remove a specific ship instance (OwnedShip.ID)
// from the commander's stored elite fleet configurations.
func RemoveEliteTargetShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13111
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13112, err
	}
	shipID := payload.GetShipId()

	if client.Commander.OwnedShipsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return 0, 13112, err
		}
	}
	if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
		return 0, 13112, fmt.Errorf("ship not owned: %d", shipID)
	}

	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response := buildSC13112Response([]*protobuf.FLEET_INFO{})
			return client.SendMessage(13112, response)
		}
		return 0, 13112, err
	}
	if len(state.State) == 0 {
		response := buildSC13112Response([]*protobuf.FLEET_INFO{})
		return client.SendMessage(13112, response)
	}

	fleets, err := parseEliteFleetFromState(state.State)
	if err != nil {
		return 0, 13112, err
	}
	updated := removeShipFromFleets(fleets, shipID)
	updatedState, err := setEliteFleetInState(state.State, updated)
	if err != nil {
		return 0, 13112, err
	}
	state.State = updatedState
	if err := orm.UpsertChapterState(orm.GormDB, state); err != nil {
		return 0, 13112, err
	}

	response := buildSC13112Response(updated)
	return client.SendMessage(13112, response)
}

func buildSC13112Response(fleets []*protobuf.FLEET_INFO) *protobuf.SC_13112 {
	if fleets == nil {
		fleets = []*protobuf.FLEET_INFO{}
	}
	return &protobuf.SC_13112{FleetList: fleets}
}

func parseEliteFleetFromState(state []byte) ([]*protobuf.FLEET_INFO, error) {
	if len(state) == 0 {
		return []*protobuf.FLEET_INFO{}, nil
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state, &current); err != nil {
		return nil, err
	}
	unknown := current.ProtoReflect().GetUnknown()
	fleets := []*protobuf.FLEET_INFO{}
	for len(unknown) > 0 {
		num, typ, n := protowire.ConsumeTag(unknown)
		if n < 0 {
			return nil, protowire.ParseError(n)
		}
		unknown = unknown[n:]
		if typ == protowire.BytesType {
			value, m := protowire.ConsumeBytes(unknown)
			if m < 0 {
				return nil, protowire.ParseError(m)
			}
			if num == eliteFleetStateField {
				var fleet protobuf.FLEET_INFO
				if err := proto.Unmarshal(value, &fleet); err != nil {
					return nil, err
				}
				fleets = append(fleets, &fleet)
			}
			unknown = unknown[m:]
			continue
		}
		m, err := skipProtowireValue(num, typ, unknown)
		if err != nil {
			return nil, err
		}
		unknown = unknown[m:]
	}
	return fleets, nil
}

func setEliteFleetInState(state []byte, fleets []*protobuf.FLEET_INFO) ([]byte, error) {
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state, &current); err != nil {
		return nil, err
	}
	unknown := current.ProtoReflect().GetUnknown()
	unknown = filterUnknownField(unknown, eliteFleetStateField)
	for _, fleet := range fleets {
		data, err := proto.Marshal(fleet)
		if err != nil {
			return nil, err
		}
		unknown = protowire.AppendTag(unknown, eliteFleetStateField, protowire.BytesType)
		unknown = protowire.AppendBytes(unknown, data)
	}
	current.ProtoReflect().SetUnknown(unknown)
	return proto.Marshal(&current)
}

func removeShipFromFleets(fleets []*protobuf.FLEET_INFO, shipID uint32) []*protobuf.FLEET_INFO {
	for _, fleet := range fleets {
		removeShipFromTeams(fleet.GetMainTeam(), shipID)
		removeShipFromTeams(fleet.GetSubmarineTeam(), shipID)
		removeShipFromTeams(fleet.GetSupportTeam(), shipID)
	}
	return fleets
}

func removeShipFromTeams(teams []*protobuf.TEAM_INFO, shipID uint32) {
	for _, team := range teams {
		ships := team.GetShipList()
		if len(ships) == 0 {
			continue
		}
		filtered := ships[:0]
		for _, id := range ships {
			if id != shipID {
				filtered = append(filtered, id)
			}
		}
		team.ShipList = filtered
	}
}

func filterUnknownField(b []byte, field protowire.Number) []byte {
	out := []byte{}
	for len(b) > 0 {
		fieldStart := b
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return fieldStart
		}
		b = b[n:]
		valueLen, err := skipProtowireValue(num, typ, b)
		if err != nil {
			return fieldStart
		}
		if num != field {
			out = append(out, fieldStart[:n+valueLen]...)
		}
		b = b[valueLen:]
	}
	return out
}

func skipProtowireValue(num protowire.Number, typ protowire.Type, b []byte) (int, error) {
	switch typ {
	case protowire.VarintType:
		_, n := protowire.ConsumeVarint(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	case protowire.Fixed32Type:
		_, n := protowire.ConsumeFixed32(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	case protowire.Fixed64Type:
		_, n := protowire.ConsumeFixed64(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	case protowire.BytesType:
		_, n := protowire.ConsumeBytes(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	case protowire.StartGroupType:
		_, n := protowire.ConsumeGroup(num, b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("unsupported protowire type: %v", typ)
	}
}
