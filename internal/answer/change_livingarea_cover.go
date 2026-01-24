package answer

import (
	"strconv"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChangeLivingAreaCover(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11030
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11031, err
	}
	response := protobuf.SC_11031{Result: proto.Uint32(0)}
	coverID := payload.GetLivingareaCoverId()
	if coverID != 0 {
		if _, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/livingarea_cover.json", strconv.FormatUint(uint64(coverID), 10)); err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11031, &response)
		}
		owned, err := orm.CommanderHasLivingAreaCover(client.Commander.CommanderID, coverID)
		if err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11031, &response)
		}
		if !owned {
			response.Result = proto.Uint32(2)
			return client.SendMessage(11031, &response)
		}
	}
	client.Commander.LivingAreaCoverID = coverID
	if err := orm.GormDB.Model(client.Commander).Update("living_area_cover_id", coverID).Error; err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11031, &response)
}
