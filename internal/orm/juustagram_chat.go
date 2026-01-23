package orm

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

func ListJuustagramGroups(commanderID uint32, offset int, limit int) ([]JuustagramGroup, int64, error) {
	var total int64
	if err := GormDB.Model(&JuustagramGroup{}).Where("commander_id = ?", commanderID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var groups []JuustagramGroup
	if err := GormDB.
		Where("commander_id = ?", commanderID).
		Order("group_id asc").
		Limit(limit).
		Offset(offset).
		Preload("ChatGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("chat_group_id asc")
		}).
		Preload("ChatGroups.ReplyList", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence asc")
		}).
		Find(&groups).Error; err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func GetJuustagramGroup(commanderID uint32, groupID uint32) (*JuustagramGroup, error) {
	var group JuustagramGroup
	if err := GormDB.
		Where("commander_id = ? AND group_id = ?", commanderID, groupID).
		Preload("ChatGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("chat_group_id asc")
		}).
		Preload("ChatGroups.ReplyList", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence asc")
		}).
		First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func UpdateJuustagramGroup(commanderID uint32, groupID uint32, skinID *uint32, favorite *uint32, curChatGroup *uint32) error {
	updates := make(map[string]any)
	if skinID != nil {
		updates["skin_id"] = *skinID
	}
	if favorite != nil {
		updates["favorite"] = *favorite
	}
	if curChatGroup != nil {
		updates["cur_chat_group"] = *curChatGroup
	}
	if len(updates) == 0 {
		return nil
	}
	result := GormDB.Model(&JuustagramGroup{}).
		Where("commander_id = ? AND group_id = ?", commanderID, groupID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func CreateJuustagramChatGroup(commanderID uint32, groupID uint32, chatGroupID uint32, opTime uint32) (*JuustagramChatGroup, error) {
	group, err := GetJuustagramGroup(commanderID, groupID)
	if err != nil {
		return nil, err
	}
	entry := JuustagramChatGroup{
		CommanderID:   commanderID,
		GroupRecordID: group.ID,
		ChatGroupID:   chatGroupID,
		OpTime:        opTime,
		ReadFlag:      0,
	}
	if err := GormDB.Create(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func GetJuustagramChatGroup(commanderID uint32, chatGroupID uint32) (*JuustagramChatGroup, error) {
	var group JuustagramChatGroup
	if err := GormDB.
		Where("commander_id = ? AND chat_group_id = ?", commanderID, chatGroupID).
		Preload("ReplyList", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence asc")
		}).
		First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func AddJuustagramChatReply(commanderID uint32, chatGroupID uint32, chatID uint32, value uint32, now uint32) (*JuustagramChatGroup, error) {
	var group JuustagramChatGroup
	if err := GormDB.
		Where("commander_id = ? AND chat_group_id = ?", commanderID, chatGroupID).
		First(&group).Error; err != nil {
		return nil, err
	}
	var maxSequence uint32
	if err := GormDB.Model(&JuustagramReply{}).
		Where("chat_group_record_id = ?", group.ID).
		Select("coalesce(max(sequence), 0)").
		Scan(&maxSequence).Error; err != nil {
		return nil, err
	}
	entry := JuustagramReply{
		ChatGroupRecordID: group.ID,
		Sequence:          maxSequence + 1,
		Key:               chatID,
		Value:             value,
	}
	if err := GormDB.Create(&entry).Error; err != nil {
		return nil, err
	}
	if err := GormDB.Model(&JuustagramChatGroup{}).
		Where("id = ?", group.ID).
		Updates(map[string]any{"op_time": now, "read_flag": 0}).Error; err != nil {
		return nil, err
	}
	return GetJuustagramChatGroup(commanderID, chatGroupID)
}

func SetJuustagramChatGroupRead(commanderID uint32, chatGroupIDs []uint32) error {
	return MarkJuustagramChatGroupsRead(commanderID, chatGroupIDs)
}

func EnsureJuustagramGroupExists(commanderID uint32, groupID uint32, chatGroupID uint32) (*JuustagramGroup, error) {
	group, err := GetJuustagramGroup(commanderID, groupID)
	if err == nil {
		return group, nil
	}
	if err != gorm.ErrRecordNotFound {
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
