package orm

type Buff struct {
	ID          uint32 `gorm:"primary_key" json:"id"`
	Name        string `gorm:"size:256;not_null" json:"name"`
	Description string `gorm:"type:text;not_null" json:"desc"`
	MaxTime     int    `gorm:"default:0;not_null" json:"max_time"`
	BenefitType string `gorm:"size:50;not_null" json:"benefit_type"`
}
