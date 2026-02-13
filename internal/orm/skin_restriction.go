package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

// GlobalSkinRestriction defines global hide/show restrictions for skins.
type GlobalSkinRestriction struct {
	SkinID uint32 `gorm:"primaryKey"`
	Type   uint32 `gorm:"not_null"`
}

// GlobalSkinRestrictionWindow defines time-based overrides for skin shop availability.
type GlobalSkinRestrictionWindow struct {
	ID        uint32 `gorm:"primaryKey"`
	SkinID    uint32 `gorm:"not_null"`
	Type      uint32 `gorm:"not_null"`
	StartTime uint32 `gorm:"not_null"`
	StopTime  uint32 `gorm:"not_null"`
}

func ListGlobalSkinRestrictions() ([]GlobalSkinRestriction, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT skin_id, type
FROM global_skin_restrictions
ORDER BY skin_id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]GlobalSkinRestriction, 0)
	for rows.Next() {
		var row GlobalSkinRestriction
		if err := rows.Scan(&row.SkinID, &row.Type); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func ListGlobalSkinRestrictionWindows() ([]GlobalSkinRestrictionWindow, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, skin_id, type, start_time, stop_time
FROM global_skin_restriction_windows
ORDER BY skin_id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]GlobalSkinRestrictionWindow, 0)
	for rows.Next() {
		var row GlobalSkinRestrictionWindow
		if err := rows.Scan(&row.ID, &row.SkinID, &row.Type, &row.StartTime, &row.StopTime); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
