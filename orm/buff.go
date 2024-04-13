package orm

type Buff struct {
	ID          uint32 `gorm:"primary_key"`
	Name        string `gorm:"size:50"`
	Description string `gorm:"size:170"`
	MaxTime     int    `gorm:"default:0;not null"`
	BenefitType string `gorm:"size:50;not null"`
}
