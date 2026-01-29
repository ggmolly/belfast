package types

import "github.com/ggmolly/belfast/internal/orm"

type JuustagramTemplate = orm.JuustagramTemplate

type JuustagramNpcTemplate = orm.JuustagramNpcTemplate

type JuustagramLanguage = orm.JuustagramLanguage

type JuustagramShipGroupTemplate = orm.JuustagramShipGroupTemplate

type JuustagramMessageState = orm.JuustagramMessageState

type JuustagramPlayerDiscussState = orm.JuustagramPlayerDiscuss

type JuustagramLanguageListResponse struct {
	Entries []JuustagramLanguage `json:"entries"`
}

type JuustagramTemplateDeleteRequest struct {
	ID uint32 `json:"id"`
}

type JuustagramNpcTemplateDeleteRequest struct {
	ID uint32 `json:"id"`
}

type JuustagramShipGroupDeleteRequest struct {
	ShipGroup uint32 `json:"ship_group"`
}

type JuustagramLanguageDeleteRequest struct {
	Key string `json:"key"`
}

type JuustagramMessage struct {
	ID            uint32                    `json:"id"`
	Time          uint32                    `json:"time"`
	Text          string                    `json:"text"`
	Picture       string                    `json:"picture"`
	PlayerDiscuss []JuustagramPlayerDiscuss `json:"player_discuss"`
	NpcDiscuss    []JuustagramNpcComment    `json:"npc_discuss"`
	NpcReply      []JuustagramNpcComment    `json:"npc_reply"`
	Good          uint32                    `json:"good"`
	IsGood        uint32                    `json:"is_good"`
	IsRead        uint32                    `json:"is_read"`
}

type JuustagramPlayerDiscuss struct {
	ID       uint32   `json:"id"`
	Time     uint32   `json:"time"`
	TextList []string `json:"text_list"`
	Text     string   `json:"text"`
	NpcReply uint32   `json:"npc_reply"`
}

type JuustagramNpcComment struct {
	ID       uint32   `json:"id"`
	Time     uint32   `json:"time"`
	Text     string   `json:"text"`
	NpcReply []uint32 `json:"npc_reply"`
}

type JuustagramDiscussOption struct {
	DiscussID  uint32 `json:"discuss_id"`
	Index      uint32 `json:"index"`
	Text       string `json:"text"`
	NpcReplyID uint32 `json:"npc_reply_id"`
}

type JuustagramDiscussSelection struct {
	DiscussID   uint32 `json:"discuss_id"`
	OptionIndex uint32 `json:"option_index"`
	NpcReplyID  uint32 `json:"npc_reply_id"`
	CommentTime uint32 `json:"comment_time"`
}

type JuustagramTemplateListResponse struct {
	Templates []JuustagramTemplate `json:"templates"`
	Meta      PaginationMeta       `json:"meta"`
}

type JuustagramNpcTemplateListResponse struct {
	Templates []JuustagramNpcTemplate `json:"templates"`
	Meta      PaginationMeta          `json:"meta"`
}

type JuustagramShipGroupListResponse struct {
	Groups []JuustagramShipGroupTemplate `json:"groups"`
	Meta   PaginationMeta                `json:"meta"`
}

type JuustagramMessageListResponse struct {
	Messages []JuustagramMessage `json:"messages"`
	Meta     PaginationMeta      `json:"meta"`
}

type JuustagramMessageResponse struct {
	Message JuustagramMessage `json:"message"`
}

type JuustagramMessageStateListResponse struct {
	States []JuustagramMessageState `json:"states"`
}

type JuustagramMessageStateResponse struct {
	State JuustagramMessageState `json:"state"`
}

type JuustagramDiscussResponse struct {
	Options    []JuustagramDiscussOption    `json:"options"`
	Selections []JuustagramDiscussSelection `json:"selections"`
}

type JuustagramPlayerDiscussListResponse struct {
	Entries []JuustagramPlayerDiscussState `json:"entries"`
}

type JuustagramPlayerDiscussResponse struct {
	Entry JuustagramPlayerDiscussState `json:"entry"`
}

type JuustagramMessageUpdateRequest struct {
	Read *bool `json:"read"`
	Like *bool `json:"like"`
}

type JuustagramMessageStateUpdateRequest struct {
	IsRead    uint32 `json:"is_read"`
	IsGood    uint32 `json:"is_good"`
	GoodCount uint32 `json:"good_count"`
	UpdatedAt uint32 `json:"updated_at"`
}

type JuustagramDiscussRequest struct {
	DiscussID   uint32 `json:"discuss_id"`
	OptionIndex uint32 `json:"option_index"`
}

type JuustagramPlayerDiscussUpdateRequest struct {
	OptionIndex uint32 `json:"option_index"`
	NpcReplyID  uint32 `json:"npc_reply_id"`
	CommentTime uint32 `json:"comment_time"`
}
