package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ConfirmSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	if client.Commander == nil {
		return 0, 11610, errors.New("missing commander")
	}
	var payload protobuf.CS_11609
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11610, err
	}
	password := payload.GetPassword()
	if !isSecondaryPasswordValid(password) {
		response := protobuf.SC_11610{Result: proto.Uint32(1)}
		return client.SendMessage(11610, &response)
	}
	settings, err := orm.GetSecondaryPasswordSettings(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11610, err
	}
	if settings.PasswordHash == "" {
		response := protobuf.SC_11610{Result: proto.Uint32(1)}
		return client.SendMessage(11610, &response)
	}
	now := currentUnixTime()
	settings, err = resetSecondaryPasswordLockoutIfExpired(client.Commander.CommanderID, settings, now)
	if err != nil {
		return 0, 11610, err
	}
	if settings.FailCd != nil && now < *settings.FailCd {
		response := protobuf.SC_11610{Result: proto.Uint32(40)}
		return client.SendMessage(11610, &response)
	}
	if !verifySecondaryPassword(settings.PasswordHash, password) {
		settings, err = recordSecondaryPasswordFailure(client.Commander.CommanderID, settings, now)
		if err != nil {
			return 0, 11610, err
		}
		response := protobuf.SC_11610{Result: proto.Uint32(9)}
		return client.SendMessage(11610, &response)
	}
	if err := orm.ResetSecondaryPasswordLockout(orm.GormDB, client.Commander.CommanderID); err != nil {
		return 0, 11610, err
	}
	response := protobuf.SC_11610{Result: proto.Uint32(0)}
	return client.SendMessage(11610, &response)
}
