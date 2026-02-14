package orm

import (
	"errors"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestGetOrCreateActiveEvent(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})

	event, err := GetOrCreateActiveEvent(nil, 100, 200)
	if err != nil {
		t.Fatalf("get or create: %v", err)
	}
	if event.CommanderID != 100 || event.CollectionID != 200 {
		t.Fatalf("unexpected ids")
	}
	if event.StartTime != 0 || event.FinishTime != 0 {
		t.Fatalf("expected default times")
	}

	second, err := GetOrCreateActiveEvent(nil, 100, 200)
	if err != nil {
		t.Fatalf("get existing: %v", err)
	}
	if second.CommanderID != 100 || second.CollectionID != 200 {
		t.Fatalf("unexpected ids")
	}
}

func TestGetActiveEventCount(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})

	if err := SaveEventCollection(nil, &EventCollection{CommanderID: 1, CollectionID: 10, StartTime: 1, FinishTime: 2, ShipIDs: Int64List{1}}); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if err := SaveEventCollection(nil, &EventCollection{CommanderID: 1, CollectionID: 11, StartTime: 1, FinishTime: 2, ShipIDs: Int64List{2}}); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if err := SaveEventCollection(nil, &EventCollection{CommanderID: 2, CollectionID: 12, StartTime: 1, FinishTime: 2, ShipIDs: Int64List{3}}); err != nil {
		t.Fatalf("seed event: %v", err)
	}

	count, err := GetActiveEventCount(nil, 1)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
}

func TestCancelEventCollection(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})

	if err := SaveEventCollection(nil, &EventCollection{CommanderID: 5, CollectionID: 99, StartTime: 1, FinishTime: 2, ShipIDs: Int64List{1}}); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if err := CancelEventCollection(nil, 5, 99); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if _, err := GetEventCollection(nil, 5, 99); err == nil {
		t.Fatalf("expected record not found")
	} else if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveEventCollectionUpdatesExistingRow(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})

	if err := SaveEventCollection(nil, &EventCollection{CommanderID: 9, CollectionID: 77, StartTime: 1, FinishTime: 2, ShipIDs: Int64List{1}}); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if err := SaveEventCollection(nil, &EventCollection{CommanderID: 9, CollectionID: 77, StartTime: 3, FinishTime: 4, ShipIDs: Int64List{2, 3}}); err != nil {
		t.Fatalf("update event: %v", err)
	}

	stored, err := GetEventCollection(nil, 9, 77)
	if err != nil {
		t.Fatalf("load event: %v", err)
	}
	if stored.StartTime != 3 || stored.FinishTime != 4 {
		t.Fatalf("expected updated times, got start=%d finish=%d", stored.StartTime, stored.FinishTime)
	}
	if len(stored.ShipIDs) != 2 || stored.ShipIDs[0] != 2 || stored.ShipIDs[1] != 3 {
		t.Fatalf("expected updated ship ids, got %v", stored.ShipIDs)
	}
}
