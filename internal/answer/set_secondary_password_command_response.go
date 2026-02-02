package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func SetSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11605
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11606, err
	}
	response := protobuf.SC_11606{Result: proto.Uint32(0)}
	if !isSecondaryPasswordValid(payload.GetPassword()) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11606, &response)
	}
	state, err := orm.GetOrCreateSecondaryPasswordState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11606, err
	}
	if state.State > 0 || state.PasswordHash != "" {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11606, &response)
	}
	hash, err := hashSecondaryPassword(payload.GetPassword())
	if err != nil {
		return 0, 11606, err
	}
	systemList := sanitizeSecondarySystemList(payload.GetSystemList())
	state.PasswordHash = hash
	state.Notice = payload.GetNotice()
	state.SystemList = orm.ToInt64List(systemList)
	state.State = 1
	state.FailCount = 0
	state.FailCd = 0
	if err := orm.SaveSecondaryPasswordState(orm.GormDB, state); err != nil {
		return 0, 11606, err
	}
	return client.SendMessage(11606, &response)
}
