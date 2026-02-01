package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func SendCmd(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11100
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11101, err
	}

	cmd := payload.GetCmd()
	response := protobuf.SC_11101{
		Result: proto.Uint32(1),
	}

	switch cmd {
	case "into":
		response.Result = proto.Uint32(0)
		response.Msg = proto.String("CMD:into Result:ok")
	case "world":
		if payload.GetArg1() == "reset" {
			response.Result = proto.Uint32(0)
			response.Msg = proto.String("CMD:world Result:ok")
		} else {
			response.Msg = proto.String(fmt.Sprintf("CMD:%s Result:fail", cmd))
		}
	case "kick":
		response.Result = proto.Uint32(0)
		response.Msg = proto.String("CMD:kick Result:ok")
		sentBytes, packetId, err := client.SendMessage(11101, &response)
		if err != nil {
			return sentBytes, packetId, err
		}
		if err := client.Disconnect(consts.DR_CONNECTION_TO_SERVER_LOST); err != nil {
			return sentBytes, packetId, err
		}
		return sentBytes, packetId, nil
	default:
		response.Msg = proto.String(fmt.Sprintf("CMD:%s Result:fail", cmd))
	}

	return client.SendMessage(11101, &response)
}
