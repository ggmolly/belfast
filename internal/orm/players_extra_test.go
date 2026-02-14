package orm

import (
	"context"
	"encoding/json"
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

func TestLoadCommanderWithDetailsHydratesCompensations(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CompensationAttachment{})
	clearTable(t, &Compensation{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 111, AccountID: 111, Name: "CompLoad", Level: 1}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled)
VALUES ($1, $2, $3, 0, $4, now(), 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Level, commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	comp := Compensation{CommanderID: commander.CommanderID, Title: "T", Text: "Body", ExpiresAt: time.Now().Add(time.Hour)}
	comp.Attachments = []CompensationAttachment{{Type: 1, ItemID: 1, Quantity: 1}}
	if err := comp.Create(); err != nil {
		t.Fatalf("seed compensation: %v", err)
	}

	loaded, err := LoadCommanderWithDetails(commander.CommanderID)
	if err != nil {
		t.Fatalf("load commander details: %v", err)
	}
	if len(loaded.Compensations) != 1 {
		t.Fatalf("expected 1 compensation, got %d", len(loaded.Compensations))
	}
	if len(loaded.Compensations[0].Attachments) != 1 {
		t.Fatalf("expected 1 compensation attachment, got %d", len(loaded.Compensations[0].Attachments))
	}
}

func TestCommanderAddShipUsesConfigDrivenDefaultEquipmentSlots(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedShipEquipment{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &ConfigEntry{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})
	clearTable(t, &Rarity{})

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO rarities (id, name) VALUES ($1, $2)`, int64(2), "Common"); err != nil {
		t.Fatalf("seed rarity: %v", err)
	}

	ship := Ship{TemplateID: 9101, Name: "SlotTest", EnglishName: "SlotTest", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := ship.Create(); err != nil {
		t.Fatalf("seed ship: %v", err)
	}

	if err := UpsertConfigEntry(shipDataTemplateCategory, "9101", json.RawMessage(`{"id":9101,"equip_1":[1],"equip_2":[2],"equip_3":[3],"equip_4":[4],"equip_5":[],"equip_id_1":2201,"equip_id_2":2202,"equip_id_3":2203}`)); err != nil {
		t.Fatalf("seed ship equip config: %v", err)
	}

	commander := Commander{CommanderID: 112, AccountID: 112, Name: "ShipInit"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	owned, err := commander.AddShip(ship.TemplateID)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}

	equipments, err := ListOwnedShipEquipment(commander.CommanderID, owned.ID)
	if err != nil {
		t.Fatalf("list ship equipment: %v", err)
	}
	if len(equipments) != 4 {
		t.Fatalf("expected 4 equipment slots, got %d", len(equipments))
	}
	if equipments[0].EquipID != 2201 || equipments[1].EquipID != 2202 || equipments[2].EquipID != 2203 || equipments[3].EquipID != 0 {
		t.Fatalf("unexpected default equipment ids: %+v", equipments)
	}
}

func TestBuildRetrieveGreedyHydratesShipFields(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Build{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})
	clearTable(t, &Rarity{})

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO rarities (id, name) VALUES ($1, $2)`, int64(2), "Common"); err != nil {
		t.Fatalf("seed rarity: %v", err)
	}
	ship := Ship{TemplateID: 9102, Name: "GreedyShip", EnglishName: "GreedyShip", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := ship.Create(); err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	commander := Commander{CommanderID: 113, AccountID: 113, Name: "GreedyBuilder"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	build := Build{BuilderID: commander.CommanderID, ShipID: ship.TemplateID, PoolID: 3, FinishesAt: time.Now().UTC().Add(time.Hour)}
	if err := build.Create(); err != nil {
		t.Fatalf("create build: %v", err)
	}

	loaded := Build{ID: build.ID}
	if err := loaded.Retrieve(true); err != nil {
		t.Fatalf("retrieve build greedy: %v", err)
	}
	if loaded.Ship.TemplateID != ship.TemplateID || loaded.Ship.Name != ship.Name {
		t.Fatalf("expected hydrated ship, got %+v", loaded.Ship)
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

func TestListCommandersUnlimitedWithNegativeOffset(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Commander{})

	rows := []Commander{
		{CommanderID: 1301, AccountID: 1, Name: "A", Level: 1},
		{CommanderID: 1302, AccountID: 2, Name: "B", Level: 1},
		{CommanderID: 1303, AccountID: 3, Name: "C", Level: 1},
	}
	for _, c := range rows {
		if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled)
VALUES ($1, $2, $3, 0, $4, now(), 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false)`, int64(c.CommanderID), int64(c.AccountID), c.Level, c.Name); err != nil {
			t.Fatalf("seed commander: %v", err)
		}
	}

	commanders, err := ListCommanders(PlayerQueryParams{Offset: -1, Limit: -1})
	if err != nil {
		t.Fatalf("list commanders: %v", err)
	}
	if len(commanders.Commanders) != len(rows) {
		t.Fatalf("expected %d commanders, got %d", len(rows), len(commanders.Commanders))
	}
}

func TestListCommandersRespectsRequestedLimit(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Commander{})

	seedCount := 520
	now := time.Now().UTC()
	for i := 0; i < seedCount; i++ {
		if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled)
	VALUES ($1, $2, $3, 0, $4, $5, 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false)`, int64(5000+i), int64(5000+i), i+1, "Commander-"+time.Duration(i).String(), now.Add(-time.Duration(i)*time.Minute)); err != nil {
			t.Fatalf("seed commander: %v", err)
		}
	}

	result, err := ListCommanders(PlayerQueryParams{Offset: 0, Limit: 600})
	if err != nil {
		t.Fatalf("list commanders: %v", err)
	}
	if result.Total != int64(seedCount) {
		t.Fatalf("expected total %d, got %d", seedCount, result.Total)
	}
	if len(result.Commanders) != seedCount {
		t.Fatalf("expected %d commanders, got %d", seedCount, len(result.Commanders))
	}

	searchResult, err := SearchCommanders(PlayerQueryParams{Offset: 0, Limit: 600, Search: "Commander-"})
	if err != nil {
		t.Fatalf("search commanders: %v", err)
	}
	if searchResult.Total != int64(seedCount) {
		t.Fatalf("expected search total %d, got %d", seedCount, searchResult.Total)
	}
	if len(searchResult.Commanders) != seedCount {
		t.Fatalf("expected %d search commanders, got %d", seedCount, len(searchResult.Commanders))
	}
}

func TestNormalizePlayersPaginationNoORMCap(t *testing.T) {
	offset, limit, unlimited := normalizePlayersPagination(0, 600)
	if offset != 0 || limit != 600 || unlimited {
		t.Fatalf("expected offset=0 limit=600 unlimited=false, got offset=%d limit=%d unlimited=%v", offset, limit, unlimited)
	}

	offset, limit, unlimited = normalizePlayersPagination(-10, -1)
	if offset != 0 || limit != 0 || !unlimited {
		t.Fatalf("expected normalized unlimited pagination, got offset=%d limit=%d unlimited=%v", offset, limit, unlimited)
	}
}
