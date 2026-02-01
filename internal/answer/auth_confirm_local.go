package answer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	localLoginResultOK             = 0
	localLoginResultInvalidAccount = 1010
	localLoginResultWrongPassword  = 1020
	localLoginResultDatabaseError  = 11
)

func HandleLocalLogin(payload *protobuf.CS_10020, client *connection.Client) (int, int, error) {
	account := strings.TrimSpace(payload.GetArg1())
	if account == "" {
		return sendLocalLoginFailure(client, localLoginResultInvalidAccount)
	}

	var local orm.LocalAccount
	if err := orm.GormDB.Where("account = ?", account).First(&local).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendLocalLoginFailure(client, localLoginResultInvalidAccount)
		}
		logger.LogEvent("Server", "SC_10021", fmt.Sprintf("failed to fetch local account: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		return sendLocalLoginFailure(client, localLoginResultDatabaseError)
	}

	password := payload.GetArg2()
	valid, err := auth.VerifyPassword(password, local.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidHash) {
			if local.Password != password {
				return sendLocalLoginFailure(client, localLoginResultWrongPassword)
			}
			authConfig := auth.NormalizeConfig(config.Current().Auth)
			authConfig.PasswordMinLength = 1
			passwordHash, _, hashErr := auth.HashPassword(password, authConfig)
			if hashErr != nil {
				logger.LogEvent("Server", "SC_10021", fmt.Sprintf("failed to hash legacy password: %s", hashErr.Error()), logger.LOG_LEVEL_ERROR)
				return sendLocalLoginFailure(client, localLoginResultDatabaseError)
			}
			if updateErr := orm.GormDB.Model(&local).Update("password", passwordHash).Error; updateErr != nil {
				logger.LogEvent("Server", "SC_10021", fmt.Sprintf("failed to update password: %s", updateErr.Error()), logger.LOG_LEVEL_ERROR)
				return sendLocalLoginFailure(client, localLoginResultDatabaseError)
			}
			valid = true
		} else {
			logger.LogEvent("Server", "SC_10021", fmt.Sprintf("failed to verify password: %s", err.Error()), logger.LOG_LEVEL_ERROR)
			return sendLocalLoginFailure(client, localLoginResultDatabaseError)
		}
	}
	if !valid {
		return sendLocalLoginFailure(client, localLoginResultWrongPassword)
	}

	client.AuthArg2 = local.Arg2
	response := protobuf.SC_10021{
		Result:       proto.Uint32(localLoginResultOK),
		AccountId:    proto.Uint32(0),
		ServerTicket: proto.String(formatServerTicket(client.AuthArg2)),
		Device:       proto.Uint32(0),
	}

	var mapping orm.YostarusMap
	if err := orm.GormDB.Where("arg2 = ?", local.Arg2).First(&mapping).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if config.Current().CreatePlayer.SkipOnboarding {
				accountID, err := client.CreateCommander(local.Arg2)
				if err != nil {
					logger.LogEvent("Server", "SC_10021", fmt.Sprintf("failed to create commander: %s", err.Error()), logger.LOG_LEVEL_ERROR)
					response.Result = proto.Uint32(localLoginResultDatabaseError)
					return client.SendMessage(10021, &response)
				}
				response.AccountId = proto.Uint32(accountID)
			}
		} else {
			logger.LogEvent("Server", "SC_10021", fmt.Sprintf("failed to fetch account mapping: %s", err.Error()), logger.LOG_LEVEL_ERROR)
			response.Result = proto.Uint32(localLoginResultDatabaseError)
			return client.SendMessage(10021, &response)
		}
	} else {
		response.AccountId = proto.Uint32(mapping.AccountID)
	}

	updateServerList(config.Current().Servers)
	response.Serverlist = Servers
	logger.LogEvent("Server", "SC_10021", fmt.Sprintf("sending %d servers", len(response.Serverlist)), logger.LOG_LEVEL_WARN)
	return client.SendMessage(10021, &response)
}

func sendLocalLoginFailure(client *connection.Client, result uint32) (int, int, error) {
	response := protobuf.SC_10021{
		Result:       proto.Uint32(result),
		AccountId:    proto.Uint32(0),
		ServerTicket: proto.String(formatServerTicket(0)),
		Device:       proto.Uint32(0),
	}
	updateServerList(config.Current().Servers)
	response.Serverlist = Servers
	return client.SendMessage(10021, &response)
}
