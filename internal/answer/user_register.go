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
	rngutil "github.com/ggmolly/belfast/internal/rng"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	registerResultOK             = 0
	registerResultInvalidAccount = 1010
	registerResultAccountExists  = 1011
	registerResultNumericAccount = 1012
	registerResultDatabaseError  = 11
)

var localAccountRand = rngutil.NewLockedRand()

func RegisterAccount(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_10001
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 10002, err
	}
	response := protobuf.SC_10002{Result: proto.Uint32(registerResultOK)}

	account := strings.TrimSpace(payload.GetAccount())
	if account == "" {
		response.Result = proto.Uint32(registerResultInvalidAccount)
		return client.SendMessage(10002, &response)
	}
	if isNumericOnly(account) {
		response.Result = proto.Uint32(registerResultNumericAccount)
		return client.SendMessage(10002, &response)
	}

	var existing orm.LocalAccount
	if err := orm.GormDB.Where("account = ?", account).First(&existing).Error; err == nil {
		response.Result = proto.Uint32(registerResultAccountExists)
		return client.SendMessage(10002, &response)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LogEvent("Server", "SC_10002", fmt.Sprintf("failed to check account: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		response.Result = proto.Uint32(registerResultDatabaseError)
		return client.SendMessage(10002, &response)
	}

	arg2, err := nextLocalArg2(orm.GormDB)
	if err != nil {
		logger.LogEvent("Server", "SC_10002", fmt.Sprintf("failed to allocate arg2: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		response.Result = proto.Uint32(registerResultDatabaseError)
		return client.SendMessage(10002, &response)
	}

	authConfig := auth.NormalizeConfig(config.Current().Auth)
	authConfig.PasswordMinLength = 1
	passwordHash, _, err := auth.HashPassword(payload.GetPassword(), authConfig)
	if err != nil {
		logger.LogEvent("Server", "SC_10002", fmt.Sprintf("failed to hash password: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		response.Result = proto.Uint32(registerResultDatabaseError)
		return client.SendMessage(10002, &response)
	}

	entry := orm.LocalAccount{
		Arg2:     arg2,
		Account:  account,
		Password: passwordHash,
		MailBox:  payload.GetMailBox(),
	}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		logger.LogEvent("Server", "SC_10002", fmt.Sprintf("failed to create account: %s", err.Error()), logger.LOG_LEVEL_ERROR)
		response.Result = proto.Uint32(registerResultDatabaseError)
		return client.SendMessage(10002, &response)
	}

	return client.SendMessage(10002, &response)
}

func isNumericOnly(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func nextLocalArg2(db *gorm.DB) (uint32, error) {
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		candidate := localAccountRand.Uint32()
		if candidate == 0 {
			continue
		}
		if exists, err := localArg2Exists(db, candidate); err != nil {
			return 0, err
		} else if exists {
			continue
		}
		return candidate, nil
	}
	return 0, fmt.Errorf("exhausted arg2 candidates")
}

func localArg2Exists(db *gorm.DB, value uint32) (bool, error) {
	var local orm.LocalAccount
	if err := db.Select("arg2").Where("arg2 = ?", value).First(&local).Error; err == nil {
		return true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	var mapping orm.YostarusMap
	if err := db.Select("arg2").Where("arg2 = ?", value).First(&mapping).Error; err == nil {
		return true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	return false, nil
}
