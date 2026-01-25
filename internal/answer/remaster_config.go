package answer

import (
	"encoding/json"
	"fmt"

	"github.com/ggmolly/belfast/internal/orm"
)

const remasterConfigCategory = "ShareCfg/re_map_template.json"

type remasterTemplate struct {
	ID       uint32     `json:"id"`
	DropGain [][]uint32 `json:"drop_gain"`
}

type remasterDropGain struct {
	ChapterID uint32
	Pos       uint32
	DropType  uint32
	DropID    uint32
	MaxCount  uint32
}

type remasterDropKey struct {
	ChapterID uint32
	Pos       uint32
}

type gameSetEntry struct {
	KeyValue uint32 `json:"key_value"`
}

func listRemasterDropGains() ([]remasterDropGain, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, remasterConfigCategory)
	if err != nil {
		return nil, err
	}
	drops := make([]remasterDropGain, 0)
	for _, entry := range entries {
		var template remasterTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return nil, err
		}
		for index, gain := range template.DropGain {
			if len(gain) == 0 {
				continue
			}
			if len(gain) < 4 {
				return nil, fmt.Errorf("invalid remaster drop_gain entry for %s", entry.Key)
			}
			drops = append(drops, remasterDropGain{
				ChapterID: gain[0],
				Pos:       uint32(index + 1),
				DropType:  gain[1],
				DropID:    gain[2],
				MaxCount:  gain[3],
			})
		}
	}
	return drops, nil
}

func buildRemasterDropGainMap(entries []remasterDropGain) map[remasterDropKey]remasterDropGain {
	lookup := make(map[remasterDropKey]remasterDropGain, len(entries))
	for _, entry := range entries {
		lookup[remasterDropKey{ChapterID: entry.ChapterID, Pos: entry.Pos}] = entry
	}
	return lookup
}

func loadGamesetValue(key string) (uint32, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/gameset.json", key)
	if err != nil {
		return 0, err
	}
	var data gameSetEntry
	if err := json.Unmarshal(entry.Data, &data); err != nil {
		return 0, err
	}
	return data.KeyValue, nil
}
