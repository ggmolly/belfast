package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func ResourcesInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_22001
	response.OilWellLevel = proto.Uint32(1)
	response.OilWellLvUpTime = proto.Uint32(1)
	response.GoldWellLevel = proto.Uint32(1)
	response.GoldWellLvUpTime = proto.Uint32(1)
	response.ClassLv = proto.Uint32(1)
	response.ClassLvUpTime = proto.Uint32(1)
	response.SkillClassNum = proto.Uint32(1)
	response.DailyFinishBuffCnt = proto.Uint32(1)
	response.Class = &protobuf.NAVALACADEMY_CLASS{
		Proficiency: proto.Uint32(0),
	}
	return client.SendMessage(22001, &response)
}
