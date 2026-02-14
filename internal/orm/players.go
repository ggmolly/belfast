package orm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

type PlayerQueryParams struct {
	Offset       int
	Limit        int
	FilterBanned bool
	FilterOnline bool
	OnlineIDs    []uint32
	MinLevel     int
	Search       string
}

type PlayerListResult struct {
	Commanders []Commander
	Total      int64
}

type PlayerBanStatus struct {
	Banned   bool
	LiftTime *time.Time
}

func ListCommanders(params PlayerQueryParams) (PlayerListResult, error) {
	return listOrSearchCommanders(params, false)
}

func SearchCommanders(params PlayerQueryParams) (PlayerListResult, error) {
	return listOrSearchCommanders(params, true)
}

func listOrSearchCommanders(params PlayerQueryParams, includeSearch bool) (PlayerListResult, error) {
	if db.DefaultStore == nil {
		return PlayerListResult{}, errors.New("db not initialized")
	}
	ctx := context.Background()

	whereSQL, args := buildPlayerFilters(params, includeSearch)

	countSQL := "SELECT COUNT(1) FROM commanders" + whereSQL
	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return PlayerListResult{}, err
	}

	offset, limit, unlimited := normalizePagination(params.Offset, params.Limit)
	if !unlimited {
		if limit > 500 {
			limit = 500
		}
	}

	listSQL := "SELECT commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled, deleted_at FROM commanders" + whereSQL + " ORDER BY last_login DESC OFFSET $" + fmt.Sprint(len(args)+1)
	args = append(args, offset)
	if !unlimited {
		listSQL += " LIMIT $" + fmt.Sprint(len(args)+1)
		args = append(args, limit)
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, listSQL, args...)
	if err != nil {
		return PlayerListResult{}, err
	}
	defer rows.Close()

	commanders := make([]Commander, 0)
	if !unlimited {
		commanders = make([]Commander, 0, limit)
	}
	for rows.Next() {
		var c Commander
		var deletedAt *time.Time
		if err := rows.Scan(
			&c.CommanderID,
			&c.AccountID,
			&c.Level,
			&c.Exp,
			&c.Name,
			&c.LastLogin,
			&c.GuideIndex,
			&c.NewGuideIndex,
			&c.NameChangeCooldown,
			&c.RoomID,
			&c.ExchangeCount,
			&c.DrawCount1,
			&c.DrawCount10,
			&c.SupportRequisitionCount,
			&c.SupportRequisitionMonth,
			&c.CollectAttackCount,
			&c.AccPayLv,
			&c.LivingAreaCoverID,
			&c.SelectedIconFrameID,
			&c.SelectedChatFrameID,
			&c.SelectedBattleUIID,
			&c.DisplayIconID,
			&c.DisplaySkinID,
			&c.DisplayIconThemeID,
			&c.Manifesto,
			&c.DormName,
			&c.RandomShipMode,
			&c.RandomFlagShipEnabled,
			&deletedAt,
		); err != nil {
			return PlayerListResult{}, err
		}
		if deletedAt != nil {
			c.DeletedAt = deletedAt
		}
		commanders = append(commanders, c)
	}
	if err := rows.Err(); err != nil {
		return PlayerListResult{}, err
	}
	return PlayerListResult{Commanders: commanders, Total: total}, nil
}

func buildPlayerFilters(params PlayerQueryParams, includeSearch bool) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 8)
	idx := 1
	clauses = append(clauses, "deleted_at IS NULL")

	if params.MinLevel > 0 {
		clauses = append(clauses, fmt.Sprintf("level >= $%d", idx))
		args = append(args, params.MinLevel)
		idx++
	}
	if params.FilterBanned {
		now := time.Now().UTC()
		clauses = append(clauses, fmt.Sprintf("EXISTS (SELECT 1 FROM punishments WHERE punishments.punished_id = commanders.commander_id AND (punishments.lift_timestamp IS NULL OR punishments.lift_timestamp > $%d))", idx))
		args = append(args, now)
		idx++
	}
	if params.FilterOnline {
		if len(params.OnlineIDs) > 0 {
			clauses = append(clauses, fmt.Sprintf("commander_id = ANY($%d::bigint[])", idx))
			ids := make([]int64, 0, len(params.OnlineIDs))
			for _, v := range params.OnlineIDs {
				ids = append(ids, int64(v))
			}
			args = append(args, ids)
			idx++
		} else {
			clauses = append(clauses, "1 = 0")
		}
	}
	if includeSearch {
		if strings.TrimSpace(params.Search) != "" {
			search := strings.ToLower(strings.TrimSpace(params.Search))
			clauses = append(clauses, fmt.Sprintf("LOWER(name) LIKE $%d", idx))
			args = append(args, "%"+search+"%")
			idx++
		}
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func LoadCommanderWithDetails(id uint32) (Commander, error) {
	return loadCommanderWithDetailsSQLC(id)
}

func GetBanStatus(commanderID uint32) (PlayerBanStatus, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `SELECT id, punished_id, lift_timestamp, is_permanent FROM punishments WHERE punished_id = $1 ORDER BY id DESC LIMIT 1`, int64(commanderID))
	var p Punishment
	var lift *time.Time
	if err := row.Scan(&p.ID, &p.PunishedID, &lift, &p.IsPermanent); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PlayerBanStatus{Banned: false, LiftTime: nil}, nil
		}
		return PlayerBanStatus{}, err
	}
	p.LiftTimestamp = lift
	if p.LiftTimestamp != nil {
		if time.Now().UTC().Before(*p.LiftTimestamp) {
			return PlayerBanStatus{Banned: true, LiftTime: p.LiftTimestamp}, nil
		}
		return PlayerBanStatus{Banned: false, LiftTime: nil}, nil
	}
	return PlayerBanStatus{Banned: true, LiftTime: nil}, nil
}

