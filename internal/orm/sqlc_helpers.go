package orm

import (
	"fmt"
	"math"
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
	v, err := uint32FromInt64Checked(value.Int64)
	if err != nil {
		return nil
	}
	return &v
}

func pgInt8PtrToUint32PtrChecked(value pgtype.Int8) (*uint32, error) {
	if !value.Valid {
		return nil, nil
	}
	v, err := uint32FromInt64Checked(value.Int64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func Uint32FromInt64Checked(value int64) (uint32, error) {
	return uint32FromInt64Checked(value)
}

func uint32FromInt64Checked(value int64) (uint32, error) {
	if value < 0 || value > math.MaxUint32 {
		return 0, fmt.Errorf("value %d out of uint32 range", value)
	}
	return uint32(value), nil
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

type rowScanner interface {
	Scan(dest ...any) error
}

type anyRows interface {
	rowScanner
	Close()
	Err() error
	Next() bool
}
