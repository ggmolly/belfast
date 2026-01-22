package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ShipBuild(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12002
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 12002, err
	}
	// Count = amount of ships to build
	// Costtype = 0 = wisdom cube, 1 = tickets
	var response protobuf.SC_12003

	var goldCost uint32
	var cubeCost uint32
	switch data.GetId() {
	case 2: // Light
		goldCost = 600
		cubeCost = 1
	case 3: // Heavy
		goldCost = 1500
		cubeCost = 3
	case 1: // Special
		goldCost = 1500
		cubeCost = 3
	default:
		response.Result = proto.Uint32(1) // unknown pool id
		return client.SendMessage(12003, &response)
	}

	goldCost *= data.GetCount()
	cubeCost *= data.GetCount()

	// Check if player has enough gold / cubes
	if data.GetCosttype() == 0 {
		if !client.Commander.HasEnoughGold(goldCost) {
			response.Result = proto.Uint32(2) // not enough gold
			return client.SendMessage(12003, &response)
		}
		if !client.Commander.HasEnoughCube(cubeCost) {
			response.Result = proto.Uint32(3) // not enough cubes
			return client.SendMessage(12003, &response)
		}
	}

	response.BuildInfo = make([]*protobuf.BUILDINFO, data.GetCount())
	runningBuilds := len(client.Commander.Builds)
	// We have to get the number of builds that are running and add it to the build id
	// this is because the game expects the build id to be sequential
	for i := 0; i < int(data.GetCount()); i++ {
		build, buildTime, err := client.Commander.CreateBuild(data.GetId(), &runningBuilds)
		if err != nil {
			return 0, 12003, err
		}
		response.BuildInfo[i] = orm.ToProtoBuildInfo(orm.BuildInfoPayload{
			Build:     build,
			PoolID:    build.PoolID,
			BuildTime: buildTime,
		})
	}
	response.Result = proto.Uint32(0)
	client.Commander.ConsumeItem(20001, cubeCost) // consume cubes
	client.Commander.ConsumeResource(1, goldCost) // consume gold
	if err := client.Commander.IncrementDrawCount(data.GetCount()); err != nil {
		return 0, 12003, err
	}
	return client.SendMessage(12003, &response)
}
