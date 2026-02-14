package orm

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestNormalizePaginationKeepsUnlimitedAndClampsOffset(t *testing.T) {
	tests := []struct {
		offset      int
		limit       int
		expectedOff int
		expectedLim int
		expectedUnl bool
	}{
		{offset: -5, limit: 0, expectedOff: 0, expectedLim: 0, expectedUnl: true},
		{offset: -1, limit: -10, expectedOff: 0, expectedLim: 0, expectedUnl: true},
	}

	for _, tt := range tests {
		off, lim, unlimited := normalizePagination(tt.offset, tt.limit)
		if off != tt.expectedOff || lim != tt.expectedLim || unlimited != tt.expectedUnl {
			t.Fatalf("normalizePagination(%d, %d) = (%d, %d, %t), want (%d, %d, %t)", tt.offset, tt.limit, off, lim, unlimited, tt.expectedOff, tt.expectedLim, tt.expectedUnl)
		}
	}
}

func TestListRaritiesRespectsUnlimitedOffset(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Rarity{})

	for _, r := range []Rarity{{ID: 1, Name: "Common"}, {ID: 2, Name: "Rare"}, {ID: 3, Name: "Elite"}} {
		if err := CreateRarity(&r); err != nil {
			t.Fatalf("create rarity: %v", err)
		}
	}

	rarities, total, err := ListRarities(-4, 0)
	if err != nil {
		t.Fatalf("list rarities: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	if len(rarities) != 3 {
		t.Fatalf("expected 3 rarities, got %d", len(rarities))
	}
	if rarities[0].ID != 1 || rarities[1].ID != 2 || rarities[2].ID != 3 {
		t.Fatalf("expected rarities ordered by id asc, got ids %d,%d,%d", rarities[0].ID, rarities[1].ID, rarities[2].ID)
	}

	page, pageTotal, err := ListRarities(1, 1)
	if err != nil {
		t.Fatalf("list rarities page: %v", err)
	}
	if pageTotal != 3 {
		t.Fatalf("expected paged total 3, got %d", pageTotal)
	}
	if len(page) != 1 {
		t.Fatalf("expected 1 rarity on page, got %d", len(page))
	}
	if page[0].ID != 2 {
		t.Fatalf("expected offset=1 limit=1 to return id 2, got %d", page[0].ID)
	}
}

func TestListNoticesUnlimitedAndOffsetClamped(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Notice{})

	n1 := Notice{ID: 1, Version: "1", BtnTitle: "Btn", Title: "Title1", TitleImage: "Img", TimeDesc: "Now", Content: "Body1", TagType: 1, Icon: 1, Track: "T"}
	n2 := Notice{ID: 2, Version: "2", BtnTitle: "Btn", Title: "Title2", TitleImage: "Img", TimeDesc: "Now", Content: "Body2", TagType: 2, Icon: 2, Track: "T"}
	if err := n1.Create(); err != nil {
		t.Fatalf("seed notice1: %v", err)
	}
	if err := n2.Create(); err != nil {
		t.Fatalf("seed notice2: %v", err)
	}

	noticeResult, err := ListNotices(NoticeQueryParams{Offset: -2, Limit: 0})
	if err != nil {
		t.Fatalf("list notices: %v", err)
	}
	if noticeResult.Total != 2 {
		t.Fatalf("expected total 2, got %d", noticeResult.Total)
	}
	if len(noticeResult.Notices) != 2 {
		t.Fatalf("expected 2 notices, got %d", len(noticeResult.Notices))
	}
	if noticeResult.Notices[0].ID != 2 || noticeResult.Notices[1].ID != 1 {
		t.Fatalf("expected notices ordered by id desc, got ids %d,%d", noticeResult.Notices[0].ID, noticeResult.Notices[1].ID)
	}

	secondPage, err := ListNotices(NoticeQueryParams{Offset: 1, Limit: 1})
	if err != nil {
		t.Fatalf("list notices page: %v", err)
	}
	if secondPage.Total != 2 {
		t.Fatalf("expected paged total 2, got %d", secondPage.Total)
	}
	if len(secondPage.Notices) != 1 {
		t.Fatalf("expected 1 notice on page, got %d", len(secondPage.Notices))
	}
	if secondPage.Notices[0].ID != 1 {
		t.Fatalf("expected offset=1 limit=1 to return id 1, got %d", secondPage.Notices[0].ID)
	}
}

func TestListExchangeCodesUnlimited(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ExchangeCodeRedeem{})
	clearTable(t, &ExchangeCode{})

	code1 := ExchangeCode{Code: "ABC123", Platform: "ios", Quota: 1, Rewards: json.RawMessage(`[]`)}
	code2 := ExchangeCode{Code: "DEF456", Platform: "android", Quota: 1, Rewards: json.RawMessage(`[]`)}
	if err := CreateExchangeCode(&code1); err != nil {
		t.Fatalf("seed code1: %v", err)
	}
	if err := CreateExchangeCode(&code2); err != nil {
		t.Fatalf("seed code2: %v", err)
	}

	codes, total, err := ListExchangeCodes(-10, -1)
	if err != nil {
		t.Fatalf("list exchange codes: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(codes) != 2 {
		t.Fatalf("expected 2 codes, got %d", len(codes))
	}
	if codes[0].Code != "ABC123" || codes[1].Code != "DEF456" {
		t.Fatalf("expected codes ordered by id asc, got %s,%s", codes[0].Code, codes[1].Code)
	}

	page, pageTotal, err := ListExchangeCodes(1, 1)
	if err != nil {
		t.Fatalf("list exchange codes page: %v", err)
	}
	if pageTotal != 2 {
		t.Fatalf("expected paged total 2, got %d", pageTotal)
	}
	if len(page) != 1 {
		t.Fatalf("expected 1 code on page, got %d", len(page))
	}
	if page[0].Code != "DEF456" {
		t.Fatalf("expected offset=1 limit=1 to return DEF456, got %s", page[0].Code)
	}
}

func TestCommanderExistsIgnoresSoftDeletedRows(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Commander{})

	activeID := uint32(9001)
	deletedID := uint32(9002)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled)
VALUES ($1, $2, $3, 0, $4, now(), 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false)`, int64(activeID), int64(activeID), 1, "Active"); err != nil {
		t.Fatalf("seed active commander: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled, deleted_at)
VALUES ($1, $2, $3, 0, $4, now(), 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false, now())`, int64(deletedID), int64(deletedID), 1, "Deleted"); err != nil {
		t.Fatalf("seed deleted commander: %v", err)
	}

	if err := CommanderExists(activeID); err != nil {
		t.Fatalf("active commander expected to exist: %v", err)
	}
	if err := CommanderExists(deletedID); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected soft-deleted commander to be hidden, got %v", err)
	}
	if err := CommanderExists(9999); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected missing commander to be hidden, got %v", err)
	}
}
