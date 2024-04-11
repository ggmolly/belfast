package answer

import (
	"fmt"

	"github.com/bettercallmolly/belfast/connection"
	"github.com/bettercallmolly/belfast/logger"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	USER_STATUS_OK     = 0
	USER_STATUS_BANNED = 17
)

var validSC10023 protobuf.SC_10023

func JoinServer(buffer *[]byte, client *connection.Client) (int, int, error) {
	var protoData protobuf.CS_10022
	err := proto.Unmarshal((*buffer), &protoData)
	if err != nil {
		return 0, 10023, err
	}

	err = client.GetCommander(protoData.GetAccountId())
	if err != nil {
		logger.LogEvent("Server", "SC_10023", fmt.Sprintf("failed to fetch commander: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		return 0, 10023, err
	}

	client.Commander.Load()

	if len(client.Commander.Punishments) > 0 {
		if client.Commander.Punishments[0].IsPermanent {
			logger.LogEvent("Database", "Punishments", fmt.Sprintf("Permanent punishment found for uid=%d", protoData.GetAccountId()), logger.LOG_LEVEL_ERROR)
			validSC10023.UserId = proto.Uint32(0)
		} else {
			logger.LogEvent("Database", "Punishments", fmt.Sprintf("Temporary punishment found for uid=%d, lifting at %s", protoData.GetAccountId(), client.Commander.Punishments[0].LiftTimestamp.String()), logger.LOG_LEVEL_INFO)
			validSC10023.UserId = proto.Uint32(uint32(client.Commander.Punishments[0].LiftTimestamp.Unix()))
		}
		validSC10023.Result = proto.Uint32(USER_STATUS_BANNED)
	} else {
		logger.LogEvent("Database", "Punishments", fmt.Sprintf("No punishments found for uid=%d", protoData.GetAccountId()), logger.LOG_LEVEL_INFO)
	}

	return client.SendMessage(10023, &validSC10023)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC10023)
}
