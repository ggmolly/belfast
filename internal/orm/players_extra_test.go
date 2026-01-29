package orm

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestPlayerQueriesAndFilters(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Punishment{})
	clearTable(t, &Commander{})

	commanders := []Commander{
		{CommanderID: 100, AccountID: 100, Name: "Alpha", Level: 10, LastLogin: time.Now()},
		{CommanderID: 101, AccountID: 101, Name: "Beta", Level: 2, LastLogin: time.Now().Add(-time.Hour)},
	}
	for i := range commanders {
		if err := GormDB.Create(&commanders[i]).Error; err != nil {
			t.Fatalf("seed commander: %v", err)
		}
	}
	future := time.Now().Add(time.Hour)
	if err := GormDB.Create(&Punishment{PunishedID: 100, LiftTimestamp: &future}).Error; err != nil {
		t.Fatalf("seed punishment: %v", err)
	}

	list, err := ListCommanders(GormDB, PlayerQueryParams{Offset: 0, Limit: 10, MinLevel: 5})
	if err != nil {
		t.Fatalf("list commanders: %v", err)
	}
	if list.Total != 1 {
		t.Fatalf("expected 1 commander, got %d", list.Total)
	}
	search, err := SearchCommanders(GormDB, PlayerQueryParams{Offset: 0, Limit: 10, Search: "alpha"})
	if err != nil {
		t.Fatalf("search commanders: %v", err)
	}
	if search.Total != 1 {
		t.Fatalf("expected 1 search result")
	}
	filtered, err := ListCommanders(GormDB, PlayerQueryParams{FilterBanned: true})
	if err != nil {
		t.Fatalf("filter banned: %v", err)
	}
	if filtered.Total != 1 {
		t.Fatalf("expected 1 banned commander")
	}
	online, err := ListCommanders(GormDB, PlayerQueryParams{FilterOnline: true, OnlineIDs: []uint32{100}})
	if err != nil {
		t.Fatalf("filter online: %v", err)
	}
	if online.Total != 1 {
		t.Fatalf("expected 1 online commander")
	}
	noneOnline, err := ListCommanders(GormDB, PlayerQueryParams{FilterOnline: true, OnlineIDs: nil})
	if err != nil {
		t.Fatalf("filter online empty: %v", err)
	}
	if noneOnline.Total != 0 {
		t.Fatalf("expected 0 online commanders")
	}
}

func TestLoadCommanderWithDetails(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Commander{})
	clearTable(t, &Item{})
	clearTable(t, &CommanderItem{})

	commander := Commander{CommanderID: 110, AccountID: 110, Name: "Details", Level: 1}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	item := Item{ID: 400, Name: "Item", Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	if err := GormDB.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if err := GormDB.Create(&CommanderItem{CommanderID: commander.CommanderID, ItemID: item.ID, Count: 1}).Error; err != nil {
		t.Fatalf("seed commander item: %v", err)
	}
	loaded, err := LoadCommanderWithDetails(commander.CommanderID)
	if err != nil {
		t.Fatalf("load commander details: %v", err)
	}
	if loaded.CommanderID != commander.CommanderID {
		t.Fatalf("unexpected commander loaded")
	}
}

func TestBanStatusAndActivePunishment(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Punishment{})

	status, err := GetBanStatus(999)
	if err != nil {
		t.Fatalf("ban status: %v", err)
	}
	if status.Banned {
		t.Fatalf("expected not banned")
	}

	future := time.Now().Add(time.Hour)
	if err := GormDB.Create(&Punishment{PunishedID: 120, LiftTimestamp: &future}).Error; err != nil {
		t.Fatalf("seed punishment: %v", err)
	}
	status, err = GetBanStatus(120)
	if err != nil || !status.Banned {
		t.Fatalf("expected banned status")
	}
	if _, err := ActivePunishment(120); err != nil {
		t.Fatalf("expected active punishment")
	}

	past := time.Now().Add(-time.Hour)
	if err := GormDB.Create(&Punishment{PunishedID: 121, LiftTimestamp: &past}).Error; err != nil {
		t.Fatalf("seed punishment past: %v", err)
	}
	status, err = GetBanStatus(121)
	if err != nil || status.Banned {
		t.Fatalf("expected not banned status")
	}
	if _, err := ActivePunishment(121); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no active punishment")
	}

	if err := GormDB.Create(&Punishment{PunishedID: 122, LiftTimestamp: nil, IsPermanent: true}).Error; err != nil {
		t.Fatalf("seed permanent punishment: %v", err)
	}
	status, err = GetBanStatus(122)
	if err != nil || !status.Banned {
		t.Fatalf("expected permanent ban")
	}
}
