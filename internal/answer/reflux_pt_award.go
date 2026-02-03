package answer

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func RefluxGetPTAward(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11755
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11756, err
	}
	now := uint32(time.Now().Unix())
	ptTemplates, _, ptItemID, err := loadReturnPtTemplates()
	if err != nil {
		return 0, 11756, err
	}
	_, signIDs, err := loadReturnSignTemplates()
	if err != nil {
		return 0, 11756, err
	}
	state, err := orm.GetOrCreateRefluxState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11756, err
	}
	response := protobuf.SC_11756{Result: proto.Uint32(1), AwardList: []*protobuf.DROPINFO{}}
	if state.Active != 1 {
		return client.SendMessage(11756, &response)
	}
	if isRefluxExpired(state.ReturnTime, uint32(len(signIDs)), now) {
		state.Active = 0
		if err := orm.SaveRefluxState(orm.GormDB, state); err != nil {
			return 0, 11756, err
		}
		return client.SendMessage(11756, &response)
	}
	if err := ensureCommanderLoaded(client, "Reflux"); err != nil {
		return 0, 11756, err
	}
	if ptItemID != 0 {
		ptCount, _, err := getCommanderItemCountFromDB(client, ptItemID)
		if err != nil {
			return 0, 11756, err
		}
		state.Pt = ptCount
	}
	nextStage := state.PtStage + 1
	config, ok := ptTemplates[nextStage]
	if !ok {
		return client.SendMessage(11756, &response)
	}
	if state.Pt < config.PtRequire {
		return client.SendMessage(11756, &response)
	}
	levelIndex, err := selectLevelIndex(state.ReturnLv, config.Level)
	if err != nil {
		return 0, 11756, err
	}
	drops, err := buildAwardDrops([][]uint32{config.AwardDisplay[levelIndex]})
	if err != nil {
		return 0, 11756, err
	}
	dropMap := make(map[string]*protobuf.DROPINFO, len(drops))
	for _, drop := range drops {
		key := fmt.Sprintf("%d_%d", drop.GetType(), drop.GetId())
		dropMap[key] = drop
	}
	if err := applyDropList(client, dropMap); err != nil {
		return 0, 11756, err
	}
	state.PtStage++
	if err := orm.SaveRefluxState(orm.GormDB, state); err != nil {
		return 0, 11756, err
	}
	response.Result = proto.Uint32(0)
	response.AwardList = drops
	return client.SendMessage(11756, &response)
}
