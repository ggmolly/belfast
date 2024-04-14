package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/logger"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	USER_STATUS_OK     = 0
	USER_STATUS_BANNED = 17
)

func JoinServer(buffer *[]byte, client *connection.Client) (int, int, error) {
	var protoData protobuf.CS_10022
	err := proto.Unmarshal((*buffer), &protoData)
	if err != nil {
		return 0, 10023, err
	}

	response := protobuf.SC_10023{
		Result:       proto.Uint32(0),
		ServerTicket: proto.String("=*=*=*=BELFAST=*=*=*="),
		ServerLoad:   proto.Uint32(0),
		DbLoad:       proto.Uint32(0),
	}

	err = client.GetCommander(protoData.GetAccountId())
	if err != nil {
		logger.LogEvent("Server", "SC_10023", fmt.Sprintf("failed to fetch commander (id=%d): %s", protoData.GetAccountId(), err.Error()), logger.LOG_LEVEL_ERROR)
		return 0, 10023, err
	}

	client.Commander.Load()

	if len(client.Commander.Punishments) > 0 {
		if client.Commander.Punishments[0].IsPermanent {
			logger.LogEvent("Database", "Punishments", fmt.Sprintf("Permanent punishment found for uid=%d", protoData.GetAccountId()), logger.LOG_LEVEL_ERROR)
			response.UserId = proto.Uint32(0)
		} else {
			logger.LogEvent("Database", "Punishments", fmt.Sprintf("Temporary punishment found for uid=%d, lifting at %s", protoData.GetAccountId(), client.Commander.Punishments[0].LiftTimestamp.String()), logger.LOG_LEVEL_INFO)
			response.UserId = proto.Uint32(uint32(client.Commander.Punishments[0].LiftTimestamp.Unix()))
		}
		response.Result = proto.Uint32(USER_STATUS_BANNED)
	} else {
		logger.LogEvent("Database", "Punishments", fmt.Sprintf("No punishments found for uid=%d", protoData.GetAccountId()), logger.LOG_LEVEL_INFO)
		response.UserId = proto.Uint32(client.Commander.CommanderID)
	}

	return client.SendMessage(10023, &response)
}
