package orm

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
)

func mapSQLCNotFound(err error) error {
	return db.MapNotFound(err)
}

func pgTextFromPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func pgTextPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func pgInt8FromUint32Ptr(value *uint32) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(*value), Valid: true}
}

func pgInt8PtrToUint32Ptr(value pgtype.Int8) *uint32 {
	if !value.Valid {
		return nil
	}
	v := uint32(value.Int64)
	return &v
}

func pgTimestamptzFromPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}

func pgTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgTimestamptzPtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}

func pgBoolFromPtr(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *value, Valid: true}
}

func pgBoolPtr(value pgtype.Bool) *bool {
	if !value.Valid {
		return nil
	}
	v := value.Bool
	return &v
}
