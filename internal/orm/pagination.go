package orm

import "gorm.io/gorm"

func ApplyPagination(query *gorm.DB, offset int, limit int) *gorm.DB {
	query = query.Offset(offset)
	if limit > 0 {
		query = query.Limit(limit)
	}
	return query
}
