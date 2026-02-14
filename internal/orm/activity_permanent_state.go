package orm

import (
	"context"
	"encoding/json"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type ActivityPermanentState struct {
	CommanderID         uint32    `gorm:"primary_key"`
	CurrentActivityID   uint32    `gorm:"not_null;default:0"`
	FinishedActivityIDs Int64List `gorm:"type:text;not_null;default:'[]'"`
}

func GetOrCreateActivityPermanentState(commanderID uint32) (*ActivityPermanentState, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetActivityPermanentState(ctx, int64(commanderID))
	err = db.MapNotFound(err)
	if err != nil {
		if !db.IsNotFound(err) {
			return nil, err
		}
		created, createErr := db.DefaultStore.Queries.CreateActivityPermanentState(ctx, int64(commanderID))
		if createErr != nil {
			return nil, createErr
		}
		state, parseErr := activityPermanentStateFromRow(created.CommanderID, created.CurrentActivityID, created.FinishedActivityIds)
		if parseErr != nil {
			return nil, parseErr
		}
		return state, nil
	}
	state, err := activityPermanentStateFromRow(row.CommanderID, row.CurrentActivityID, row.FinishedActivityIds)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func SaveActivityPermanentState(state *ActivityPermanentState) error {
	ctx := context.Background()
	finishedRaw, err := json.Marshal([]int64(state.FinishedActivityIDs))
	if err != nil {
		return err
	}
	return db.DefaultStore.Queries.UpsertActivityPermanentState(ctx, gen.UpsertActivityPermanentStateParams{
		CommanderID:         int64(state.CommanderID),
		CurrentActivityID:   int64(state.CurrentActivityID),
		FinishedActivityIds: finishedRaw,
	})
}

func activityPermanentStateFromRow(commanderID int64, currentActivityID int64, finishedActivityIDs []byte) (*ActivityPermanentState, error) {
	finished := Int64List{}
	if len(finishedActivityIDs) > 0 {
		if err := json.Unmarshal(finishedActivityIDs, &finished); err != nil {
			return nil, err
		}
	}
	state := &ActivityPermanentState{
		CommanderID:         uint32(commanderID),
		CurrentActivityID:   uint32(currentActivityID),
		FinishedActivityIDs: finished,
	}
	return state, nil
}

func (state *ActivityPermanentState) FinishedList() []uint32 {
	return ToUint32List(state.FinishedActivityIDs)
}

func (state *ActivityPermanentState) HasFinished(activityID uint32) bool {
	for _, id := range state.FinishedList() {
		if id == activityID {
			return true
		}
	}
	return false
}

func (state *ActivityPermanentState) AddFinished(activityID uint32) {
	if state.HasFinished(activityID) {
		return
	}
	ids := append(state.FinishedList(), activityID)
	state.FinishedActivityIDs = ToInt64List(ids)
}
