package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/expedition"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func GetSubmarineExpeditionInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13401
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13402, err
	}

	now := time.Now()
	weekStart := weekStartMondayUTC(now)
	weekStartUnix := uint32(weekStart.Unix())

	state, err := orm.GetSubmarineState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 13402, err
		}
		state = &orm.SubmarineExpeditionState{CommanderID: client.Commander.CommanderID, LastRefreshTime: weekStartUnix}
		if err := orm.UpsertSubmarineState(orm.GormDB, state); err != nil {
			return 0, 13402, err
		}
	}

	if state.LastRefreshTime < weekStartUnix {
		if err := orm.ResetWeeklyRefresh(orm.GormDB, client.Commander.CommanderID, weekStartUnix); err != nil {
			return 0, 13402, err
		}
		state.WeeklyRefreshCount = 0
		state.LastRefreshTime = weekStartUnix
	}

	chapters, err := expedition.LoadSubmarineChapters()
	if err != nil {
		logger.WithFields("SubmarineExpedition", logger.PacketFields(13401, "in")...).Warn("failed to load submarine expedition config")
		chapters = nil
	}
	chapterList := make([]*protobuf.PRO_CHAPTER_SUBMARINE, 0)
	commanderLevel := client.Commander.Level
	for _, chapter := range chapters {
		if commanderLevel < int(chapter.MinLevel) {
			continue
		}
		chapterList = append(chapterList, &protobuf.PRO_CHAPTER_SUBMARINE{
			ChapterId:  proto.Uint32(chapter.ChapterID),
			ActiveTime: proto.Uint32(0),
			Index:      proto.Uint32(chapter.Index),
		})
	}

	refreshCount := remainingWeeklyRefreshes(state.WeeklyRefreshCount)
	nextRefresh := uint32(nextWeeklyResetUTC(weekStart).Unix())
	response := protobuf.SC_13402{
		NextRefreshTime: proto.Uint32(nextRefresh),
		RefreshCount:    proto.Uint32(refreshCount),
		ChapterList:     chapterList,
		Progress:        proto.Uint32(state.OverallProgress),
	}
	_ = payload // type filter currently ignored
	return client.SendMessage(13402, &response)
}

func weekStartMondayUTC(now time.Time) time.Time {
	utc := now.UTC()
	year, month, day := utc.Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	daysSinceMonday := (int(utc.Weekday()) - int(time.Monday) + 7) % 7
	return start.AddDate(0, 0, -daysSinceMonday)
}

func nextWeeklyResetUTC(weekStart time.Time) time.Time {
	return weekStart.Add(7 * 24 * time.Hour)
}

func remainingWeeklyRefreshes(used uint32) uint32 {
	if used >= 4 {
		return 0
	}
	return 4 - used
}
