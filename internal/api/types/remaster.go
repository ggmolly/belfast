package types

import "time"

type PlayerRemasterStateResponse struct {
	TicketCount      uint32    `json:"ticket_count"`
	DailyCount       uint32    `json:"daily_count"`
	LastDailyResetAt time.Time `json:"last_daily_reset_at"`
}

type PlayerRemasterStateUpdateRequest struct {
	TicketCount      *uint32 `json:"ticket_count" validate:"omitempty"`
	DailyCount       *uint32 `json:"daily_count" validate:"omitempty"`
	LastDailyResetAt *string `json:"last_daily_reset_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type PlayerRemasterProgressEntry struct {
	ChapterID uint32    `json:"chapter_id"`
	Pos       uint32    `json:"pos"`
	Count     uint32    `json:"count"`
	Received  bool      `json:"received"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PlayerRemasterProgressResponse struct {
	Progress []PlayerRemasterProgressEntry `json:"progress"`
}

type PlayerRemasterProgressCreateRequest struct {
	ChapterID uint32 `json:"chapter_id" validate:"required,gt=0"`
	Pos       uint32 `json:"pos" validate:"required,gt=0"`
	Count     uint32 `json:"count" validate:"required"`
	Received  *bool  `json:"received"`
}

type PlayerRemasterProgressUpdateRequest struct {
	Count    *uint32 `json:"count" validate:"omitempty"`
	Received *bool   `json:"received"`
}
