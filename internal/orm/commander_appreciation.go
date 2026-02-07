package orm

import (
	"errors"

	"gorm.io/gorm"
)

// CommanderAppreciationState stores commander-scoped Appreciation bitsets.
// These bitsets are surfaced in SC_11003 so the client can rebuild local lists.
type CommanderAppreciationState struct {
	CommanderID        uint32    `gorm:"primaryKey;autoIncrement:false"`
	MusicNo            uint32    `gorm:"not_null;default:0"`
	MusicMode          uint32    `gorm:"not_null;default:0"`
	CartoonReadMark    Int64List `gorm:"type:text;not_null;default:'[]'"`
	CartoonCollectMark Int64List `gorm:"type:text;not_null;default:'[]'"`
	GalleryUnlocks     Int64List `gorm:"type:text;not_null;default:'[]'"`
}

const maxAppreciationMarkBuckets = 4096

func GetOrCreateCommanderAppreciationState(db *gorm.DB, commanderID uint32) (*CommanderAppreciationState, error) {
	var state CommanderAppreciationState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		state = CommanderAppreciationState{
			CommanderID:        commanderID,
			CartoonReadMark:    Int64List{},
			CartoonCollectMark: Int64List{},
			GalleryUnlocks:     Int64List{},
		}
		if err := db.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	if state.CartoonReadMark == nil {
		state.CartoonReadMark = Int64List{}
	}
	if state.CartoonCollectMark == nil {
		state.CartoonCollectMark = Int64List{}
	}
	if state.GalleryUnlocks == nil {
		state.GalleryUnlocks = Int64List{}
	}
	return &state, nil
}

func SaveCommanderAppreciationState(db *gorm.DB, state *CommanderAppreciationState) error {
	return db.Save(state).Error
}

func SetCommanderCartoonReadMark(db *gorm.DB, commanderID uint32, cartoonID uint32) error {
	state, err := GetOrCreateCommanderAppreciationState(db, commanderID)
	if err != nil {
		return err
	}
	marks := ToUint32List(state.CartoonReadMark)
	marks = updateBitsetMark(marks, cartoonID, true)
	state.CartoonReadMark = ToInt64List(marks)
	return SaveCommanderAppreciationState(db, state)
}

func SetCommanderCartoonCollectMark(db *gorm.DB, commanderID uint32, cartoonID uint32, liked bool) error {
	state, err := GetOrCreateCommanderAppreciationState(db, commanderID)
	if err != nil {
		return err
	}
	marks := ToUint32List(state.CartoonCollectMark)
	marks = updateBitsetMark(marks, cartoonID, liked)
	state.CartoonCollectMark = ToInt64List(marks)
	return SaveCommanderAppreciationState(db, state)
}

func SetCommanderAppreciationGalleryUnlock(db *gorm.DB, commanderID uint32, galleryID uint32) error {
	state, err := GetOrCreateCommanderAppreciationState(db, commanderID)
	if err != nil {
		return err
	}
	unlocks := ToUint32List(state.GalleryUnlocks)
	unlocks = updateBitsetMark(unlocks, galleryID, true)
	state.GalleryUnlocks = ToInt64List(unlocks)
	return SaveCommanderAppreciationState(db, state)
}

func updateBitsetMark(marks []uint32, id uint32, enabled bool) []uint32 {
	if id == 0 {
		return marks
	}
	bucket := int((id - 1) / 32)
	bit := uint((id - 1) % 32)
	if bucket >= maxAppreciationMarkBuckets {
		return marks
	}
	if !enabled && len(marks) < bucket+1 {
		return marks
	}
	if len(marks) < bucket+1 {
		extended := make([]uint32, bucket+1)
		copy(extended, marks)
		marks = extended
	}
	if enabled {
		marks[bucket] |= (1 << bit)
		return marks
	}
	marks[bucket] &^= (1 << bit)
	return marks
}
