package answer

import (
	"encoding/json"
	"strconv"

	"github.com/ggmolly/belfast/internal/orm"
)

type activityTemplate struct {
	ID           uint32          `json:"id"`
	Type         uint32          `json:"type"`
	ConfigID     uint32          `json:"config_id"`
	Time         json.RawMessage `json:"time"`
	ConfigClient json.RawMessage `json:"config_client"`
	ConfigData   json.RawMessage `json:"config_data"`
}

func loadActivityTemplate(id uint32) (activityTemplate, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/activity_template.json", strconv.FormatUint(uint64(id), 10))
	if err != nil {
		return activityTemplate{}, err
	}
	var template activityTemplate
	if err := json.Unmarshal(entry.Data, &template); err != nil {
		return activityTemplate{}, err
	}
	return template, nil
}
