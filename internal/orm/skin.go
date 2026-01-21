package orm

import (
	"encoding/json"
	"time"
)

type Skin struct {
	ID             uint32          `gorm:"primary_key" json:"id"`
	Name           string          `gorm:"size:128;not_null" json:"name"`
	ShipGroup      int             `gorm:"not_null" json:"ship_group"`
	Desc           string          `gorm:"type:text" json:"desc"`
	BG             string          `gorm:"size:128" json:"bg"`
	BGSp           string          `gorm:"size:128" json:"bg_sp"`
	BGM            string          `gorm:"size:128" json:"bgm"`
	Painting       string          `gorm:"size:64" json:"painting"`
	Prefab         string          `gorm:"size:64" json:"prefab"`
	ChangeSkin     json.RawMessage `gorm:"type:json" json:"change_skin"`
	ShowSkin       string          `gorm:"size:64" json:"show_skin"`
	SkeletonSkin   string          `gorm:"size:64" json:"skeleton_default_skin"`
	ShipL2DID      json.RawMessage `gorm:"type:json" json:"ship_l2d_id"`
	L2DAnimations  json.RawMessage `gorm:"type:json" json:"l2d_animations"`
	L2DDragRate    json.RawMessage `gorm:"type:json" json:"l2d_drag_rate"`
	L2DParaRange   json.RawMessage `gorm:"type:json" json:"l2d_para_range"`
	L2DSE          json.RawMessage `gorm:"type:json" json:"l2d_se"`
	L2DVoiceCalib  json.RawMessage `gorm:"type:json" json:"l2d_voice_calibrate"`
	PartScale      string          `gorm:"size:64" json:"part_scale"`
	MainUIFX       string          `gorm:"size:128" json:"main_UI_FX"`
	SpineOffset    json.RawMessage `gorm:"type:json" json:"spine_offset"`
	SpineProfile   json.RawMessage `gorm:"type:json" json:"spine_offset_profile"`
	Tag            json.RawMessage `gorm:"type:json" json:"tag"`
	Time           json.RawMessage `gorm:"type:json" json:"time"`
	GetShowing     json.RawMessage `gorm:"type:json" json:"get_showing"`
	PurchaseOffset json.RawMessage `gorm:"type:json" json:"purchase_offset"`
	ShopOffset     json.RawMessage `gorm:"type:json" json:"shop_offset"`
	RarityBG       string          `gorm:"size:128" json:"rarity_bg"`
	SpecialEffects json.RawMessage `gorm:"type:json" json:"special_effects"`
	GroupIndex     *int            `gorm:"type:int" json:"group_index"`
	Gyro           *int            `gorm:"type:int" json:"gyro"`
	HandID         *int            `gorm:"type:int" json:"hand_id"`
	Illustrator    *int            `gorm:"type:int" json:"illustrator"`
	Illustrator2   *int            `gorm:"type:int" json:"illustrator2"`
	VoiceActor     *int            `gorm:"type:int" json:"voice_actor"`
	VoiceActor2    *int            `gorm:"type:int" json:"voice_actor_2"`
	DoubleChar     *int            `gorm:"type:int" json:"double_char"`
	LipSmoothing   *int            `gorm:"type:int" json:"lip_smoothing"`
	LipSyncGain    *int            `gorm:"type:int" json:"lip_sync_gain"`
	L2DIgnoreDrag  *int            `gorm:"type:int" json:"l2d_ignore_drag"`
	SkinType       *int            `gorm:"type:int" json:"skin_type"`
	ShopID         *int            `gorm:"type:int" json:"shop_id"`
	ShopTypeID     *int            `gorm:"type:int" json:"shop_type_id"`
	ShopDynamicHX  *int            `gorm:"type:int" json:"shop_dynamic_hx"`
	SpineAction    json.RawMessage `gorm:"type:json" json:"spine_action_offset"`
	SpineUseLive2D *int            `gorm:"type:int" json:"spine_use_live2d"`
	Live2DOffset   json.RawMessage `gorm:"type:json" json:"live2d_offset"`
	Live2DProfile  json.RawMessage `gorm:"type:json" json:"live2d_offset_profile"`
	FXContainer    json.RawMessage `gorm:"type:json" json:"fx_container"`
	BoundBone      json.RawMessage `gorm:"type:json" json:"bound_bone"`
	Smoke          json.RawMessage `gorm:"type:json" json:"smoke"`
}

type OwnedSkin struct {
	CommanderID uint32     `gorm:"primaryKey"`
	SkinID      uint32     `gorm:"primaryKey"`
	ExpiresAt   *time.Time `gorm:"not_null"`
}
