package answer

import (
	"errors"
	"fmt"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
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

	accountID := protoData.GetAccountId()
	deviceID := protoData.GetDeviceId()
	if deviceID != "" {
		// try to recover account identity when the client sends account_id = 0
		var deviceMapping orm.DeviceAuthMap
		if err := orm.GormDB.Where("device_id = ?", deviceID).First(&deviceMapping).Error; err == nil {
			if client.AuthArg2 == 0 {
				client.AuthArg2 = deviceMapping.Arg2
			}
			if accountID == 0 && deviceMapping.AccountID != 0 {
				accountID = deviceMapping.AccountID
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LogEvent("Server", "SC_10023", fmt.Sprintf("failed to fetch device mapping: %s", err.Error()), logger.LOG_LEVEL_ERROR)
			return 0, 10023, err
		}
	}
	if accountID == 0 {
		if client.AuthArg2 == 0 {
			client.AuthArg2 = parseServerTicket(protoData.GetServerTicket())
		}
		if client.AuthArg2 != 0 {
			var mapping orm.YostarusMap
			if err := orm.GormDB.Where("arg2 = ?", client.AuthArg2).First(&mapping).Error; err == nil {
				accountID = mapping.AccountID
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				logger.LogEvent("Server", "SC_10023", fmt.Sprintf("failed to fetch account mapping: %s", err.Error()), logger.LOG_LEVEL_ERROR)
				return 0, 10023, err
			}
		}
		if accountID == 0 && config.Current().CreatePlayer.SkipOnboarding && client.AuthArg2 != 0 {
			createdID, err := client.CreateCommander(client.AuthArg2)
			if err != nil {
				return 0, 10023, err
			}
			accountID = createdID
		}
		if accountID == 0 {
			if deviceID != "" && client.AuthArg2 != 0 {
				if err := orm.GormDB.Save(&orm.DeviceAuthMap{
					DeviceID:  deviceID,
					Arg2:      client.AuthArg2,
					AccountID: 0,
				}).Error; err != nil {
					logger.LogEvent("Server", "SC_10023", fmt.Sprintf("failed to save device mapping: %s", err.Error()), logger.LOG_LEVEL_ERROR)
				}
			}
			response.UserId = proto.Uint32(0) // CS_10024 handles account creation.
			return client.SendMessage(10023, &response)
		}
	}

	err = client.GetCommander(accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.UserId = proto.Uint32(0) // CS_10024 handles account creation.
			return client.SendMessage(10023, &response)
		}

		logger.LogEvent("Server", "SC_10023", fmt.Sprintf("failed to fetch commander (id=%d): %s", accountID, err.Error()), logger.LOG_LEVEL_ERROR)
		return 0, 10023, err
	}

	if err := client.Commander.Load(); err != nil {
		logger.LogEvent("Server", "SC_10023", fmt.Sprintf("failed to load commander (id=%d): %s", accountID, err.Error()), logger.LOG_LEVEL_ERROR)
		return 0, 10023, err
	}

	if client.Server != nil {
		existingKicked := client.Server.DisconnectCommander(
			client.Commander.CommanderID,
			consts.DR_LOGGED_IN_ON_ANOTHER_DEVICE,
			client,
		)
		if existingKicked {
			logger.LogEvent("Server", "LoginKick",
				fmt.Sprintf("kicked previous session for commander %d", client.Commander.CommanderID),
				logger.LOG_LEVEL_INFO)
		}
	}

	if len(client.Commander.Punishments) > 0 {
		active := client.Commander.Punishments[0]
		if active.IsPermanent || active.LiftTimestamp == nil {
			logger.LogEvent("Database", "Punishments", fmt.Sprintf("Permanent punishment found for uid=%d", accountID), logger.LOG_LEVEL_ERROR)
			response.UserId = proto.Uint32(0)
		} else {
			logger.LogEvent("Database", "Punishments", fmt.Sprintf("Temporary punishment found for uid=%d, lifting at %s", accountID, active.LiftTimestamp.String()), logger.LOG_LEVEL_INFO)
			response.UserId = proto.Uint32(uint32(active.LiftTimestamp.Unix()))
		}
		response.Result = proto.Uint32(USER_STATUS_BANNED)
	} else {
		logger.LogEvent("Database", "Punishments", fmt.Sprintf("No punishments found for uid=%d", accountID), logger.LOG_LEVEL_INFO)
		response.Result = proto.Uint32(USER_STATUS_OK)
		response.UserId = proto.Uint32(client.Commander.CommanderID)
	}

	if deviceID != "" && client.AuthArg2 != 0 {
		if err := orm.GormDB.Save(&orm.DeviceAuthMap{
			DeviceID:  deviceID,
			Arg2:      client.AuthArg2,
			AccountID: accountID,
		}).Error; err != nil {
			logger.LogEvent("Server", "SC_10023", fmt.Sprintf("failed to save device mapping: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		}
	}

	return client.SendMessage(10023, &response)
}
