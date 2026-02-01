package answer

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
)

const surveyActivityType = 101

type surveyActivity struct {
	ActivityID    uint32
	SurveyID      uint32
	RequiredLevel uint32
}

func activeSurveyActivity(commanderLevel uint32, surveyID uint32) (*surveyActivity, error) {
	allowlist, err := loadActivityAllowlist()
	if err != nil {
		return nil, err
	}
	for _, activityID := range allowlist {
		template, err := loadActivityTemplate(activityID)
		if err != nil {
			return nil, err
		}
		if template.Type != surveyActivityType {
			continue
		}
		if len(template.ConfigData) == 0 {
			return nil, errors.New("survey activity missing config_data")
		}
		var config []uint32
		if err := json.Unmarshal(template.ConfigData, &config); err != nil {
			return nil, err
		}
		if len(config) < 2 {
			return nil, errors.New("survey activity config_data requires [open_flag, level]")
		}
		if config[0] != 1 {
			continue
		}
		if commanderLevel < config[1] {
			continue
		}
		if template.ConfigID != surveyID {
			continue
		}
		return &surveyActivity{
			ActivityID:    template.ID,
			SurveyID:      template.ConfigID,
			RequiredLevel: config[1],
		}, nil
	}
	return nil, nil
}

func upsertSurveyState(commanderID uint32, surveyID uint32) error {
	state := orm.SurveyState{
		CommanderID: commanderID,
		SurveyID:    surveyID,
		CompletedAt: time.Now().UTC(),
	}
	return orm.UpsertSurveyState(orm.GormDB, &state)
}
