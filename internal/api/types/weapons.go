package types

type WeaponPayload struct {
	ID                   uint32  `json:"id"`
	ActionIndex          string  `json:"action_index"`
	AimType              int     `json:"aim_type"`
	Angle                int     `json:"angle"`
	AttackAttribute      int     `json:"attack_attribute"`
	AttackAttributeRatio int     `json:"attack_attribute_ratio"`
	AutoAftercast        RawJSON `json:"auto_aftercast"`
	AxisAngle            int     `json:"axis_angle"`
	BarrageID            RawJSON `json:"barrage_ID"`
	BulletID             RawJSON `json:"bullet_ID"`
	ChargeParam          RawJSON `json:"charge_param"`
	Corrected            int     `json:"corrected"`
	Damage               int     `json:"damage"`
	EffectMove           int     `json:"effect_move"`
	Expose               int     `json:"expose"`
	FireFX               string  `json:"fire_fx"`
	FireFXLoopType       int     `json:"fire_fx_loop_type"`
	FireSFX              string  `json:"fire_sfx"`
	InitialOverHeat      int     `json:"initial_over_heat"`
	MinRange             int     `json:"min_range"`
	OxyType              RawJSON `json:"oxy_type"`
	PrecastParam         RawJSON `json:"precast_param"`
	Queue                int     `json:"queue"`
	Range                int     `json:"range"`
	RecoverTime          RawJSON `json:"recover_time"`
	ReloadMax            int     `json:"reload_max"`
	SearchCondition      RawJSON `json:"search_condition"`
	SearchType           int     `json:"search_type"`
	ShakeScreen          int     `json:"shakescreen"`
	SpawnBound           RawJSON `json:"spawn_bound"`
	Suppress             int     `json:"suppress"`
	TorpedoAmmo          int     `json:"torpedo_ammo"`
	Type                 int     `json:"type"`
}

type WeaponListResponse struct {
	Weapons []WeaponPayload `json:"weapons"`
	Meta    PaginationMeta  `json:"meta"`
}
