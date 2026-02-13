package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

type ChapterProgress struct {
	CommanderID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	ChapterID        uint32 `gorm:"primaryKey;autoIncrement:false"`
	Progress         uint32 `gorm:"not_null"`
	KillBossCount    uint32 `gorm:"not_null"`
	KillEnemyCount   uint32 `gorm:"not_null"`
	TakeBoxCount     uint32 `gorm:"not_null"`
	DefeatCount      uint32 `gorm:"not_null"`
	TodayDefeatCount uint32 `gorm:"not_null"`
	PassCount        uint32 `gorm:"not_null"`
	UpdatedAt        uint32 `gorm:"not_null"`
}

func GetChapterProgress(commanderID uint32, chapterID uint32) (*ChapterProgress, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, chapter_id, progress, kill_boss_count, kill_enemy_count, take_box_count, defeat_count, today_defeat_count, pass_count, updated_at
FROM chapter_progress
WHERE commander_id = $1 AND chapter_id = $2
`, int64(commanderID), int64(chapterID))
	var progress ChapterProgress
	err := row.Scan(
		&progress.CommanderID,
		&progress.ChapterID,
		&progress.Progress,
		&progress.KillBossCount,
		&progress.KillEnemyCount,
		&progress.TakeBoxCount,
		&progress.DefeatCount,
		&progress.TodayDefeatCount,
		&progress.PassCount,
		&progress.UpdatedAt,
	)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func ListChapterProgress(commanderID uint32) ([]ChapterProgress, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT commander_id, chapter_id, progress, kill_boss_count, kill_enemy_count, take_box_count, defeat_count, today_defeat_count, pass_count, updated_at
FROM chapter_progress
WHERE commander_id = $1
ORDER BY chapter_id ASC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	progress := make([]ChapterProgress, 0)
	for rows.Next() {
		var row ChapterProgress
		if err := rows.Scan(&row.CommanderID, &row.ChapterID, &row.Progress, &row.KillBossCount, &row.KillEnemyCount, &row.TakeBoxCount, &row.DefeatCount, &row.TodayDefeatCount, &row.PassCount, &row.UpdatedAt); err != nil {
			return nil, err
		}
		progress = append(progress, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return progress, nil
}

func UpsertChapterProgress(progress *ChapterProgress) error {
	ctx := context.Background()
	progress.UpdatedAt = uint32(time.Now().Unix())
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO chapter_progress (
  commander_id,
  chapter_id,
  progress,
  kill_boss_count,
  kill_enemy_count,
  take_box_count,
  defeat_count,
  today_defeat_count,
  pass_count,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (commander_id, chapter_id)
DO UPDATE SET
  progress = EXCLUDED.progress,
  kill_boss_count = EXCLUDED.kill_boss_count,
  kill_enemy_count = EXCLUDED.kill_enemy_count,
  take_box_count = EXCLUDED.take_box_count,
  defeat_count = EXCLUDED.defeat_count,
  today_defeat_count = EXCLUDED.today_defeat_count,
  pass_count = EXCLUDED.pass_count,
  updated_at = EXCLUDED.updated_at
`, int64(progress.CommanderID), int64(progress.ChapterID), int64(progress.Progress), int64(progress.KillBossCount), int64(progress.KillEnemyCount), int64(progress.TakeBoxCount), int64(progress.DefeatCount), int64(progress.TodayDefeatCount), int64(progress.PassCount), int64(progress.UpdatedAt))
	return err
}

func DeleteChapterProgress(commanderID uint32, chapterID uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM chapter_progress WHERE commander_id = $1 AND chapter_id = $2`, int64(commanderID), int64(chapterID))
	return err
}
