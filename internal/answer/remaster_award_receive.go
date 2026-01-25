package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func RemasterAwardReceive(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13507
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13508, err
	}
	gains, err := listRemasterDropGains()
	if err != nil {
		return 0, 13508, err
	}
	lookup := buildRemasterDropGainMap(gains)
	key := remasterDropKey{ChapterID: payload.GetChapterId(), Pos: payload.GetPos()}
	config, ok := lookup[key]
	response := protobuf.SC_13508{Result: proto.Uint32(1), DropList: []*protobuf.DROPINFO{}}
	if !ok {
		return client.SendMessage(13508, &response)
	}
	progress, err := getRemasterProgress(orm.GormDB, client.Commander.CommanderID, config.ChapterID, config.Pos)
	if err != nil {
		return 0, 13508, err
	}
	if progress.Received || progress.Count < config.MaxCount {
		return client.SendMessage(13508, &response)
	}
	if client.Commander.CommanderItemsMap == nil && client.Commander.MiscItemsMap == nil {
		logger.LogEvent("Remaster", "Award", "commander maps missing, reloading commander", logger.LOG_LEVEL_INFO)
		if err := client.Commander.Load(); err != nil {
			logger.LogEvent("Remaster", "Award", "commander load failed", logger.LOG_LEVEL_ERROR)
			return 0, 13508, err
		}
	}
	drop := newDropInfo(config.DropType, config.DropID, 1)
	drops := map[string]*protobuf.DROPINFO{fmt.Sprintf("%d_%d", config.DropType, config.DropID): drop}
	if err := applyDropList(client, drops); err != nil {
		return 0, 13508, err
	}
	progress.Received = true
	if err := orm.UpsertRemasterProgress(orm.GormDB, progress); err != nil {
		return 0, 13508, err
	}
	response.Result = proto.Uint32(0)
	response.DropList = []*protobuf.DROPINFO{drop}
	return client.SendMessage(13508, &response)
}
