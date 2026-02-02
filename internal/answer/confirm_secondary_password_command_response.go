package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ConfirmSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11609
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11610, err
	}
	response := protobuf.SC_11610{Result: proto.Uint32(0)}
	state, err := orm.GetOrCreateSecondaryPasswordState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11610, err
	}
	if state.State == 0 || state.PasswordHash == "" {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11610, &response)
	}
	now := uint32(time.Now().Unix())
	if secondaryPasswordLocked(state, now) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11610, &response)
	}
	valid, err := verifySecondaryPassword(payload.GetPassword(), state.PasswordHash)
	if err != nil {
		return 0, 11610, err
	}
	if !valid {
		applySecondaryPasswordFailure(state, now)
		if err := orm.SaveSecondaryPasswordState(orm.GormDB, state); err != nil {
			return 0, 11610, err
		}
		response.Result = proto.Uint32(9)
		return client.SendMessage(11610, &response)
	}
	state.State = 2
	state.FailCount = 0
	state.FailCd = 0
	if err := orm.SaveSecondaryPasswordState(orm.GormDB, state); err != nil {
		return 0, 11610, err
	}
	return client.SendMessage(11610, &response)
}
