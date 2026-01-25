package orm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type JuustagramTemplate struct {
	ID                uint32               `gorm:"primary_key" json:"id"`
	GroupID           uint32               `gorm:"not_null;index" json:"group_id"`
	ShipGroup         uint32               `gorm:"not_null;index" json:"ship_group"`
	Name              string               `gorm:"not_null" json:"name"`
	Sculpture         string               `gorm:"not_null" json:"sculpture"`
	PicturePersist    string               `gorm:"not_null" json:"picture_persist"`
	MessagePersist    string               `gorm:"not_null" json:"message_persist"`
	IsActive          uint32               `gorm:"not_null;default:0" json:"is_active"`
	NpcDiscussPersist JuustagramUint32List `gorm:"type:text;not_null;default:'[]'" json:"npc_discuss_persist"`
	Time              JuustagramTimeConfig `gorm:"type:text;not_null;default:'[]'" json:"time"`
	TimePersist       JuustagramTimeConfig `gorm:"type:text;not_null;default:'[]'" json:"time_persist"`
}

type JuustagramNpcTemplate struct {
	ID              uint32               `gorm:"primary_key" json:"id"`
	ShipGroup       uint32               `gorm:"not_null;index" json:"ship_group"`
	MessagePersist  string               `gorm:"not_null" json:"message_persist"`
	NpcReplyPersist JuustagramReplyList  `gorm:"type:text;not_null;default:'[]'" json:"npc_reply_persist"`
	TimePersist     JuustagramTimeConfig `gorm:"type:text;not_null;default:'[]'" json:"time_persist"`
}

type JuustagramLanguage struct {
	Key   string `gorm:"primary_key" json:"key"`
	Value string `gorm:"not_null" json:"value"`
}

type JuustagramShipGroupTemplate struct {
	ShipGroup   uint32 `gorm:"primary_key" json:"ship_group"`
	Name        string `gorm:"not_null" json:"name"`
	Background  string `gorm:"not_null" json:"background"`
	Sculpture   string `gorm:"not_null" json:"sculpture"`
	SculptureII string `gorm:"not_null" json:"sculpture_ii"`
	Nationality uint32 `gorm:"not_null" json:"nationality"`
	Type        uint32 `gorm:"not_null" json:"type"`
}

type JuustagramMessageState struct {
	ID          uint32 `gorm:"primary_key" json:"id"`
	CommanderID uint32 `gorm:"not_null;index:idx_juus_message_state,unique" json:"commander_id"`
	MessageID   uint32 `gorm:"not_null;index:idx_juus_message_state,unique" json:"message_id"`
	IsRead      uint32 `gorm:"not_null;default:0" json:"is_read"`
	IsGood      uint32 `gorm:"not_null;default:0" json:"is_good"`
	GoodCount   uint32 `gorm:"not_null;default:0" json:"good"`
	UpdatedAt   uint32 `gorm:"not_null;default:0" json:"updated_at"`
}

type JuustagramPlayerDiscuss struct {
	ID          uint32 `gorm:"primary_key" json:"id"`
	CommanderID uint32 `gorm:"not_null;index:idx_juus_discuss_state,unique" json:"commander_id"`
	MessageID   uint32 `gorm:"not_null;index:idx_juus_discuss_state,unique" json:"message_id"`
	DiscussID   uint32 `gorm:"not_null;index:idx_juus_discuss_state,unique" json:"discuss_id"`
	OptionIndex uint32 `gorm:"not_null" json:"option_index"`
	NpcReplyID  uint32 `gorm:"not_null;default:0" json:"npc_reply_id"`
	CommentTime uint32 `gorm:"not_null;default:0" json:"comment_time"`
}

type JuustagramUint32List []uint32

type JuustagramReplyList []uint32

type JuustagramTimeConfig [][]int

func (list JuustagramUint32List) Value() (driver.Value, error) {
	return marshalJuustagramJSON(list)
}

func (list *JuustagramUint32List) Scan(value any) error {
	return scanJuustagramJSON(value, list)
}

func (list *JuustagramReplyList) UnmarshalJSON(data []byte) error {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 || string(data) == "\"\"" {
		*list = JuustagramReplyList{}
		return nil
	}
	var parsed []uint32
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*list = JuustagramReplyList(parsed)
	return nil
}

func (list JuustagramReplyList) Value() (driver.Value, error) {
	return marshalJuustagramJSON(list)
}

func (list *JuustagramReplyList) Scan(value any) error {
	return scanJuustagramJSON(value, list)
}

func (config JuustagramTimeConfig) Value() (driver.Value, error) {
	return marshalJuustagramJSON(config)
}

