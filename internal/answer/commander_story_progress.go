package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CommanderStoryProgress(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_13001{}
	state, err := orm.GetOrCreateRemasterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 13001, err
	}
	progress, err := orm.ListChapterProgress(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 13001, err
	}
	if orm.ApplyRemasterDailyReset(state, time.Now()) {
		if err := orm.GormDB.Save(state).Error; err != nil {
			return 0, 13001, err
		}
	}
	chapterList := make([]*protobuf.CHAPTERINFO, 0, len(progress))
	for _, entry := range progress {
		killBossCount := entry.KillBossCount
		killEnemyCount := entry.KillEnemyCount
		takeBoxCount := entry.TakeBoxCount
		if entry.PassCount >= 3 {
			template, err := loadChapterTemplate(entry.ChapterID, 0)
			if err != nil {
				return 0, 13001, err
			}
			if template != nil {
				if template.Num1 > 0 && killBossCount < template.Num1 {
					killBossCount = template.Num1
				}
				if template.Num2 > 0 && killEnemyCount < template.Num2 {
					killEnemyCount = template.Num2
				}
				if template.Num3 > 0 && takeBoxCount < template.Num3 {
					takeBoxCount = template.Num3
				}
			}
		}
		chapterList = append(chapterList, &protobuf.CHAPTERINFO{
			Id:               proto.Uint32(entry.ChapterID),
			Progress:         proto.Uint32(entry.Progress),
			KillBossCount:    proto.Uint32(killBossCount),
			KillEnemyCount:   proto.Uint32(killEnemyCount),
			TakeBoxCount:     proto.Uint32(takeBoxCount),
			DefeatCount:      proto.Uint32(entry.DefeatCount),
			TodayDefeatCount: proto.Uint32(entry.TodayDefeatCount),
			PassCount:        proto.Uint32(entry.PassCount),
		})
	}
	response.ChapterList = chapterList
	response.ReactChapter = &protobuf.REACTCHAPTER_INFO{
		Count:           proto.Uint32(state.TicketCount),
		ActiveTimestamp: proto.Uint32(uint32(state.LastDailyResetAt.Unix())),
		ActiveId:        proto.Uint32(state.ActiveChapterID),
		DailyCount:      proto.Uint32(state.DailyCount),
	}
	return client.SendMessage(13001, &response)
}
