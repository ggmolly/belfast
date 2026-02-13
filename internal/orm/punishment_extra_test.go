package orm

import (
	"context"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

func TestPunishmentCRUDAndRetrieve(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Punishment{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 90, AccountID: 90, Name: "Pun"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled)
VALUES ($1, $2, 1, 0, $3, now(), 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	lift := time.Now().Add(time.Hour)
	punish := Punishment{PunishedID: commander.CommanderID, LiftTimestamp: &lift}
	if err := punish.Create(); err != nil {
		t.Fatalf("create punishment: %v", err)
	}
	punish.IsPermanent = true
	if err := punish.Update(); err != nil {
		t.Fatalf("update punishment: %v", err)
	}
	loaded := Punishment{ID: punish.ID}
	if err := loaded.Retrieve(false); err != nil {
		t.Fatalf("retrieve punishment: %v", err)
	}
	if err := loaded.Retrieve(true); err != nil {
		t.Fatalf("retrieve punishment greedy: %v", err)
	}
	if err := loaded.Delete(); err != nil {
		t.Fatalf("delete punishment: %v", err)
	}
}
