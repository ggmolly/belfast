package orm

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ActivityFleet struct {
	CommanderID uint32                 `gorm:"primaryKey"`
	ActivityID  uint32                 `gorm:"primaryKey"`
	GroupList   ActivityFleetGroupList `gorm:"type:text;not_null;default:'[]'"`
}

type ActivityFleetCommander struct {
	Pos uint32 `json:"pos"`
	ID  uint32 `json:"id"`
}

type ActivityFleetGroup struct {
	ID         uint32                   `json:"id"`
	ShipList   []uint32                 `json:"ship_list"`
	Commanders []ActivityFleetCommander `json:"commanders"`
}

type ActivityFleetGroupList []ActivityFleetGroup

func (list ActivityFleetGroupList) Value() (driver.Value, error) {
	payload, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return string(payload), nil
}

func (list *ActivityFleetGroupList) Scan(value any) error {
	if value == nil {
		*list = nil
		return nil
	}
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), list)
	case []byte:
		return json.Unmarshal(v, list)
	default:
		return fmt.Errorf("unsupported ActivityFleetGroupList type: %T", value)
	}
}

func LoadActivityFleetGroups(commanderID uint32, activityID uint32) (ActivityFleetGroupList, bool, error) {
	var entry ActivityFleet
	err := GormDB.Where("commander_id = ? AND activity_id = ?", commanderID, activityID).First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return entry.GroupList, true, nil
}

func SaveActivityFleetGroups(commanderID uint32, activityID uint32, groups ActivityFleetGroupList) error {
	var entry ActivityFleet
	err := GormDB.Where("commander_id = ? AND activity_id = ?", commanderID, activityID).First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			entry = ActivityFleet{
				CommanderID: commanderID,
				ActivityID:  activityID,
				GroupList:   groups,
			}
			return GormDB.Create(&entry).Error
		}
		return err
	}
	entry.GroupList = groups
	return GormDB.Save(&entry).Error
}
