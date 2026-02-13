package orm

import (
	"context"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

func ListJuustagramGroups(commanderID uint32, offset int, limit int) ([]JuustagramGroup, int64, error) {
	ctx := context.Background()
	total, err := db.DefaultStore.Queries.CountJuustagramGroupsByCommander(ctx, int64(commanderID))
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = int(total)
	}
	groupRows, err := db.DefaultStore.Queries.ListJuustagramGroupsByCommanderPaged(ctx, gen.ListJuustagramGroupsByCommanderPagedParams{
		CommanderID: int64(commanderID),
		Offset:      int32(offset),
		Limit:       int32(limit),
	})
	if err != nil {
		return nil, 0, err
	}
	groups := make([]JuustagramGroup, 0, len(groupRows))
	groupIDs := make([]int64, 0, len(groupRows))
	for _, r := range groupRows {
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
		return groups, total, nil
	}
	chatRows, err := db.DefaultStore.Queries.ListJuustagramChatGroupsByGroupRecordIDs(ctx, gen.ListJuustagramChatGroupsByGroupRecordIDsParams{CommanderID: int64(commanderID), Column2: groupIDs})
	if err != nil {
		return nil, 0, err
	}
	groupIndex := make(map[uint32]int, len(groups))
	for i := range groups {
		groupIndex[groups[i].ID] = i
	}
	chatGroupByID := make(map[uint32]*JuustagramChatGroup, len(chatRows))
	chatIDs := make([]int64, 0, len(chatRows))
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
	if len(chatIDs) > 0 {
		replyRows, err := db.DefaultStore.Queries.ListJuustagramRepliesByChatGroupRecordIDs(ctx, chatIDs)
		if err != nil {
			return nil, 0, err
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
	}
	return groups, total, nil
}

func GetJuustagramGroup(commanderID uint32, groupID uint32) (*JuustagramGroup, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetJuustagramGroupByCommanderAndGroupID(ctx, gen.GetJuustagramGroupByCommanderAndGroupIDParams{CommanderID: int64(commanderID), GroupID: int64(groupID)})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	group := JuustagramGroup{
		ID:           uint32(row.ID),
		CommanderID:  uint32(row.CommanderID),
		GroupID:      uint32(row.GroupID),
		SkinID:       uint32(row.SkinID),
		Favorite:     uint32(row.Favorite),
		CurChatGroup: uint32(row.CurChatGroup),
		ChatGroups:   []JuustagramChatGroup{},
	}
	chatRows, err := db.DefaultStore.Queries.ListJuustagramChatGroupsByGroupRecordIDs(ctx, gen.ListJuustagramChatGroupsByGroupRecordIDsParams{CommanderID: int64(commanderID), Column2: []int64{int64(group.ID)}})
	if err != nil {
		return nil, err
	}
	chatGroupByID := make(map[uint32]*JuustagramChatGroup, len(chatRows))
	chatIDs := make([]int64, 0, len(chatRows))
	for _, cr := range chatRows {
		cg := JuustagramChatGroup{
			ID:            uint32(cr.ID),
			CommanderID:   uint32(cr.CommanderID),
			GroupRecordID: uint32(cr.GroupRecordID),
			ChatGroupID:   uint32(cr.ChatGroupID),
			OpTime:        uint32(cr.OpTime),
			ReadFlag:      uint32(cr.ReadFlag),
			ReplyList:     []JuustagramReply{},
		}
		group.ChatGroups = append(group.ChatGroups, cg)
		chatGroupByID[uint32(cr.ID)] = &group.ChatGroups[len(group.ChatGroups)-1]
		chatIDs = append(chatIDs, cr.ID)
	}
	if len(chatIDs) > 0 {
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
	}
	return &group, nil
}

func UpdateJuustagramGroup(commanderID uint32, groupID uint32, skinID *uint32, favorite *uint32, curChatGroup *uint32) error {
	ctx := context.Background()
	if skinID == nil && favorite == nil && curChatGroup == nil {
		return nil
	}
	params := gen.UpdateJuustagramGroupParams{
		CommanderID:  int64(commanderID),
		GroupID:      int64(groupID),
		SkinID:       pgInt8FromUint32Ptr(skinID),
		Favorite:     pgInt8FromUint32Ptr(favorite),
		CurChatGroup: pgInt8FromUint32Ptr(curChatGroup),
	}
	tag, err := db.DefaultStore.Queries.UpdateJuustagramGroup(ctx, params)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func CreateJuustagramChatGroup(commanderID uint32, groupID uint32, chatGroupID uint32, opTime uint32) (*JuustagramChatGroup, error) {
	group, err := GetJuustagramGroup(commanderID, groupID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	id, err := db.DefaultStore.Queries.CreateJuustagramChatGroup(ctx, gen.CreateJuustagramChatGroupParams{
		CommanderID:   int64(commanderID),
		GroupRecordID: int64(group.ID),
		ChatGroupID:   int64(chatGroupID),
		OpTime:        int64(opTime),
		ReadFlag:      0,
	})
	if err != nil {
		return nil, err
	}
	entry := JuustagramChatGroup{
		ID:            uint32(id),
		CommanderID:   commanderID,
		GroupRecordID: group.ID,
		ChatGroupID:   chatGroupID,
		OpTime:        opTime,
		ReadFlag:      0,
		ReplyList:     []JuustagramReply{},
	}
	return &entry, nil
}

func GetJuustagramChatGroup(commanderID uint32, chatGroupID uint32) (*JuustagramChatGroup, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetJuustagramChatGroupByCommanderAndChatGroupID(ctx, gen.GetJuustagramChatGroupByCommanderAndChatGroupIDParams{CommanderID: int64(commanderID), ChatGroupID: int64(chatGroupID)})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	group := JuustagramChatGroup{
		ID:            uint32(row.ID),
		CommanderID:   uint32(row.CommanderID),
		GroupRecordID: uint32(row.GroupRecordID),
		ChatGroupID:   uint32(row.ChatGroupID),
		OpTime:        uint32(row.OpTime),
		ReadFlag:      uint32(row.ReadFlag),
		ReplyList:     []JuustagramReply{},
	}
	replies, err := db.DefaultStore.Queries.ListJuustagramRepliesByChatGroupRecordIDs(ctx, []int64{int64(group.ID)})
	if err != nil {
		return nil, err
	}
	for _, rr := range replies {
		group.ReplyList = append(group.ReplyList, JuustagramReply{
			ID:                uint32(rr.ID),
			ChatGroupRecordID: uint32(rr.ChatGroupRecordID),
			Sequence:          uint32(rr.Sequence),
			Key:               uint32(rr.Key),
			Value:             uint32(rr.Value),
		})
	}
	return &group, nil
}

func AddJuustagramChatReply(commanderID uint32, chatGroupID uint32, chatID uint32, value uint32, now uint32) (*JuustagramChatGroup, error) {
	ctx := context.Background()
	if err := db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		cg, err := q.GetJuustagramChatGroupByCommanderAndChatGroupID(ctx, gen.GetJuustagramChatGroupByCommanderAndChatGroupIDParams{CommanderID: int64(commanderID), ChatGroupID: int64(chatGroupID)})
		if err != nil {
			return err
		}
		maxSeq, err := q.GetMaxJuustagramReplySequence(ctx, cg.ID)
		if err != nil {
			return err
		}
		if _, err := q.CreateJuustagramReply(ctx, gen.CreateJuustagramReplyParams{ChatGroupRecordID: cg.ID, Sequence: maxSeq + 1, Key: int64(chatID), Value: int64(value)}); err != nil {
			return err
		}
		return q.UpdateJuustagramChatGroupOpTimeReadFlag(ctx, gen.UpdateJuustagramChatGroupOpTimeReadFlagParams{ID: cg.ID, CommanderID: int64(commanderID), OpTime: int64(now), ReadFlag: 0})
	}); err != nil {
		return nil, err
	}
	return GetJuustagramChatGroup(commanderID, chatGroupID)
}

func SetJuustagramChatGroupRead(commanderID uint32, chatGroupIDs []uint32) error {
	return MarkJuustagramChatGroupsRead(commanderID, chatGroupIDs)
}

func DeleteJuustagramGroup(commanderID uint32, groupID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM juustagram_groups
WHERE commander_id = $1
  AND group_id = $2
`, int64(commanderID), int64(groupID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteJuustagramChatGroup(commanderID uint32, chatGroupID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM juustagram_chat_groups
WHERE commander_id = $1
  AND chat_group_id = $2
`, int64(commanderID), int64(chatGroupID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func GetJuustagramGroupByRecordID(commanderID uint32, groupRecordID uint32) (*JuustagramGroup, error) {
	ctx := context.Background()
	var groupID int64
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT group_id
FROM juustagram_groups
WHERE id = $1
  AND commander_id = $2
`, int64(groupRecordID), int64(commanderID)).Scan(&groupID)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return GetJuustagramGroup(commanderID, uint32(groupID))
}

func GetJuustagramChatGroupRecordID(commanderID uint32, chatGroupID uint32) (uint32, error) {
	ctx := context.Background()
	var recordID int64
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id
FROM juustagram_chat_groups
WHERE commander_id = $1
  AND chat_group_id = $2
`, int64(commanderID), int64(chatGroupID)).Scan(&recordID)
	err = db.MapNotFound(err)
	if err != nil {
		return 0, err
	}
	return uint32(recordID), nil
}

func DeleteJuustagramReply(chatGroupRecordID uint32, sequence uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM juustagram_replies
WHERE chat_group_record_id = $1
  AND sequence = $2
`, int64(chatGroupRecordID), int64(sequence))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func SetJuustagramCurrentChatGroup(commanderID uint32, chatGroupID uint32) error {
	ctx := context.Background()
	chatGroup, err := db.DefaultStore.Queries.GetJuustagramChatGroupByCommanderAndChatGroupID(ctx, gen.GetJuustagramChatGroupByCommanderAndChatGroupIDParams{CommanderID: int64(commanderID), ChatGroupID: int64(chatGroupID)})
	if err != nil {
		return db.MapNotFound(err)
	}
	return db.DefaultStore.Queries.UpdateJuustagramGroupCurrentChatGroupByID(ctx, gen.UpdateJuustagramGroupCurrentChatGroupByIDParams{ID: chatGroup.GroupRecordID, CommanderID: int64(commanderID), CurChatGroup: int64(chatGroupID)})
}

func EnsureJuustagramGroupExists(commanderID uint32, groupID uint32, chatGroupID uint32) (*JuustagramGroup, error) {
	group, err := GetJuustagramGroup(commanderID, groupID)
	if err == nil {
		return group, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}
	created, err := CreateJuustagramGroup(commanderID, groupID, chatGroupID)
	if err != nil {
		return nil, err
	}
	return GetJuustagramGroup(commanderID, created.GroupID)
}

func ValidateJuustagramChatGroupIDs(groupIDs []uint32) error {
	if len(groupIDs) == 0 {
		return errors.New("chat_group_ids is required")
	}
	for _, id := range groupIDs {
		if id == 0 {
			return errors.New("chat_group_id must be > 0")
		}
	}
	return nil
}

func DefaultJuustagramOpTime() uint32 {
	return uint32(time.Now().Unix())
}
