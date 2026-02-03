package answer

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func RefluxSign(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11753
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11754, err
	}
	now := uint32(time.Now().Unix())
	signTemplates, signIDs, err := loadReturnSignTemplates()
	if err != nil {
		return 0, 11754, err
	}
	state, err := orm.GetOrCreateRefluxState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11754, err
	}
	response := protobuf.SC_11754{Result: proto.Uint32(1), AwardList: []*protobuf.DROPINFO{}}
	if state.Active != 1 {
		return client.SendMessage(11754, &response)
	}
	if isRefluxExpired(state.ReturnTime, uint32(len(signIDs)), now) {
		state.Active = 0
		if err := orm.SaveRefluxState(orm.GormDB, state); err != nil {
			return 0, 11754, err
		}
		return client.SendMessage(11754, &response)
	}
	if state.SignCnt >= uint32(len(signIDs)) {
		return client.SendMessage(11754, &response)
	}
	if isSameDay(state.SignLastTime, now) {
		return client.SendMessage(11754, &response)
	}
	if err := ensureCommanderLoaded(client, "Reflux"); err != nil {
		return 0, 11754, err
	}
	nextID := state.SignCnt + 1
	config, ok := signTemplates[nextID]
	if !ok {
		return client.SendMessage(11754, &response)
	}
	levelIndex, err := selectLevelIndex(state.ReturnLv, config.Level)
	if err != nil {
		return 0, 11754, err
	}
	display := config.AwardDisplay[levelIndex]
	drops, err := buildAwardDrops(display)
	if err != nil {
		return 0, 11754, err
	}
	dropMap := make(map[string]*protobuf.DROPINFO, len(drops))
	for _, drop := range drops {
		key := fmt.Sprintf("%d_%d", drop.GetType(), drop.GetId())
		dropMap[key] = drop
	}
	if err := applyDropList(client, dropMap); err != nil {
		return 0, 11754, err
	}
	state.SignCnt++
	state.SignLastTime = now
	if err := orm.SaveRefluxState(orm.GormDB, state); err != nil {
		return 0, 11754, err
	}
	response.Result = proto.Uint32(0)
	response.AwardList = drops
	return client.SendMessage(11754, &response)
}
