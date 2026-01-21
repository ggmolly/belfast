package orm

import "encoding/json"

type Weapon struct {
	ID                   uint32          `gorm:"primary_key" json:"id"`
	ActionIndex          string          `gorm:"size:32;not_null" json:"action_index"`
	AimType              int             `gorm:"not_null" json:"aim_type"`
	Angle                int             `gorm:"not_null" json:"angle"`
	AttackAttribute      int             `gorm:"not_null" json:"attack_attribute"`
	AttackAttributeRatio int             `gorm:"not_null" json:"attack_attribute_ratio"`
	AutoAftercast        json.RawMessage `gorm:"type:json" json:"auto_aftercast"`
	AxisAngle            int             `gorm:"not_null" json:"axis_angle"`
	BarrageID            json.RawMessage `gorm:"type:json" json:"barrage_ID"`
	BulletID             json.RawMessage `gorm:"type:json" json:"bullet_ID"`
	ChargeParam          json.RawMessage `gorm:"type:json" json:"charge_param"`
	Corrected            int             `gorm:"not_null" json:"corrected"`
	Damage               int             `gorm:"not_null" json:"damage"`
	EffectMove           int             `gorm:"not_null" json:"effect_move"`
	Expose               int             `gorm:"not_null" json:"expose"`
	FireFX               string          `gorm:"size:64" json:"fire_fx"`
	FireFXLoopType       int             `gorm:"not_null" json:"fire_fx_loop_type"`
	FireSFX              string          `gorm:"size:64" json:"fire_sfx"`
	InitialOverHeat      int             `gorm:"not_null" json:"initial_over_heat"`
	MinRange             int             `gorm:"not_null" json:"min_range"`
	OxyType              json.RawMessage `gorm:"type:json" json:"oxy_type"`
	PrecastParam         json.RawMessage `gorm:"type:json" json:"precast_param"`
	Queue                int             `gorm:"not_null" json:"queue"`
	Range                int             `gorm:"not_null" json:"range"`
	RecoverTime          json.RawMessage `gorm:"type:json" json:"recover_time"`
	ReloadMax            int             `gorm:"not_null" json:"reload_max"`
	SearchCondition      json.RawMessage `gorm:"type:json" json:"search_condition"`
	SearchType           int             `gorm:"not_null" json:"search_type"`
	ShakeScreen          int             `gorm:"not_null" json:"shakescreen"`
	SpawnBound           json.RawMessage `gorm:"type:json" json:"spawn_bound"`
	Suppress             int             `gorm:"not_null" json:"suppress"`
	TorpedoAmmo          int             `gorm:"not_null" json:"torpedo_ammo"`
	Type                 int             `gorm:"not_null" json:"type"`
}
