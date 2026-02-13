package orm

import (
	"context"
	"fmt"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type CommanderTB struct {
	CommanderID uint32 `gorm:"primary_key"`
	State       []byte `gorm:"not_null"`
	Permanent   []byte `gorm:"not_null"`
}

func GetCommanderTB(commanderID uint32) (*CommanderTB, error) {
	ctx := context.Background()
	entry := CommanderTB{}
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, state, permanent
FROM commander_tbs
WHERE commander_id = $1
`, int64(commanderID)).Scan(&entry.CommanderID, &entry.State, &entry.Permanent)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func NewCommanderTB(commanderID uint32, info *protobuf.TBINFO, permanent *protobuf.TBPERMANENT) (*CommanderTB, error) {
	stateBytes, err := proto.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tb state: %w", err)
	}
	permanentBytes, err := proto.Marshal(permanent)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tb permanent state: %w", err)
	}
	return &CommanderTB{
		CommanderID: commanderID,
		State:       stateBytes,
		Permanent:   permanentBytes,
	}, nil
}

func (entry *CommanderTB) Decode() (*protobuf.TBINFO, *protobuf.TBPERMANENT, error) {
	state := &protobuf.TBINFO{}
	if err := proto.Unmarshal(entry.State, state); err != nil {
		return nil, nil, fmt.Errorf("failed to decode tb state: %w", err)
	}
	permanent := &protobuf.TBPERMANENT{}
	if err := proto.Unmarshal(entry.Permanent, permanent); err != nil {
		return nil, nil, fmt.Errorf("failed to decode tb permanent state: %w", err)
	}
	return state, permanent, nil
}

func (entry *CommanderTB) Encode(info *protobuf.TBINFO, permanent *protobuf.TBPERMANENT) error {
	stateBytes, err := proto.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to encode tb state: %w", err)
	}
	permanentBytes, err := proto.Marshal(permanent)
	if err != nil {
		return fmt.Errorf("failed to encode tb permanent state: %w", err)
	}
	entry.State = stateBytes
	entry.Permanent = permanentBytes
	return nil
}

func SaveCommanderTB(entry *CommanderTB, info *protobuf.TBINFO, permanent *protobuf.TBPERMANENT) error {
	if err := entry.Encode(info, permanent); err != nil {
		return err
	}
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO commander_tbs (commander_id, state, permanent)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id)
DO UPDATE SET
  state = EXCLUDED.state,
  permanent = EXCLUDED.permanent
`, int64(entry.CommanderID), entry.State, entry.Permanent)
	return err
}

func DeleteCommanderTB(commanderID uint32) (bool, error) {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM commander_tbs
WHERE commander_id = $1
`, int64(commanderID))
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
