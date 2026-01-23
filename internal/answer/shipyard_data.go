package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type blueprintTemplate struct {
	ID     uint32 `json:"id"`
	ShipID uint32 `json:"ship_id"`
}

func ShipyardData(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_63100{
		ColdTime:                 proto.Uint32(0),
		DailyCatchupStrengthen:   proto.Uint32(0),
		DailyCatchupStrengthenUr: proto.Uint32(0),
	}
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/ship_data_blueprint.json")
	if err != nil {
		return 0, 63100, err
	}
	response.BlueprintList = make([]*protobuf.BLUPRINTINFO, 0, len(entries))
	for _, entry := range entries {
		var template blueprintTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return 0, 63100, err
		}
		shipID := template.ShipID
		if shipID == 0 {
			shipID = template.ID
		}
		response.BlueprintList = append(response.BlueprintList, &protobuf.BLUPRINTINFO{
			Id:             proto.Uint32(template.ID),
			ShipId:         proto.Uint32(shipID),
			StartTime:      proto.Uint32(0),
			BluePrintLevel: proto.Uint32(0),
			Exp:            proto.Uint32(0),
			StartDuration:  proto.Uint32(0),
		})
	}
	return client.SendMessage(63100, &response)
}
