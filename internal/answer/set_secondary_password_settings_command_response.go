package answer

import (
	"errors"
	"sort"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func SetSecondaryPasswordSettingsCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	if client.Commander == nil {
		return 0, 11608, errors.New("missing commander")
	}
	var payload protobuf.CS_11607
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11608, err
	}
	password := payload.GetPassword()
	if !isSecondaryPasswordValid(password) {
		response := protobuf.SC_11608{Result: proto.Uint32(1)}
		return client.SendMessage(11608, &response)
	}
	settings, err := orm.GetSecondaryPasswordSettings(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11608, err
	}
	if settings.PasswordHash == "" {
		response := protobuf.SC_11608{Result: proto.Uint32(1)}
		return client.SendMessage(11608, &response)
	}
	now := currentUnixTime()
	settings, err = resetSecondaryPasswordLockoutIfExpired(client.Commander.CommanderID, settings, now)
	if err != nil {
		return 0, 11608, err
	}
	if settings.FailCd != nil && now < *settings.FailCd {
		response := protobuf.SC_11608{Result: proto.Uint32(40)}
		return client.SendMessage(11608, &response)
	}
	if !verifySecondaryPassword(settings.PasswordHash, password) {
		settings, err = recordSecondaryPasswordFailure(client.Commander.CommanderID, settings, now)
		if err != nil {
			return 0, 11608, err
		}
		response := protobuf.SC_11608{Result: proto.Uint32(9)}
		return client.SendMessage(11608, &response)
	}
	list := payload.GetSystemList()
	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})
	if len(list) == 0 {
		settings.PasswordHash = ""
		settings.Notice = ""
		settings.SystemList = orm.Int64List{}
		settings.FailCount = 0
		settings.FailCd = nil
	} else {
		settings.SystemList = orm.ToInt64List(list)
		settings.FailCount = 0
		settings.FailCd = nil
	}
	if err := orm.UpsertSecondaryPasswordSettings(orm.GormDB, settings); err != nil {
		return 0, 11608, err
	}
	response := protobuf.SC_11608{Result: proto.Uint32(0)}
	return client.SendMessage(11608, &response)
}
