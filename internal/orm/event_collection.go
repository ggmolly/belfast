package orm

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
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

func GetOrCreateActiveEvent(_ any, commanderID uint32, collectionID uint32) (*EventCollection, error) {
	ctx := context.Background()
	event, err := GetEventCollection(nil, commanderID, collectionID)
	if err == nil {
		return event, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}
	shipIDsJSON, err := json.Marshal(Int64List{})
	if err != nil {
		return nil, err
	}
	if err := db.DefaultStore.Queries.CreateEventCollection(ctx, gen.CreateEventCollectionParams{
		CommanderID:  int64(commanderID),
		CollectionID: int64(collectionID),
		StartTime:    0,
		FinishTime:   0,
		ShipIds:      shipIDsJSON,
	}); err != nil {
		return nil, err
	}
	return GetEventCollection(nil, commanderID, collectionID)
}

func SaveEventCollection(_ any, event *EventCollection) error {
	ctx := context.Background()
	shipIDsJSON, err := json.Marshal(event.ShipIDs)
	if err != nil {
		return err
	}
	return db.DefaultStore.Queries.UpdateEventCollection(ctx, gen.UpdateEventCollectionParams{
		CommanderID:  int64(event.CommanderID),
		CollectionID: int64(event.CollectionID),
		StartTime:    int64(event.StartTime),
		FinishTime:   int64(event.FinishTime),
		ShipIds:      shipIDsJSON,
	})
}

func GetActiveEventCount(_ any, commanderID uint32) (int, error) {
	ctx := context.Background()
	count, err := db.DefaultStore.Queries.CountActiveEventCollections(ctx, int64(commanderID))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetEventCollection(_ any, commanderID, collectionID uint32) (*EventCollection, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetEventCollection(ctx, gen.GetEventCollectionParams{CommanderID: int64(commanderID), CollectionID: int64(collectionID)})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	var shipIDs Int64List
	if err := json.Unmarshal(row.ShipIds, &shipIDs); err != nil {
		return nil, err
	}
	if shipIDs == nil {
		shipIDs = Int64List{}
	}
	event := EventCollection{
		CommanderID:  uint32(row.CommanderID),
		CollectionID: uint32(row.CollectionID),
		StartTime:    uint32(row.StartTime),
		FinishTime:   uint32(row.FinishTime),
		ShipIDs:      shipIDs,
	}
	return &event, nil
}

func CancelEventCollection(_ any, commanderID, collectionID uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.DeleteEventCollection(ctx, gen.DeleteEventCollectionParams{CommanderID: int64(commanderID), CollectionID: int64(collectionID)})
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

func GetBusyEventShipIDs(_ any, commanderID uint32) (map[uint32]struct{}, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListBusyEventCollections(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	events := make([]EventCollection, 0, len(rows))
	for _, r := range rows {
		var shipIDs Int64List
		if err := json.Unmarshal(r.ShipIds, &shipIDs); err != nil {
			return nil, err
		}
		events = append(events, EventCollection{
			CommanderID:  uint32(r.CommanderID),
			CollectionID: uint32(r.CollectionID),
			StartTime:    uint32(r.StartTime),
			FinishTime:   uint32(r.FinishTime),
			ShipIDs:      shipIDs,
		})
	}
	return busyShipIDsFromEvents(events), nil
}

func ListActiveEventCollections(commanderID uint32) ([]EventCollection, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListBusyEventCollections(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	events := make([]EventCollection, 0, len(rows))
	for _, r := range rows {
		var shipIDs Int64List
		if err := json.Unmarshal(r.ShipIds, &shipIDs); err != nil {
			return nil, err
		}
		events = append(events, EventCollection{
			CommanderID:  uint32(r.CommanderID),
			CollectionID: uint32(r.CollectionID),
			StartTime:    uint32(r.StartTime),
			FinishTime:   uint32(r.FinishTime),
			ShipIDs:      shipIDs,
		})
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].CollectionID < events[j].CollectionID
	})
	return events, nil
}
