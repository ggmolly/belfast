package orm

type Buff struct {
	ID          uint32 `gorm:"primary_key" json:"id"`
	Name        string `gorm:"size:50" json:"name"`
	Description string `gorm:"size:170" json:"desc"`
	MaxTime     int    `gorm:"default:0;not_null" json:"max_time"`
	BenefitType string `gorm:"size:50;not_null" json:"benefit_type"`
}
