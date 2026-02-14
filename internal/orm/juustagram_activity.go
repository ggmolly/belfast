package orm

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

func CreateJuustagramTemplate(template *JuustagramTemplate) error {
	ctx := context.Background()
	npcDiscussPersist, err := juustagramJSONString(template.NpcDiscussPersist)
	if err != nil {
		return err
	}
	timeValue, err := juustagramJSONString(template.Time)
	if err != nil {
		return err
	}
	timePersist, err := juustagramJSONString(template.TimePersist)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO juustagram_templates (
	id,
	group_id,
	ship_group,
	name,
	sculpture,
	picture_persist,
	message_persist,
	is_active,
	npc_discuss_persist,
	time,
	time_persist
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`,
		int64(template.ID),
		int64(template.GroupID),
		int64(template.ShipGroup),
		template.Name,
		template.Sculpture,
		template.PicturePersist,
		template.MessagePersist,
		int64(template.IsActive),
		npcDiscussPersist,
		timeValue,
		timePersist,
	)
	return err
}

func UpdateJuustagramTemplate(template *JuustagramTemplate) error {
	ctx := context.Background()
	npcDiscussPersist, err := juustagramJSONString(template.NpcDiscussPersist)
	if err != nil {
		return err
	}
	timeValue, err := juustagramJSONString(template.Time)
	if err != nil {
		return err
	}
	timePersist, err := juustagramJSONString(template.TimePersist)
	if err != nil {
		return err
	}
	return db.DefaultStore.Queries.UpsertJuustagramTemplate(ctx, gen.UpsertJuustagramTemplateParams{
		ID:                int64(template.ID),
		GroupID:           int64(template.GroupID),
		ShipGroup:         int64(template.ShipGroup),
		Name:              template.Name,
		Sculpture:         template.Sculpture,
		PicturePersist:    template.PicturePersist,
		MessagePersist:    template.MessagePersist,
		IsActive:          int64(template.IsActive),
		NpcDiscussPersist: npcDiscussPersist,
		Time:              timeValue,
		TimePersist:       timePersist,
	})
}

func DeleteJuustagramTemplate(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM juustagram_templates WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func CreateJuustagramNpcTemplate(template *JuustagramNpcTemplate) error {
	ctx := context.Background()
	npcReplyPersist, err := juustagramJSONString(template.NpcReplyPersist)
	if err != nil {
		return err
	}
	timePersist, err := juustagramJSONString(template.TimePersist)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO juustagram_npc_templates (
	id,
	ship_group,
	message_persist,
	npc_reply_persist,
	time_persist
)
VALUES ($1, $2, $3, $4, $5)
`,
		int64(template.ID),
		int64(template.ShipGroup),
		template.MessagePersist,
		npcReplyPersist,
		timePersist,
	)
	return err
}

func UpdateJuustagramNpcTemplate(template *JuustagramNpcTemplate) error {
	ctx := context.Background()
	npcReplyPersist, err := juustagramJSONString(template.NpcReplyPersist)
	if err != nil {
		return err
	}
	timePersist, err := juustagramJSONString(template.TimePersist)
	if err != nil {
		return err
	}
	return db.DefaultStore.Queries.UpsertJuustagramNpcTemplate(ctx, gen.UpsertJuustagramNpcTemplateParams{
		ID:              int64(template.ID),
		ShipGroup:       int64(template.ShipGroup),
		MessagePersist:  template.MessagePersist,
		NpcReplyPersist: npcReplyPersist,
		TimePersist:     timePersist,
	})
}

