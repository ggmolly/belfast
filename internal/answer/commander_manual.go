package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CommanderManualInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_22300{
		Handbooks: []*protobuf.TUTHANDBOOK{
			{
				// TODO: Populate with tutorial_handbook_task data per page.
				Id:    proto.Uint32(100101),
				Pt:    proto.Uint32(0),
				Award: proto.Uint32(0),
			},
		},
		FinishedTaskIds: []uint32{},
	}
	return client.SendMessage(22300, &response)
}

func CommanderManualGetTask(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_22302
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 22303, err
	}
	response := protobuf.SC_22303{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(22303, &response)
}

func CommanderManualGetPtAward(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_22304
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 22305, err
	}
	response := protobuf.SC_22305{
		Result:   proto.Uint32(0),
		DropList: []*protobuf.DROPINFO{},
	}
	return client.SendMessage(22305, &response)
}
