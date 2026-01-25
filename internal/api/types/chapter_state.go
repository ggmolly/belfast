package types

type ChapterCellPos struct {
	Row    uint32 `json:"row"`
	Column uint32 `json:"column"`
}

type ChapterCellInfo struct {
	Pos      ChapterCellPos `json:"pos"`
	ItemType uint32         `json:"item_type"`
	ItemID   *uint32        `json:"item_id,omitempty"`
	ItemFlag *uint32        `json:"item_flag,omitempty"`
	ItemData *uint32        `json:"item_data,omitempty"`
	ExtraID  []uint32       `json:"extra_id"`
}

type ChapterShip struct {
	ID     uint32 `json:"id"`
	HpRant uint32 `json:"hp_rant"`
}

type ChapterCommander struct {
	Pos uint32 `json:"pos"`
	ID  uint32 `json:"id"`
}

type ChapterStrategy struct {
	ID    uint32 `json:"id"`
	Count uint32 `json:"count"`
}

type ChapterGroup struct {
	ID               uint32             `json:"id"`
	ShipList         []ChapterShip      `json:"ship_list"`
	Pos              ChapterCellPos     `json:"pos"`
	StepCount        uint32             `json:"step_count"`
	BoxStrategyList  []ChapterStrategy  `json:"box_strategy_list"`
	ShipStrategyList []ChapterStrategy  `json:"ship_strategy_list"`
	StrategyIds      []uint32           `json:"strategy_ids"`
	Bullet           uint32             `json:"bullet"`
	StartPos         ChapterCellPos     `json:"start_pos"`
	CommanderList    []ChapterCommander `json:"commander_list"`
	MoveStepDown     uint32             `json:"move_step_down"`
	KillCount        uint32             `json:"kill_count"`
	FleetId          uint32             `json:"fleet_id"`
	VisionLv         uint32             `json:"vision_lv"`
}

type ChapterCellFlag struct {
	Pos      ChapterCellPos `json:"pos"`
	FlagList []uint32       `json:"flag_list"`
}

type ChapterFleetDuty struct {
	Key   uint32 `json:"key"`
	Value uint32 `json:"value"`
}

type ChapterState struct {
	ID                    uint32             `json:"id"`
	Time                  uint32             `json:"time"`
	CellList              []ChapterCellInfo  `json:"cell_list"`
	MainGroupList         []ChapterGroup     `json:"main_group_list"`
	AiList                []ChapterCellInfo  `json:"ai_list"`
	EscortList            []ChapterCellInfo  `json:"escort_list"`
	Round                 uint32             `json:"round"`
	IsSubmarineAutoAttack uint32             `json:"is_submarine_auto_attack"`
	OperationBuff         []uint32           `json:"operation_buff"`
	ModelActCount         uint32             `json:"model_act_count"`
	BuffList              []uint32           `json:"buff_list"`
	LoopFlag              uint32             `json:"loop_flag"`
	ExtraFlagList         []uint32           `json:"extra_flag_list"`
	CellFlagList          []ChapterCellFlag  `json:"cell_flag_list"`
	ChapterHp             uint32             `json:"chapter_hp"`
	ChapterStrategyList   []ChapterStrategy  `json:"chapter_strategy_list"`
	KillCount             uint32             `json:"kill_count"`
	InitShipCount         uint32             `json:"init_ship_count"`
	ContinuousKillCount   uint32             `json:"continuous_kill_count"`
	BattleStatistics      []ChapterStrategy  `json:"battle_statistics"`
	FleetDuties           []ChapterFleetDuty `json:"fleet_duties"`
	MoveStepCount         uint32             `json:"move_step_count"`
	SubmarineGroupList    []ChapterGroup     `json:"submarine_group_list"`
	SupportGroupList      []ChapterGroup     `json:"support_group_list"`
}

type PlayerChapterStateResponse struct {
	ChapterID uint32       `json:"chapter_id"`
	UpdatedAt uint32       `json:"updated_at"`
	State     ChapterState `json:"state"`
}

type PlayerChapterStateListResponse struct {
	States []PlayerChapterStateResponse `json:"states"`
	Meta   PaginationMeta               `json:"meta"`
}

type PlayerChapterStateCreateRequest struct {
	State ChapterState `json:"state" validate:"required"`
}

type PlayerChapterStateUpdateRequest struct {
	State ChapterState `json:"state" validate:"required"`
}
