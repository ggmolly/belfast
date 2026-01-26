package types

type ChapterProgress struct {
	ChapterID        uint32 `json:"chapter_id"`
	Progress         uint32 `json:"progress"`
	KillBossCount    uint32 `json:"kill_boss_count"`
	KillEnemyCount   uint32 `json:"kill_enemy_count"`
	TakeBoxCount     uint32 `json:"take_box_count"`
	DefeatCount      uint32 `json:"defeat_count"`
	TodayDefeatCount uint32 `json:"today_defeat_count"`
	PassCount        uint32 `json:"pass_count"`
	UpdatedAt        uint32 `json:"updated_at"`
}

type PlayerChapterProgressResponse struct {
	Progress ChapterProgress `json:"progress"`
}

type PlayerChapterProgressListResponse struct {
	Progress []PlayerChapterProgressResponse `json:"progress"`
	Meta     PaginationMeta                  `json:"meta"`
}

type PlayerChapterProgressCreateRequest struct {
	Progress ChapterProgress `json:"progress" validate:"required"`
}

type PlayerChapterProgressUpdateRequest struct {
	Progress ChapterProgress `json:"progress" validate:"required"`
}
