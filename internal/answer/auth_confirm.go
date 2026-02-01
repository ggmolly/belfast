package answer

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var protoValidAnswer protobuf.SC_10021

func updateServerList(servers []config.ServerConfig) {
	statuses := getServerStatusCache(servers)
	Servers = buildServerInfo(servers, statuses)
	protoValidAnswer.Serverlist = Servers
}

func Forge_SC10021(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_10020
	err := proto.Unmarshal(*buffer, &payload)
	if err != nil {
		return 0, 10021, fmt.Errorf("failed to unmarshal payload: %s", err.Error())
	}
	if payload.GetLoginType() == 2 {
		return HandleLocalLogin(&payload, client)
	}

	var yostarusAuth orm.YostarusMap
	intArg2, err := strconv.Atoi(payload.GetArg2())
	if err != nil {
		return 0, 10021, fmt.Errorf("failed to convert arg2 to int: %s", err.Error())
	}
	client.AuthArg2 = uint32(intArg2)
	protoValidAnswer.ServerTicket = proto.String(formatServerTicket(client.AuthArg2))

	err = orm.GormDB.
		Where("arg2 = ?", intArg2).
		First(&yostarusAuth).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if config.Current().CreatePlayer.SkipOnboarding {
				// skip onboarding by creating the account on auth
				accountID, err := client.CreateCommander(uint32(intArg2))
				if err != nil {
					logger.LogEvent("Server", "SC_10021", fmt.Sprintf("failed to create commander: %s", err.Error()), logger.LOG_LEVEL_ERROR)
					return 0, 10021, err
				}
				protoValidAnswer.AccountId = proto.Uint32(accountID)
			} else {
				protoValidAnswer.AccountId = proto.Uint32(0) // CS_10024 handles account creation.
			}
		} else {
			logger.LogEvent("Server", "SC_10021", fmt.Sprintf("failed to fetch account for arg2 %d: %s", intArg2, err.Error()), logger.LOG_LEVEL_ERROR)
			return 0, 10021, err
		}
	} else {
		protoValidAnswer.AccountId = proto.Uint32(yostarusAuth.AccountID)
	}

	// Update server list
	updateServerList(config.Current().Servers)
	logger.LogEvent("Server", "SC_10021", fmt.Sprintf("sending %d servers", len(protoValidAnswer.Serverlist)), logger.LOG_LEVEL_WARN)
	return client.SendMessage(10021, &protoValidAnswer)
}

func init() {
	protoValidAnswer.ServerTicket = proto.String("=*=*=*=BELFAST=*=*=*=")
	protoValidAnswer.Serverlist = Servers
	protoValidAnswer.Result = proto.Uint32(0)
	protoValidAnswer.Device = proto.Uint32(0)
}
