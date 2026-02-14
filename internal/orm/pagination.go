package orm

func normalizePagination(offset int, limit int) (int, int, bool) {
	if offset < 0 {
		offset = 0
	}

	if limit <= 0 {
		return offset, 0, true
	}

	return offset, limit, false
}