func (config *JuustagramTimeConfig) Scan(value any) error {
	return scanJuustagramJSON(value, config)
}

func marshalJuustagramJSON(value any) (driver.Value, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(payload), nil
}

func scanJuustagramJSON(value any, target any) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), target)
	case []byte:
		return json.Unmarshal(v, target)
	default:
		return fmt.Errorf("unsupported Juustagram type: %T", value)
	}
}

func GetJuustagramTemplate(id uint32) (*JuustagramTemplate, error) {
	var template JuustagramTemplate
	if err := GormDB.First(&template, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func ListJuustagramTemplates(offset int, limit int) ([]JuustagramTemplate, int64, error) {
	var total int64
	if err := GormDB.Model(&JuustagramTemplate{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var templates []JuustagramTemplate
	query := ApplyPagination(GormDB.Order("id asc"), offset, limit)
	if err := query.Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func GetJuustagramNpcTemplate(id uint32) (*JuustagramNpcTemplate, error) {
	var template JuustagramNpcTemplate
	if err := GormDB.First(&template, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func ListJuustagramNpcTemplates(offset int, limit int) ([]JuustagramNpcTemplate, int64, error) {
	var total int64
	if err := GormDB.Model(&JuustagramNpcTemplate{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var templates []JuustagramNpcTemplate
	query := ApplyPagination(GormDB.Order("id asc"), offset, limit)
	if err := query.Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func GetJuustagramShipGroupTemplate(shipGroup uint32) (*JuustagramShipGroupTemplate, error) {
	var template JuustagramShipGroupTemplate
	if err := GormDB.First(&template, "ship_group = ?", shipGroup).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func ListJuustagramShipGroupTemplates(offset int, limit int) ([]JuustagramShipGroupTemplate, int64, error) {
	var total int64
	if err := GormDB.Model(&JuustagramShipGroupTemplate{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var templates []JuustagramShipGroupTemplate
	query := ApplyPagination(GormDB.Order("ship_group asc"), offset, limit)
	if err := query.Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func GetJuustagramLanguage(key string) (string, error) {
	var entry JuustagramLanguage
	if err := GormDB.First(&entry, "key = ?", key).Error; err != nil {
		return "", err
	}
	return entry.Value, nil
}

func ListJuustagramLanguageByPrefix(prefix string) ([]JuustagramLanguage, error) {
	var entries []JuustagramLanguage
	if err := GormDB.Where("key LIKE ?", prefix+"%").Order("key asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func ListJuustagramOpReplies(messageID uint32) ([]JuustagramNpcTemplate, error) {
	prefix := fmt.Sprintf("op_reply_%d_", messageID)
	var replies []JuustagramNpcTemplate
	if err := GormDB.Where("message_persist LIKE ?", prefix+"%").Order("id asc").Find(&replies).Error; err != nil {
		return nil, err
	}
	return replies, nil
}

func GetJuustagramMessageState(commanderID uint32, messageID uint32) (*JuustagramMessageState, error) {
	var state JuustagramMessageState
	if err := GormDB.First(&state, "commander_id = ? AND message_id = ?", commanderID, messageID).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func GetOrCreateJuustagramMessageState(commanderID uint32, messageID uint32, now uint32) (*JuustagramMessageState, error) {
	state, err := GetJuustagramMessageState(commanderID, messageID)
	if err == nil {
		return state, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	created := JuustagramMessageState{
		CommanderID: commanderID,
		MessageID:   messageID,
		IsRead:      0,
		IsGood:      0,
		GoodCount:   0,
		UpdatedAt:   now,
	}
	if err := GormDB.Create(&created).Error; err != nil {
		return nil, err
	}
	return &created, nil
}

func SaveJuustagramMessageState(state *JuustagramMessageState) error {
	return GormDB.Save(state).Error
}

func GetJuustagramPlayerDiscuss(commanderID uint32, messageID uint32, discussID uint32) (*JuustagramPlayerDiscuss, error) {
	var entry JuustagramPlayerDiscuss
	if err := GormDB.First(&entry, "commander_id = ? AND message_id = ? AND discuss_id = ?", commanderID, messageID, discussID).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func ListJuustagramPlayerDiscuss(commanderID uint32, messageID uint32) ([]JuustagramPlayerDiscuss, error) {
	var entries []JuustagramPlayerDiscuss
	if err := GormDB.Where("commander_id = ? AND message_id = ?", commanderID, messageID).Order("discuss_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func UpsertJuustagramPlayerDiscuss(entry *JuustagramPlayerDiscuss) error {
	return GormDB.Save(entry).Error
}
