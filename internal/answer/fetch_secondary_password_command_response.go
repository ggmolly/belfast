package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FetchSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	if client.Commander == nil {
		return 0, 11604, errors.New("missing commander")
	}
	settings, err := orm.GetSecondaryPasswordSettings(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11604, err
	}
	now := time.Now().Unix()
	if settings.FailCd != nil && now >= *settings.FailCd {
		settings.FailCd = nil
		settings.FailCount = 0
		if err := orm.ResetSecondaryPasswordLockout(orm.GormDB, client.Commander.CommanderID); err != nil {
			return 0, 11604, err
		}
	}
	state := uint32(0)
	if settings.PasswordHash != "" && len(settings.SystemList) > 0 {
		state = 1
	}
	response := protobuf.SC_11604{
		State:      proto.Uint32(state),
		SystemList: orm.ToUint32List(settings.SystemList),
		FailCount:  proto.Uint32(settings.FailCount),
		Notice:     proto.String(settings.Notice),
	}
	if settings.FailCd != nil {
		response.FailCd = proto.Uint32(uint32(*settings.FailCd))
	}
	return client.SendMessage(11604, &response)
}
