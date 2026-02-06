package answer

import (
	"encoding/json"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const escortMapTemplateCategory = "ShareCfg/escort_map_template.json"

type escortMapTemplate struct {
	RefreshTime  uint32          `json:"refresh_time"`
	EscortIDList json.RawMessage `json:"escort_id_list"`
}

func SubmarineChapterInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13403
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13404, err
	}

	state, err := orm.GetOrCreateRemasterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 13404, err
	}
	now := time.Now()
	if orm.ApplyRemasterDailyReset(state, now) {
		if err := orm.GormDB.Save(state).Error; err != nil {
			return 0, 13404, err
		}
	}

	response := protobuf.SC_13404{Result: proto.Uint32(1)}
	if payload.GetType() != 0 {
		return client.SendMessage(13404, &response)
	}

	chapterID, activeAt, index, ok, err := resolveActiveSubmarineChapter(now)
	if err != nil {
		return 0, 13404, err
	}
	if !ok {
		return client.SendMessage(13404, &response)
	}

	response.Result = proto.Uint32(0)
	response.ChapterId = &protobuf.PRO_CHAPTER_SUBMARINE{
		ChapterId:  proto.Uint32(chapterID),
		ActiveTime: proto.Uint32(activeAt),
		Index:      proto.Uint32(index),
	}
	return client.SendMessage(13404, &response)
}

func resolveActiveSubmarineChapter(now time.Time) (chapterID uint32, activeAt uint32, index uint32, ok bool, err error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, escortMapTemplateCategory)
	if err != nil {
		return 0, 0, 0, false, err
	}
	for _, entry := range entries {
		var template escortMapTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return 0, 0, 0, false, err
		}
		escortIDs, ok := parseEscortIDList(template.EscortIDList)
		if !ok || len(escortIDs) == 0 || template.RefreshTime == 0 {
			continue
		}
		refresh := int64(template.RefreshTime)
		nowUnix := now.Unix()
		slotIndex := uint32((nowUnix / refresh) % int64(len(escortIDs)))
		slotStart := nowUnix - (nowUnix % refresh)
		return escortIDs[slotIndex], uint32(slotStart), slotIndex + 1, true, nil
	}
	return 0, 0, 0, false, nil
}

func parseEscortIDList(raw json.RawMessage) ([]uint32, bool) {
	if len(raw) == 0 {
		return nil, false
	}
	var flat []uint32
	if err := json.Unmarshal(raw, &flat); err == nil {
		return flat, true
	}
	var nested [][]uint32
	if err := json.Unmarshal(raw, &nested); err == nil {
		out := make([]uint32, 0)
		for _, group := range nested {
			out = append(out, group...)
		}
		return out, true
	}
	return nil, false
}
