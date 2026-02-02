package answer

import (
	"errors"
	"sort"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func SetSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	if client.Commander == nil {
		return 0, 11606, errors.New("missing commander")
	}
	var payload protobuf.CS_11605
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11606, err
	}
	password := payload.GetPassword()
	settings := payload.GetSystemList()
	if !isSecondaryPasswordValid(password) || len(settings) == 0 {
		response := protobuf.SC_11606{Result: proto.Uint32(1)}
		return client.SendMessage(11606, &response)
	}
	sort.Slice(settings, func(i, j int) bool {
		return settings[i] < settings[j]
	})
	hash, err := hashSecondaryPassword(password)
	if err != nil {
		return 0, 11606, err
	}
	model := orm.SecondaryPasswordSettings{
		CommanderID:  client.Commander.CommanderID,
		PasswordHash: hash,
		Notice:       payload.GetNotice(),
		SystemList:   orm.ToInt64List(settings),
		FailCount:    0,
		FailCd:       nil,
	}
	if err := orm.UpsertSecondaryPasswordSettings(orm.GormDB, model); err != nil {
		return 0, 11606, err
	}
	response := protobuf.SC_11606{Result: proto.Uint32(0)}
	return client.SendMessage(11606, &response)
}