func DeleteJuustagramNpcTemplate(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM juustagram_npc_templates WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func CreateJuustagramShipGroupTemplate(template *JuustagramShipGroupTemplate) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO juustagram_ship_group_templates (
	ship_group,
	name,
	background,
	sculpture,
	sculpture_ii,
	nationality,
	type
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`,
		int64(template.ShipGroup),
		template.Name,
		template.Background,
		template.Sculpture,
		template.SculptureII,
		int64(template.Nationality),
		int64(template.Type),
	)
	return err
}

func UpdateJuustagramShipGroupTemplate(template *JuustagramShipGroupTemplate) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertJuustagramShipGroupTemplate(ctx, gen.UpsertJuustagramShipGroupTemplateParams{
		ShipGroup:   int64(template.ShipGroup),
		Name:        template.Name,
		Background:  template.Background,
		Sculpture:   template.Sculpture,
		SculptureIi: template.SculptureII,
		Nationality: int64(template.Nationality),
		Type:        int64(template.Type),
	})
}

func DeleteJuustagramShipGroupTemplate(shipGroup uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM juustagram_ship_group_templates WHERE ship_group = $1`, int64(shipGroup))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func CreateJuustagramLanguage(entry *JuustagramLanguage) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO juustagram_languages (key, value)
VALUES ($1, $2)
`, entry.Key, entry.Value)
	return err
}

func UpdateJuustagramLanguage(entry *JuustagramLanguage) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertJuustagramLanguage(ctx, gen.UpsertJuustagramLanguageParams{
		Key:   entry.Key,
		Value: entry.Value,
	})
}

func DeleteJuustagramLanguage(key string) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM juustagram_languages WHERE key = $1`, key)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

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
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetJuustagramTemplate(ctx, int64(id))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	template := JuustagramTemplate{
		ID:             uint32(row.ID),
		GroupID:        uint32(row.GroupID),
		ShipGroup:      uint32(row.ShipGroup),
		Name:           row.Name,
		Sculpture:      row.Sculpture,
		PicturePersist: row.PicturePersist,
		MessagePersist: row.MessagePersist,
		IsActive:       uint32(row.IsActive),
	}
	if err := template.NpcDiscussPersist.Scan(row.NpcDiscussPersist); err != nil {
		return nil, err
	}
	if err := template.Time.Scan(row.Time); err != nil {
		return nil, err
	}
	if err := template.TimePersist.Scan(row.TimePersist); err != nil {
		return nil, err
	}
	return &template, nil
}

func ListJuustagramTemplates(offset int, limit int) ([]JuustagramTemplate, int64, error) {
	ctx := context.Background()
	total, err := db.DefaultStore.Queries.CountJuustagramTemplates(ctx)
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = int(total)
	}
	rows, err := db.DefaultStore.Queries.ListJuustagramTemplates(ctx, gen.ListJuustagramTemplatesParams{Offset: int32(offset), Limit: int32(limit)})
	if err != nil {
		return nil, 0, err
	}
	templates := make([]JuustagramTemplate, 0, len(rows))
	for _, row := range rows {
		t := JuustagramTemplate{
			ID:             uint32(row.ID),
			GroupID:        uint32(row.GroupID),
			ShipGroup:      uint32(row.ShipGroup),
			Name:           row.Name,
			Sculpture:      row.Sculpture,
			PicturePersist: row.PicturePersist,
			MessagePersist: row.MessagePersist,
			IsActive:       uint32(row.IsActive),
		}
		if err := t.NpcDiscussPersist.Scan(row.NpcDiscussPersist); err != nil {
			return nil, 0, err
		}
		if err := t.Time.Scan(row.Time); err != nil {
			return nil, 0, err
		}
		if err := t.TimePersist.Scan(row.TimePersist); err != nil {
			return nil, 0, err
		}
		templates = append(templates, t)
	}
	return templates, total, nil
}

func GetJuustagramNpcTemplate(id uint32) (*JuustagramNpcTemplate, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetJuustagramNpcTemplate(ctx, int64(id))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	template := JuustagramNpcTemplate{
		ID:             uint32(row.ID),
		ShipGroup:      uint32(row.ShipGroup),
		MessagePersist: row.MessagePersist,
	}
	if err := template.NpcReplyPersist.Scan(row.NpcReplyPersist); err != nil {
		return nil, err
	}
	if err := template.TimePersist.Scan(row.TimePersist); err != nil {
		return nil, err
	}
	return &template, nil
}

func ListJuustagramNpcTemplates(offset int, limit int) ([]JuustagramNpcTemplate, int64, error) {
	ctx := context.Background()
	total, err := db.DefaultStore.Queries.CountJuustagramNpcTemplates(ctx)
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = int(total)
	}
	rows, err := db.DefaultStore.Queries.ListJuustagramNpcTemplates(ctx, gen.ListJuustagramNpcTemplatesParams{Offset: int32(offset), Limit: int32(limit)})
	if err != nil {
		return nil, 0, err
	}
	templates := make([]JuustagramNpcTemplate, 0, len(rows))
	for _, row := range rows {
		t := JuustagramNpcTemplate{
			ID:             uint32(row.ID),
			ShipGroup:      uint32(row.ShipGroup),
			MessagePersist: row.MessagePersist,
		}
		if err := t.NpcReplyPersist.Scan(row.NpcReplyPersist); err != nil {
			return nil, 0, err
		}
		if err := t.TimePersist.Scan(row.TimePersist); err != nil {
			return nil, 0, err
		}
		templates = append(templates, t)
	}
	return templates, total, nil
}

