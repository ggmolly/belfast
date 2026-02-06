package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ReqPlayerAssistShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12301
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12301, err
	}

	// `type` is currently not used server-side; client expects index-aligned responses.
	_ = data.GetType()

	response := protobuf.SC_12302{
		ShipList: make([]*protobuf.SHIPINFO, len(data.GetIdList())),
	}
	for i := range data.GetIdList() {
		response.ShipList[i] = blankAssistShipInfo()
	}

	return client.SendMessage(12302, &response)
}

func blankAssistShipInfo() *protobuf.SHIPINFO {
	return &protobuf.SHIPINFO{
		Id:         proto.Uint32(0),
		TemplateId: proto.Uint32(0),
		Level:      proto.Uint32(0),
		Exp:        proto.Uint32(0),
		Energy:     proto.Uint32(0),
		State: &protobuf.SHIPSTATE{
			State: proto.Uint32(0),
		},
		IsLocked:    proto.Uint32(0),
		Intimacy:    proto.Uint32(0),
		Proficiency: proto.Uint32(0),
		CreateTime:  proto.Uint32(0),
		SkinId:      proto.Uint32(0),
		Propose:     proto.Uint32(0),
		MaxLevel:    proto.Uint32(0),
		ActivityNpc: proto.Uint32(0),
	}
}
