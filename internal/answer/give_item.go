package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GiveItem(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11202
	err := proto.Unmarshal(*buffer, &payload)
	if err != nil {
		return 0, 11203, err
	}
	logger.LogEvent("Server", "AskItem",
		fmt.Sprintf(
			"uid=%d asked for activity_id=%d",
			client.Commander.CommanderID,
			*payload.ActivityId,
		), logger.LOG_LEVEL_DEBUG)

	// Answer with random stuff rn
	response := protobuf.SC_11203{
		Result: proto.Uint32(0), // 0 = success, otherwise the game will show an error (see: TipsMgr)
		AwardList: []*protobuf.DROPINFO{
			{
				Type:   proto.Uint32(2),
				Id:     proto.Uint32(20001),
				Number: proto.Uint32(99), // 99x 20001 -> Wisdom Cube
			},
		},
		Build:          nil,          // not building a shipgirl
		Number:         []uint32{99}, // why? what is this?
		ReturnUserList: []*protobuf.RETURN_USER_INFO{},
		// InsMessage:     &protobuf.INS_MESSAGE{}, // Here you can simulate a message from a shipgirl (?)
		CollectionList: nil,
		TaskList:       nil,
	}
	return client.SendMessage(11203, &response)
}
