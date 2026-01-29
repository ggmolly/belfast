package types

type SkillPayload struct {
	ID         uint32  `json:"id"`
	Name       string  `json:"name"`
	Desc       string  `json:"desc"`
	CD         uint32  `json:"cd"`
	Painting   RawJSON `json:"painting"`
	Picture    string  `json:"picture"`
	AniEffect  RawJSON `json:"aniEffect"`
	UIEffect   string  `json:"uiEffect"`
	EffectList RawJSON `json:"effect_list"`
}

type SkillListResponse struct {
	Skills []SkillPayload `json:"skills"`
	Meta   PaginationMeta `json:"meta"`
}
