package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type fleetTechGroup struct {
	ID uint32 `json:"id"`
}

type fleetTechTemplate struct {
	Add []any `json:"add"`
}

func TechnologyNationProxy(buffer *[]byte, client *connection.Client) (int, int, error) {
	techGroups, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/fleet_tech_group.json")
	if err != nil {
		return 0, 64000, err
	}
	techTemplates, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/fleet_tech_template.json")
	if err != nil {
		return 0, 64000, err
	}
	response := protobuf.SC_64000{
		TechList:    make([]*protobuf.FLEETTECH, 0, len(techGroups)),
		TechsetList: make([]*protobuf.TECHSET, 0),
	}
	for _, entry := range techGroups {
		var group fleetTechGroup
		if err := json.Unmarshal(entry.Data, &group); err != nil {
			return 0, 64000, err
		}
		response.TechList = append(response.TechList, &protobuf.FLEETTECH{
			GroupId:         proto.Uint32(group.ID),
			EffectTechId:    proto.Uint32(0),
			StudyTechId:     proto.Uint32(0),
			StudyFinishTime: proto.Uint32(0),
			RewardedTech:    proto.Uint32(0),
		})
	}
	for _, entry := range techTemplates {
		var template fleetTechTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return 0, 64000, err
		}
		response.TechsetList = append(response.TechsetList, buildTechSets(template.Add)...)
	}
	return client.SendMessage(64000, &response)
}

func buildTechSets(rawAdd []any) []*protobuf.TECHSET {
	results := make([]*protobuf.TECHSET, 0)
	for _, entry := range rawAdd {
		parts, ok := entry.([]any)
		if !ok || len(parts) < 3 {
			continue
		}
		shipTypes, ok := parts[0].([]any)
		if !ok {
			continue
		}
		attrType, ok := parseJSONUint32(parts[1])
		if !ok {
			continue
		}
		setValue, ok := parseJSONUint32(parts[2])
		if !ok {
			continue
		}
		for _, shipTypeValue := range shipTypes {
			shipType, ok := parseJSONUint32(shipTypeValue)
			if !ok {
				continue
			}
			results = append(results, &protobuf.TECHSET{
				ShipType: proto.Uint32(shipType),
				AttrType: proto.Uint32(attrType),
				SetValue: proto.Uint32(setValue),
			})
		}
	}
	return results
}

func parseJSONUint32(value any) (uint32, bool) {
	number, ok := value.(float64)
	if !ok {
		return 0, false
	}
	return uint32(number), true
}
