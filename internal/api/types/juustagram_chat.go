package types

type JuustagramReply struct {
	Sequence uint32 `json:"sequence"`
	Key      uint32 `json:"key"`
	Value    uint32 `json:"value"`
}

type JuustagramChatGroup struct {
	ChatGroupID uint32            `json:"chat_group_id"`
	OpTime      uint32            `json:"op_time"`
	ReadFlag    uint32            `json:"read_flag"`
	ReplyList   []JuustagramReply `json:"reply_list"`
}

type JuustagramGroup struct {
	GroupID      uint32                `json:"group_id"`
	SkinID       uint32                `json:"skin_id"`
	Favorite     uint32                `json:"favorite"`
	CurChatGroup uint32                `json:"cur_chat_group"`
	ChatGroups   []JuustagramChatGroup `json:"chat_groups"`
}

type JuustagramGroupListResponse struct {
	Groups []JuustagramGroup `json:"groups"`
	Meta   PaginationMeta    `json:"meta"`
}

type JuustagramGroupResponse struct {
	Group JuustagramGroup `json:"group"`
}

type JuustagramGroupCreateRequest struct {
	GroupID     uint32 `json:"group_id"`
	ChatGroupID uint32 `json:"chat_group_id"`
	SkinID      uint32 `json:"skin_id"`
	Favorite    uint32 `json:"favorite"`
}

type JuustagramGroupUpdateRequest struct {
	SkinID       *uint32 `json:"skin_id"`
	Favorite     *uint32 `json:"favorite"`
	CurChatGroup *uint32 `json:"cur_chat_group"`
}

type JuustagramChatGroupCreateRequest struct {
	ChatGroupID uint32 `json:"chat_group_id"`
	OpTime      uint32 `json:"op_time"`
}

type JuustagramChatReplyRequest struct {
	ChatID uint32 `json:"chat_id"`
	Value  uint32 `json:"value"`
}

type JuustagramChatReadRequest struct {
	ChatGroupIDs []uint32 `json:"chat_group_ids"`
}
