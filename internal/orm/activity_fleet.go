package orm

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
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
	if db.DefaultStore == nil {
		return nil, false, errors.New("db not initialized")
	}
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `SELECT group_list FROM activity_fleets WHERE commander_id = $1 AND activity_id = $2`, int64(commanderID), int64(activityID))
	var groups ActivityFleetGroupList
	if err := row.Scan(&groups); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return groups, true, nil
}

func SaveActivityFleetGroups(commanderID uint32, activityID uint32, groups ActivityFleetGroupList) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	payload, err := json.Marshal(groups)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO activity_fleets (
  commander_id,
  activity_id,
  group_list
) VALUES (
  $1, $2, $3
)
ON CONFLICT (commander_id, activity_id)
DO UPDATE SET group_list = EXCLUDED.group_list
`, int64(commanderID), int64(activityID), payload)
	return err
}
