package orm

import (
	"gorm.io/gorm"
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
	var groups []JuustagramGroup
	err := GormDB.
		Where("commander_id = ?", commanderID).
		Preload("ChatGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("chat_group_id ASC")
		}).
		Preload("ChatGroups.ReplyList", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Find(&groups).
		Error
	return groups, err
}

func CreateJuustagramGroup(commanderID uint32, groupID uint32, chatGroupID uint32) (*JuustagramGroup, error) {
	tx := GormDB.Begin()
	group := JuustagramGroup{
		CommanderID:  commanderID,
		GroupID:      groupID,
		SkinID:       0,
		Favorite:     0,
		CurChatGroup: chatGroupID,
	}
	if err := tx.Create(&group).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	chatGroup := JuustagramChatGroup{
		CommanderID:   commanderID,
		GroupRecordID: group.ID,
		ChatGroupID:   chatGroupID,
		OpTime:        0,
		ReadFlag:      0,
	}
	if err := tx.Create(&chatGroup).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	group.ChatGroups = []JuustagramChatGroup{chatGroup}
	return &group, nil
}

func MarkJuustagramChatGroupsRead(commanderID uint32, chatGroupIDs []uint32) error {
	if len(chatGroupIDs) == 0 {
		return GormDB.Model(&JuustagramChatGroup{}).
			Where("commander_id = ?", commanderID).
			Update("read_flag", 1).
			Error
	}
	return GormDB.Model(&JuustagramChatGroup{}).
		Where("commander_id = ?", commanderID).
		Where("chat_group_id IN ?", chatGroupIDs).
		Update("read_flag", 1).
		Error
}
