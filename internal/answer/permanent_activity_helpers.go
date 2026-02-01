package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/orm"
)

const permanentActivityConfigCategory = "ShareCfg/activity_task_permanent.json"

type permanentActivityConfig struct {
	ID uint32 `json:"id"`
}

func loadPermanentActivityIDSet() (map[uint32]struct{}, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, permanentActivityConfigCategory)
	if err != nil {
		return nil, err
	}
	ids := make(map[uint32]struct{}, len(entries))
	for _, entry := range entries {
		var activity permanentActivityConfig
		if err := json.Unmarshal(entry.Data, &activity); err != nil {
			return nil, err
		}
		ids[activity.ID] = struct{}{}
	}
	return ids, nil
}

func filterPermanentActivityIDs(ids []uint32, allowed map[uint32]struct{}) []uint32 {
	if len(ids) == 0 {
		return []uint32{}
	}
	filtered := make([]uint32, 0, len(ids))
	for _, id := range ids {
		if _, ok := allowed[id]; ok {
			filtered = append(filtered, id)
		}
	}
	return filtered
}
