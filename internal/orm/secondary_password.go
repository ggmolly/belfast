package orm

import (
	"context"
	"encoding/json"

	"github.com/ggmolly/belfast/internal/db"
)

type SecondaryPasswordState struct {
	CommanderID  uint32    `gorm:"primaryKey;autoIncrement:false"`
	PasswordHash string    `gorm:"type:text;not_null;default:''"`
	Notice       string    `gorm:"type:text;not_null;default:''"`
	SystemList   Int64List `gorm:"type:text;not_null;default:'[]'"`
	State        uint32    `gorm:"not_null;default:0"`
	FailCount    uint32    `gorm:"not_null;default:0"`
	FailCd       uint32    `gorm:"not_null;default:0"`
}

func GetOrCreateSecondaryPasswordState(commanderID uint32) (*SecondaryPasswordState, error) {
	ctx := context.Background()
	state, err := getSecondaryPasswordState(ctx, commanderID)
	err = db.MapNotFound(err)
	if err == nil {
		return state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}

	row := db.DefaultStore.Pool.QueryRow(ctx, `
INSERT INTO secondary_password_states (
  commander_id,
  password_hash,
  notice,
  system_list,
  state,
  fail_count,
  fail_cd
) VALUES (
  $1, '', '', '[]', 0, 0, 0
)
ON CONFLICT (commander_id)
DO NOTHING
RETURNING commander_id, password_hash, notice, system_list, state, fail_count, fail_cd
`, int64(commanderID))
	created, scanErr := secondaryPasswordStateFromRow(row)
	scanErr = db.MapNotFound(scanErr)
	if scanErr == nil {
		return created, nil
	}
	if !db.IsNotFound(scanErr) {
		return nil, scanErr
	}
	return getSecondaryPasswordState(ctx, commanderID)
}

func SaveSecondaryPasswordState(state *SecondaryPasswordState) error {
	ctx := context.Background()
	systemListRaw, err := json.Marshal(state.SystemList)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO secondary_password_states (
  commander_id,
  password_hash,
  notice,
  system_list,
  state,
  fail_count,
  fail_cd
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (commander_id)
DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  notice = EXCLUDED.notice,
  system_list = EXCLUDED.system_list,
  state = EXCLUDED.state,
  fail_count = EXCLUDED.fail_count,
  fail_cd = EXCLUDED.fail_cd
`, int64(state.CommanderID), state.PasswordHash, state.Notice, string(systemListRaw), int64(state.State), int64(state.FailCount), int64(state.FailCd))
	return err
}

func getSecondaryPasswordState(ctx context.Context, commanderID uint32) (*SecondaryPasswordState, error) {
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, password_hash, notice, system_list, state, fail_count, fail_cd
FROM secondary_password_states
WHERE commander_id = $1
`, int64(commanderID))
	state, err := secondaryPasswordStateFromRow(row)
	if err != nil {
		return nil, err
	}
	return state, nil
}

type secondaryPasswordScanner interface {
	Scan(dest ...any) error
}

func secondaryPasswordStateFromRow(row secondaryPasswordScanner) (*SecondaryPasswordState, error) {
	var commanderID int64
	var stateCode int64
	var failCount int64
	var failCd int64
	var systemListRaw string
	state := &SecondaryPasswordState{}
	if err := row.Scan(&commanderID, &state.PasswordHash, &state.Notice, &systemListRaw, &stateCode, &failCount, &failCd); err != nil {
		return nil, err
	}
	state.CommanderID = uint32(commanderID)
	state.State = uint32(stateCode)
	state.FailCount = uint32(failCount)
	state.FailCd = uint32(failCd)
	if systemListRaw == "" {
		state.SystemList = Int64List{}
		return state, nil
	}
	if err := json.Unmarshal([]byte(systemListRaw), &state.SystemList); err != nil {
		return nil, err
	}
	if state.SystemList == nil {
		state.SystemList = Int64List{}
	}
	return state, nil
}