func GetJuustagramShipGroupTemplate(shipGroup uint32) (*JuustagramShipGroupTemplate, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetJuustagramShipGroupTemplate(ctx, int64(shipGroup))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	template := JuustagramShipGroupTemplate{
		ShipGroup:   uint32(row.ShipGroup),
		Name:        row.Name,
		Background:  row.Background,
		Sculpture:   row.Sculpture,
		SculptureII: row.SculptureIi,
		Nationality: uint32(row.Nationality),
		Type:        uint32(row.Type),
	}
	return &template, nil
}

func ListJuustagramShipGroupTemplates(offset int, limit int) ([]JuustagramShipGroupTemplate, int64, error) {
	ctx := context.Background()
	total, err := db.DefaultStore.Queries.CountJuustagramShipGroupTemplates(ctx)
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = int(total)
	}
	rows, err := db.DefaultStore.Queries.ListJuustagramShipGroupTemplates(ctx, gen.ListJuustagramShipGroupTemplatesParams{Offset: int32(offset), Limit: int32(limit)})
	if err != nil {
		return nil, 0, err
	}
	templates := make([]JuustagramShipGroupTemplate, 0, len(rows))
	for _, row := range rows {
		templates = append(templates, JuustagramShipGroupTemplate{
			ShipGroup:   uint32(row.ShipGroup),
			Name:        row.Name,
			Background:  row.Background,
			Sculpture:   row.Sculpture,
			SculptureII: row.SculptureIi,
			Nationality: uint32(row.Nationality),
			Type:        uint32(row.Type),
		})
	}
	return templates, total, nil
}

func GetJuustagramLanguage(key string) (string, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetJuustagramLanguage(ctx, key)
	err = db.MapNotFound(err)
	if err != nil {
		return "", err
	}
	return row.Value, nil
}

func ListJuustagramLanguageByPrefix(prefix string) ([]JuustagramLanguage, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListJuustagramLanguageByPrefix(ctx, prefix+"%")
	if err != nil {
		return nil, err
	}
	entries := make([]JuustagramLanguage, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, JuustagramLanguage{Key: row.Key, Value: row.Value})
	}
	return entries, nil
}

func ListJuustagramOpReplies(messageID uint32) ([]JuustagramNpcTemplate, error) {
	prefix := fmt.Sprintf("op_reply_%d_", messageID)
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListJuustagramOpReplies(ctx, prefix+"%")
	if err != nil {
		return nil, err
	}
	replies := make([]JuustagramNpcTemplate, 0, len(rows))
	for _, row := range rows {
		r := JuustagramNpcTemplate{
			ID:             uint32(row.ID),
			ShipGroup:      uint32(row.ShipGroup),
			MessagePersist: row.MessagePersist,
		}
		if err := r.NpcReplyPersist.Scan(row.NpcReplyPersist); err != nil {
			return nil, err
		}
		if err := r.TimePersist.Scan(row.TimePersist); err != nil {
			return nil, err
		}
		replies = append(replies, r)
	}
	return replies, nil
}

func GetJuustagramMessageState(commanderID uint32, messageID uint32) (*JuustagramMessageState, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetJuustagramMessageState(ctx, gen.GetJuustagramMessageStateParams{CommanderID: int64(commanderID), MessageID: int64(messageID)})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	state := JuustagramMessageState{
		ID:          uint32(row.ID),
		CommanderID: uint32(row.CommanderID),
		MessageID:   uint32(row.MessageID),
		IsRead:      uint32(row.IsRead),
		IsGood:      uint32(row.IsGood),
		GoodCount:   uint32(row.GoodCount),
		UpdatedAt:   uint32(row.UpdatedAt),
	}
	return &state, nil
}

func GetOrCreateJuustagramMessageState(commanderID uint32, messageID uint32, now uint32) (*JuustagramMessageState, error) {
	state, err := GetJuustagramMessageState(commanderID, messageID)
	if err == nil {
		return state, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
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
	ctx := context.Background()
	id, err := db.DefaultStore.Queries.CreateJuustagramMessageState(ctx, gen.CreateJuustagramMessageStateParams{
		CommanderID: int64(commanderID),
		MessageID:   int64(messageID),
		IsRead:      int64(created.IsRead),
		IsGood:      int64(created.IsGood),
		GoodCount:   int64(created.GoodCount),
		UpdatedAt:   int64(created.UpdatedAt),
	})
	if err != nil {
		return nil, err
	}
	created.ID = uint32(id)
	return &created, nil
}

func SaveJuustagramMessageState(state *JuustagramMessageState) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpdateJuustagramMessageState(ctx, gen.UpdateJuustagramMessageStateParams{
		CommanderID: int64(state.CommanderID),
		MessageID:   int64(state.MessageID),
		IsRead:      int64(state.IsRead),
		IsGood:      int64(state.IsGood),
		GoodCount:   int64(state.GoodCount),
		UpdatedAt:   int64(state.UpdatedAt),
	})
}

