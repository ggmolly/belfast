package orm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
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
		if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled)
VALUES ($1, $2, $3, 0, $4, $5, 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false)`, int64(commanders[i].CommanderID), int64(commanders[i].AccountID), commanders[i].Level, commanders[i].Name, commanders[i].LastLogin); err != nil {
			t.Fatalf("seed commander: %v", err)
		}
	}
	future := time.Now().Add(time.Hour)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO punishments (punished_id, lift_timestamp, is_permanent) VALUES ($1, $2, $3)`, int64(100), future, false); err != nil {
		t.Fatalf("seed punishment: %v", err)
	}

	list, err := ListCommanders(PlayerQueryParams{Offset: 0, Limit: 10, MinLevel: 5})
	if err != nil {
		t.Fatalf("list commanders: %v", err)
	}
	if list.Total != 1 {
		t.Fatalf("expected 1 commander, got %d", list.Total)
	}
	search, err := SearchCommanders(PlayerQueryParams{Offset: 0, Limit: 10, Search: "alpha"})
	if err != nil {
		t.Fatalf("search commanders: %v", err)
	}
	if search.Total != 1 {
		t.Fatalf("expected 1 search result")
	}
	filtered, err := ListCommanders(PlayerQueryParams{FilterBanned: true})
	if err != nil {
		t.Fatalf("filter banned: %v", err)
	}
	if filtered.Total != 1 {
		t.Fatalf("expected 1 banned commander")
	}
	online, err := ListCommanders(PlayerQueryParams{FilterOnline: true, OnlineIDs: []uint32{100}})
	if err != nil {
		t.Fatalf("filter online: %v", err)
	}
	if online.Total != 1 {
		t.Fatalf("expected 1 online commander")
	}
	noneOnline, err := ListCommanders(PlayerQueryParams{FilterOnline: true, OnlineIDs: nil})
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
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled)
VALUES ($1, $2, $3, 0, $4, now(), 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Level, commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	item := Item{ID: 400, Name: "Item", Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6)`, int64(item.ID), item.Name, item.Rarity, item.ShopID, item.Type, item.VirtualType); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(item.ID), int64(1)); err != nil {
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
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO punishments (punished_id, lift_timestamp, is_permanent) VALUES ($1, $2, $3)`, int64(120), future, false); err != nil {
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
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO punishments (punished_id, lift_timestamp, is_permanent) VALUES ($1, $2, $3)`, int64(121), past, false); err != nil {
		t.Fatalf("seed punishment past: %v", err)
	}
	status, err = GetBanStatus(121)
	if err != nil || status.Banned {
		t.Fatalf("expected not banned status")
	}
	if _, err := ActivePunishment(121); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected no active punishment")
	}

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO punishments (punished_id, lift_timestamp, is_permanent) VALUES ($1, $2, $3)`, int64(122), nil, true); err != nil {
		t.Fatalf("seed permanent punishment: %v", err)
	}
	status, err = GetBanStatus(122)
	if err != nil || !status.Banned {
		t.Fatalf("expected permanent ban")
	}
}
