package orm

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type CommanderTB struct {
	CommanderID uint32 `gorm:"primary_key"`
	State       []byte `gorm:"type:blob;not_null"`
	Permanent   []byte `gorm:"type:blob;not_null"`
}

func GetCommanderTB(db *gorm.DB, commanderID uint32) (*CommanderTB, error) {
	var entry CommanderTB
	if err := db.Where("commander_id = ?", commanderID).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func NewCommanderTB(commanderID uint32, info *protobuf.TBINFO, permanent *protobuf.TBPERMANENT) (*CommanderTB, error) {
	stateBytes, err := proto.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tb state: %w", err)
	}
	permanentBytes, err := proto.Marshal(permanent)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tb permanent state: %w", err)
	}
	return &CommanderTB{
		CommanderID: commanderID,
		State:       stateBytes,
		Permanent:   permanentBytes,
	}, nil
}

func (entry *CommanderTB) Decode() (*protobuf.TBINFO, *protobuf.TBPERMANENT, error) {
	state := &protobuf.TBINFO{}
	if err := proto.Unmarshal(entry.State, state); err != nil {
		return nil, nil, fmt.Errorf("failed to decode tb state: %w", err)
	}
	permanent := &protobuf.TBPERMANENT{}
	if err := proto.Unmarshal(entry.Permanent, permanent); err != nil {
		return nil, nil, fmt.Errorf("failed to decode tb permanent state: %w", err)
	}
	return state, permanent, nil
}

func (entry *CommanderTB) Encode(info *protobuf.TBINFO, permanent *protobuf.TBPERMANENT) error {
	stateBytes, err := proto.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to encode tb state: %w", err)
	}
	permanentBytes, err := proto.Marshal(permanent)
	if err != nil {
		return fmt.Errorf("failed to encode tb permanent state: %w", err)
	}
	entry.State = stateBytes
	entry.Permanent = permanentBytes
	return nil
}

func SaveCommanderTB(db *gorm.DB, entry *CommanderTB, info *protobuf.TBINFO, permanent *protobuf.TBPERMANENT) error {
	if err := entry.Encode(info, permanent); err != nil {
		return err
	}
	return db.Save(entry).Error
}
