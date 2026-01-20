package answer

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	createPlayerNameMin = 4
	createPlayerNameMax = 14
)

var starterShipIDs = map[uint32]struct{}{
	101171: {}, // Laffey
	201211: {}, // Javelin
	401231: {}, // Z23
}

func CreateNewPlayer(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_10024
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 10025, err
	}

	response := protobuf.SC_10025{
		Result: proto.Uint32(0),
		UserId: proto.Uint32(0),
	}

	nickname := payload.GetNickName()
	deviceID := payload.GetDeviceId()
	if deviceID == "" {
		// device id is required to bind future connections to this account
		response.Result = proto.Uint32(1)
		return client.SendMessage(10025, &response)
	}
	nameLength := utf8.RuneCountInString(nickname)
	if nameLength < createPlayerNameMin {
		response.Result = proto.Uint32(2012)
		return client.SendMessage(10025, &response)
	}
	if nameLength > createPlayerNameMax {
		response.Result = proto.Uint32(2011)
		return client.SendMessage(10025, &response)
	}

	createConfig := config.Current().CreatePlayer
	if len(createConfig.NameBlacklist) > 0 {
		lowerName := strings.ToLower(nickname)
		for _, blocked := range createConfig.NameBlacklist {
			blocked = strings.TrimSpace(blocked)
			if blocked == "" {
				continue
			}
			if strings.Contains(lowerName, strings.ToLower(blocked)) {
				response.Result = proto.Uint32(2013)
				return client.SendMessage(10025, &response)
			}
		}
	}
	if createConfig.NameIllegalPattern != "" {
		matcher, err := regexp.Compile(createConfig.NameIllegalPattern)
		if err != nil {
			return 0, 10025, err
		}
		if matcher.MatchString(nickname) {
			response.Result = proto.Uint32(2014)
			return client.SendMessage(10025, &response)
		}
	}

	shipID := payload.GetShipId()
	if _, ok := starterShipIDs[shipID]; !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(10025, &response)
	}

	// allow account binding across different connections using device id
	var deviceMapping orm.DeviceAuthMap
	if err := orm.GormDB.Where("device_id = ?", deviceID).First(&deviceMapping).Error; err == nil {
		if deviceMapping.AccountID != 0 {
			response.Result = proto.Uint32(1011)
			return client.SendMessage(10025, &response)
		}
		if client.AuthArg2 == 0 {
			client.AuthArg2 = deviceMapping.Arg2
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LogEvent("Server", "SC_10025", fmt.Sprintf("failed to fetch device mapping: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		response.Result = proto.Uint32(18)
		return client.SendMessage(10025, &response)
	}

	if client.AuthArg2 == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(10025, &response)
	}

	var mapping orm.YostarusMap
	if err := orm.GormDB.Where("arg2 = ?", client.AuthArg2).First(&mapping).Error; err == nil {
		response.Result = proto.Uint32(1011)
		return client.SendMessage(10025, &response)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LogEvent("Server", "SC_10025", fmt.Sprintf("failed to fetch account mapping: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		response.Result = proto.Uint32(18)
		return client.SendMessage(10025, &response)
	}

	var existingCommander orm.Commander
	if err := orm.GormDB.Where("name = ?", nickname).First(&existingCommander).Error; err == nil {
		response.Result = proto.Uint32(2015)
		return client.SendMessage(10025, &response)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LogEvent("Server", "SC_10025", fmt.Sprintf("failed to check commander name: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		response.Result = proto.Uint32(18)
		return client.SendMessage(10025, &response)
	}

	accountID, err := client.CreateCommanderWithStarter(client.AuthArg2, nickname, shipID)
	if err != nil {
		logger.LogEvent("Server", "SC_10025", fmt.Sprintf("failed to create commander: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		response.Result = proto.Uint32(18)
		return client.SendMessage(10025, &response)
	}
	if err := orm.GormDB.Save(&orm.DeviceAuthMap{
		DeviceID:  deviceID,
		Arg2:      client.AuthArg2,
		AccountID: accountID,
	}).Error; err != nil {
		logger.LogEvent("Server", "SC_10025", fmt.Sprintf("failed to save device mapping: %s", err.Error()), logger.LOG_LEVEL_ERROR)
	}

	response.UserId = proto.Uint32(accountID)
	return client.SendMessage(10025, &response)
}
