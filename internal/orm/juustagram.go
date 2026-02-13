package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type JuustagramGroup struct {
	ID           uint32 `gorm:"primary_key"`
	CommanderID  uint32 `gorm:"not_null;index:idx_juus_group_commander,unique"`
	GroupID      uint32 `gorm:"not_null;index:idx_juus_group_commander,unique"`
	SkinID       uint32 `gorm:"not_null;default:0"`
	Favorite     uint32 `gorm:"not_null;default:0"`
	CurChatGroup uint32 `gorm:"not_null;default:0"`

	ChatGroups []JuustagramChatGroup `gorm:"foreignKey:GroupRecordID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type JuustagramChatGroup struct {
	ID            uint32 `gorm:"primary_key"`
	CommanderID   uint32 `gorm:"not_null;index:idx_juus_chat_group_commander,unique"`
	GroupRecordID uint32 `gorm:"not_null;index"`
	ChatGroupID   uint32 `gorm:"not_null;index:idx_juus_chat_group_commander,unique"`
	OpTime        uint32 `gorm:"not_null;default:0"`
	ReadFlag      uint32 `gorm:"not_null;default:0"`

	ReplyList []JuustagramReply `gorm:"foreignKey:ChatGroupRecordID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type JuustagramReply struct {
	ID                uint32 `gorm:"primary_key"`
	ChatGroupRecordID uint32 `gorm:"not_null;index:idx_juus_reply_order,unique"`
	Sequence          uint32 `gorm:"not_null;index:idx_juus_reply_order,unique"`
	Key               uint32 `gorm:"not_null"`
	Value             uint32 `gorm:"not_null"`
}

func GetJuustagramGroups(commanderID uint32) ([]JuustagramGroup, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListJuustagramGroups(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	groups := make([]JuustagramGroup, 0, len(rows))
	groupIDs := make([]int64, 0, len(rows))
	for _, r := range rows {
		groups = append(groups, JuustagramGroup{
			ID:           uint32(r.ID),
			CommanderID:  uint32(r.CommanderID),
			GroupID:      uint32(r.GroupID),
			SkinID:       uint32(r.SkinID),
			Favorite:     uint32(r.Favorite),
			CurChatGroup: uint32(r.CurChatGroup),
			ChatGroups:   []JuustagramChatGroup{},
		})
		groupIDs = append(groupIDs, r.ID)
	}
	if len(groupIDs) == 0 {
		return groups, nil
	}
	chatRows, err := db.DefaultStore.Queries.ListJuustagramChatGroupsByGroupRecordIDs(ctx, gen.ListJuustagramChatGroupsByGroupRecordIDsParams{CommanderID: int64(commanderID), Column2: groupIDs})
	if err != nil {
		return nil, err
	}
	chatIDs := make([]int64, 0, len(chatRows))
	groupIndex := make(map[uint32]int, len(groups))
	for i := range groups {
		groupIndex[groups[i].ID] = i
	}
	chatGroupByID := make(map[uint32]*JuustagramChatGroup, len(chatRows))
	for _, cr := range chatRows {
		gidx, ok := groupIndex[uint32(cr.GroupRecordID)]
		if !ok {
			continue
		}
		cg := JuustagramChatGroup{
			ID:            uint32(cr.ID),
			CommanderID:   uint32(cr.CommanderID),
			GroupRecordID: uint32(cr.GroupRecordID),
			ChatGroupID:   uint32(cr.ChatGroupID),
			OpTime:        uint32(cr.OpTime),
			ReadFlag:      uint32(cr.ReadFlag),
			ReplyList:     []JuustagramReply{},
		}
		groups[gidx].ChatGroups = append(groups[gidx].ChatGroups, cg)
		chatGroupByID[uint32(cr.ID)] = &groups[gidx].ChatGroups[len(groups[gidx].ChatGroups)-1]
		chatIDs = append(chatIDs, cr.ID)
	}
	if len(chatIDs) == 0 {
		return groups, nil
	}
	replyRows, err := db.DefaultStore.Queries.ListJuustagramRepliesByChatGroupRecordIDs(ctx, chatIDs)
	if err != nil {
		return nil, err
	}
	for _, rr := range replyRows {
		cg := chatGroupByID[uint32(rr.ChatGroupRecordID)]
		if cg == nil {
			continue
		}
		cg.ReplyList = append(cg.ReplyList, JuustagramReply{
			ID:                uint32(rr.ID),
			ChatGroupRecordID: uint32(rr.ChatGroupRecordID),
			Sequence:          uint32(rr.Sequence),
			Key:               uint32(rr.Key),
			Value:             uint32(rr.Value),
		})
	}
	return groups, nil
}

func CreateJuustagramGroup(commanderID uint32, groupID uint32, chatGroupID uint32) (*JuustagramGroup, error) {
	ctx := context.Background()
	var out *JuustagramGroup
	err := db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		groupRowID, err := q.CreateJuustagramGroup(ctx, gen.CreateJuustagramGroupParams{
			CommanderID:  int64(commanderID),
			GroupID:      int64(groupID),
			SkinID:       0,
			Favorite:     0,
			CurChatGroup: int64(chatGroupID),
		})
		if err != nil {
			return err
		}
		chatRowID, err := q.CreateJuustagramChatGroup(ctx, gen.CreateJuustagramChatGroupParams{
			CommanderID:   int64(commanderID),
			GroupRecordID: groupRowID,
			ChatGroupID:   int64(chatGroupID),
			OpTime:        0,
			ReadFlag:      0,
		})
		if err != nil {
			return err
		}
		out = &JuustagramGroup{
			ID:           uint32(groupRowID),
			CommanderID:  commanderID,
			GroupID:      groupID,
			SkinID:       0,
			Favorite:     0,
			CurChatGroup: chatGroupID,
			ChatGroups: []JuustagramChatGroup{{
				ID:            uint32(chatRowID),
				CommanderID:   commanderID,
				GroupRecordID: uint32(groupRowID),
				ChatGroupID:   chatGroupID,
				OpTime:        0,
				ReadFlag:      0,
				ReplyList:     []JuustagramReply{},
			}},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func MarkJuustagramChatGroupsRead(commanderID uint32, chatGroupIDs []uint32) error {
	ctx := context.Background()
	if len(chatGroupIDs) == 0 {
		return db.DefaultStore.Queries.MarkAllJuustagramChatGroupsRead(ctx, int64(commanderID))
	}
	ids := make([]int64, 0, len(chatGroupIDs))
	for _, id := range chatGroupIDs {
		ids = append(ids, int64(id))
	}
	return db.DefaultStore.Queries.MarkJuustagramChatGroupsRead(ctx, gen.MarkJuustagramChatGroupsReadParams{CommanderID: int64(commanderID), Column2: ids})
}
