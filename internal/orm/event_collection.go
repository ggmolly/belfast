package orm

import (
	"errors"

	"gorm.io/gorm"
)

// EventCollection stores the active collection/commission event state for a commander.
// A row exists while the event is active (or finished-but-unclaimed).
type EventCollection struct {
	CommanderID  uint32    `gorm:"primaryKey"`
	CollectionID uint32    `gorm:"primaryKey"`
	StartTime    uint32    `gorm:"not_null;default:0"`
	FinishTime   uint32    `gorm:"not_null;default:0"`
	ShipIDs      Int64List `gorm:"type:text;not_null;default:'[]'"`
}

func GetOrCreateActiveEvent(db *gorm.DB, commanderID uint32, collectionID uint32) (*EventCollection, error) {
	var event EventCollection
	err := db.Where("commander_id = ? AND collection_id = ?", commanderID, collectionID).First(&event).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		event = EventCollection{
			CommanderID:  commanderID,
			CollectionID: collectionID,
			StartTime:    0,
			FinishTime:   0,
			ShipIDs:      Int64List{},
		}
		if err := db.Create(&event).Error; err != nil {
			return nil, err
		}
	}
	if event.ShipIDs == nil {
		event.ShipIDs = Int64List{}
	}
	return &event, nil
}

func SaveEventCollection(db *gorm.DB, event *EventCollection) error {
	return db.Save(event).Error
}

func GetActiveEventCount(db *gorm.DB, commanderID uint32) (int, error) {
	var count int64
	if err := db.Model(&EventCollection{}).Where("commander_id = ?", commanderID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetEventCollection(db *gorm.DB, commanderID, collectionID uint32) (*EventCollection, error) {
	var event EventCollection
	if err := db.Where("commander_id = ? AND collection_id = ?", commanderID, collectionID).First(&event).Error; err != nil {
		return nil, err
	}
	if event.ShipIDs == nil {
		event.ShipIDs = Int64List{}
	}
	return &event, nil
}

func CancelEventCollection(db *gorm.DB, commanderID, collectionID uint32) error {
	return db.Where("commander_id = ? AND collection_id = ?", commanderID, collectionID).Delete(&EventCollection{}).Error
}

func busyShipIDsFromEvents(events []EventCollection) map[uint32]struct{} {
	busy := make(map[uint32]struct{})
	for _, event := range events {
		for _, shipID := range ToUint32List(event.ShipIDs) {
			busy[shipID] = struct{}{}
		}
	}
	return busy
}

func GetBusyEventShipIDs(db *gorm.DB, commanderID uint32) (map[uint32]struct{}, error) {
	var events []EventCollection
	if err := db.Where("commander_id = ?", commanderID).Find(&events).Error; err != nil {
		return nil, err
	}
	return busyShipIDsFromEvents(events), nil
}