func ActivePunishment(commanderID uint32) (*Punishment, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `SELECT id, punished_id, lift_timestamp, is_permanent FROM punishments WHERE punished_id = $1 ORDER BY id DESC LIMIT 1`, int64(commanderID))
	var p Punishment
	var lift *time.Time
	if err := row.Scan(&p.ID, &p.PunishedID, &lift, &p.IsPermanent); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}
	p.LiftTimestamp = lift
	if p.LiftTimestamp != nil && time.Now().UTC().After(*p.LiftTimestamp) {
		return nil, db.ErrNotFound
	}
	return &p, nil
}

func CommanderIDExists(commanderID uint32) (bool, error) {
	ctx := context.Background()
	var exists bool
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM commanders
	WHERE commander_id = $1
	  AND deleted_at IS NULL
)
`, int64(commanderID)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CommanderNameExists(name string) (bool, error) {
	ctx := context.Background()
	var exists bool
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM commanders
	WHERE name = $1
	  AND deleted_at IS NULL
)
`, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CommanderNameExistsExcept(name string, commanderID uint32) (bool, error) {
	ctx := context.Background()
	var exists bool
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM commanders
	WHERE name = $1
	  AND commander_id <> $2
	  AND deleted_at IS NULL
)
`, name, int64(commanderID)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateCommanderRoot(commanderID uint32, accountID uint32, name string, guideIndex uint32, newGuideIndex uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO commanders (
	commander_id,
	account_id,
	level,
	exp,
	name,
	last_login,
	guide_index,
	new_guide_index,
	name_change_cooldown,
	room_id,
	exchange_count,
	draw_count1,
	draw_count10,
	support_requisition_count,
	support_requisition_month,
	collect_attack_count,
	acc_pay_lv,
	living_area_cover_id,
	selected_icon_frame_id,
	selected_chat_frame_id,
	selected_battle_ui_id,
	display_icon_id,
	display_skin_id,
	display_icon_theme_id,
	manifesto,
	dorm_name,
	random_ship_mode,
	random_flag_ship_enabled
) VALUES (
	$1, $2, 1, 0, $3, now(), $4, $5, '1970-01-01 00:00:00+00',
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false
)
`, int64(commanderID), int64(accountID), name, int64(guideIndex), int64(newGuideIndex))
	return err
}

func GetCommanderCoreByID(commanderID uint32) (*Commander, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT
	commander_id,
	account_id,
	level,
	exp,
	name,
	last_login,
	guide_index,
	new_guide_index,
	name_change_cooldown,
	room_id,
	exchange_count,
	draw_count1,
	draw_count10,
	support_requisition_count,
	support_requisition_month,
	collect_attack_count,
	acc_pay_lv,
	living_area_cover_id,
	selected_icon_frame_id,
	selected_chat_frame_id,
	selected_battle_ui_id,
	display_icon_id,
	display_skin_id,
	display_icon_theme_id,
	manifesto,
	dorm_name,
	random_ship_mode,
	random_flag_ship_enabled,
	deleted_at
FROM commanders
WHERE commander_id = $1
	AND deleted_at IS NULL
`, int64(commanderID))

	commander := Commander{}
	var deletedAt *time.Time
	err := row.Scan(
		&commander.CommanderID,
		&commander.AccountID,
		&commander.Level,
		&commander.Exp,
		&commander.Name,
		&commander.LastLogin,
		&commander.GuideIndex,
		&commander.NewGuideIndex,
		&commander.NameChangeCooldown,
		&commander.RoomID,
		&commander.ExchangeCount,
		&commander.DrawCount1,
		&commander.DrawCount10,
		&commander.SupportRequisitionCount,
		&commander.SupportRequisitionMonth,
		&commander.CollectAttackCount,
		&commander.AccPayLv,
		&commander.LivingAreaCoverID,
		&commander.SelectedIconFrameID,
		&commander.SelectedChatFrameID,
		&commander.SelectedBattleUIID,
		&commander.DisplayIconID,
		&commander.DisplaySkinID,
		&commander.DisplayIconThemeID,
		&commander.Manifesto,
		&commander.DormName,
		&commander.RandomShipMode,
		&commander.RandomFlagShipEnabled,
		&deletedAt,
	)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	commander.DeletedAt = deletedAt
	return &commander, nil
}

func DeleteCommander(commanderID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE commanders
SET deleted_at = now()
WHERE commander_id = $1
  AND deleted_at IS NULL
`, int64(commanderID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