func ListJuustagramMessageStatesByCommander(commanderID uint32) ([]JuustagramMessageState, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, commander_id, message_id, is_read, is_good, good_count, updated_at
FROM juustagram_message_states
WHERE commander_id = $1
ORDER BY message_id ASC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make([]JuustagramMessageState, 0)
	for rows.Next() {
		var (
			state       JuustagramMessageState
			id          int64
			commanderID int64
			messageID   int64
			isRead      int64
			isGood      int64
			goodCount   int64
			updatedAt   int64
		)
		if err := rows.Scan(&id, &commanderID, &messageID, &isRead, &isGood, &goodCount, &updatedAt); err != nil {
			return nil, err
		}
		state.ID = uint32(id)
		state.CommanderID = uint32(commanderID)
		state.MessageID = uint32(messageID)
		state.IsRead = uint32(isRead)
		state.IsGood = uint32(isGood)
		state.GoodCount = uint32(goodCount)
		state.UpdatedAt = uint32(updatedAt)
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return states, nil
}

func DeleteJuustagramMessageState(commanderID uint32, messageID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM juustagram_message_states
WHERE commander_id = $1
  AND message_id = $2
`, int64(commanderID), int64(messageID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func GetJuustagramPlayerDiscuss(commanderID uint32, messageID uint32, discussID uint32) (*JuustagramPlayerDiscuss, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetJuustagramPlayerDiscuss(ctx, gen.GetJuustagramPlayerDiscussParams{CommanderID: int64(commanderID), MessageID: int64(messageID), DiscussID: int64(discussID)})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	entry := JuustagramPlayerDiscuss{
		ID:          uint32(row.ID),
		CommanderID: uint32(row.CommanderID),
		MessageID:   uint32(row.MessageID),
		DiscussID:   uint32(row.DiscussID),
		OptionIndex: uint32(row.OptionIndex),
		NpcReplyID:  uint32(row.NpcReplyID),
		CommentTime: uint32(row.CommentTime),
	}
	return &entry, nil
}

func ListJuustagramPlayerDiscuss(commanderID uint32, messageID uint32) ([]JuustagramPlayerDiscuss, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListJuustagramPlayerDiscuss(ctx, gen.ListJuustagramPlayerDiscussParams{CommanderID: int64(commanderID), MessageID: int64(messageID)})
	if err != nil {
		return nil, err
	}
	entries := make([]JuustagramPlayerDiscuss, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, JuustagramPlayerDiscuss{
			ID:          uint32(row.ID),
			CommanderID: uint32(row.CommanderID),
			MessageID:   uint32(row.MessageID),
			DiscussID:   uint32(row.DiscussID),
			OptionIndex: uint32(row.OptionIndex),
			NpcReplyID:  uint32(row.NpcReplyID),
			CommentTime: uint32(row.CommentTime),
		})
	}
	return entries, nil
}

func UpsertJuustagramPlayerDiscuss(entry *JuustagramPlayerDiscuss) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertJuustagramPlayerDiscuss(ctx, gen.UpsertJuustagramPlayerDiscussParams{
		CommanderID: int64(entry.CommanderID),
		MessageID:   int64(entry.MessageID),
		DiscussID:   int64(entry.DiscussID),
		OptionIndex: int64(entry.OptionIndex),
		NpcReplyID:  int64(entry.NpcReplyID),
		CommentTime: int64(entry.CommentTime),
	})
}

func DeleteJuustagramPlayerDiscuss(commanderID uint32, messageID uint32, discussID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM juustagram_player_discusses
WHERE commander_id = $1
  AND message_id = $2
  AND discuss_id = $3
`, int64(commanderID), int64(messageID), int64(discussID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func juustagramJSONString(value any) (string, error) {
	encoded, err := marshalJuustagramJSON(value)
	if err != nil {
		return "", err
	}
	payload, ok := encoded.(string)
	if !ok {
		return "", fmt.Errorf("unexpected juustagram payload type %T", encoded)
	}
	return payload, nil
}
