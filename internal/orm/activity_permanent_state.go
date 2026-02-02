package orm

import (
	"errors"

	"gorm.io/gorm"
)

type ActivityPermanentState struct {
	CommanderID         uint32    `gorm:"primary_key"`
	CurrentActivityID   uint32    `gorm:"not_null;default:0"`
	FinishedActivityIDs Int64List `gorm:"type:text;not_null;default:'[]'"`
}

func GetOrCreateActivityPermanentState(db *gorm.DB, commanderID uint32) (*ActivityPermanentState, error) {
	var state ActivityPermanentState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		state = ActivityPermanentState{
			CommanderID:         commanderID,
			CurrentActivityID:   0,
			FinishedActivityIDs: Int64List{},
		}
		if err := db.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	if state.FinishedActivityIDs == nil {
		state.FinishedActivityIDs = Int64List{}
	}
	return &state, nil
}

func SaveActivityPermanentState(db *gorm.DB, state *ActivityPermanentState) error {
	return db.Save(state).Error
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
