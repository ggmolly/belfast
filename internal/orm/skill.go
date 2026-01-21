package orm

import "encoding/json"

type Skill struct {
	ID         uint32          `gorm:"primary_key" json:"id"`
	Name       string          `gorm:"size:128;not_null" json:"name"`
	Desc       string          `gorm:"type:text" json:"desc"`
	CD         uint32          `gorm:"not_null" json:"cd"`
	Painting   json.RawMessage `gorm:"type:json" json:"painting"`
	Picture    string          `gorm:"size:64" json:"picture"`
	AniEffect  json.RawMessage `gorm:"type:json" json:"aniEffect"`
	UIEffect   string          `gorm:"type:text" json:"uiEffect"`
	EffectList json.RawMessage `gorm:"type:json" json:"effect_list"`
}
