package orm

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type OwnedShip struct {
	OwnerID             uint32         `gorm:"type:int;not_null"`
	ShipID              uint32         `gorm:"not_null"`
	ID                  uint32         `gorm:"primary_key"`
	Level               uint32         `gorm:"default:1;not_null"`
	MaxLevel            uint32         `gorm:"default:50;not_null"`
	Intimacy            uint32         `gorm:"default:5000;not_null"`
	IsLocked            bool           `gorm:"default:false;not_null"`
	Propose             bool           `gorm:"default:false;not_null"`
	CommonFlag          bool           `gorm:"default:false;not_null"`
	BlueprintFlag       bool           `gorm:"default:false;not_null"`
	Proficiency         bool           `gorm:"default:false;not_null"`
	ActivityNPC         uint32         `gorm:"default:0;not_null"`
	CustomName          string         `gorm:"size:30;default:'';not_null"`
	ChangeNameTimestamp time.Time      `gorm:"type:timestamp;default:'1970-01-01 01:00:00';not_null"`
	CreateTime          time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	Energy              uint32         `gorm:"default:150;not_null"`
	SkinID              uint32         `gorm:"default:0;not_null"`
	IsSecretary         bool           `gorm:"default:false;not_null"`
	SecretaryPosition   *uint32        `gorm:"default:999;not_null"`
	SecretaryPhantomID  uint32         `gorm:"default:0;not_null"`
	DeletedAt           gorm.DeletedAt `gorm:"index"` // Soft delete

	Ship      Ship      `gorm:"foreignKey:ShipID;references:TemplateID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Commander Commander `gorm:"foreignKey:OwnerID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

var (
	ErrRenameInCooldown = errors.New("renaming is still in cooldown")
	ErrNotProposed      = errors.New("commander hasn't proposed this ship")
)

func (s *OwnedShip) Create() error {
	return GormDB.Create(s).Error
}

func (s *OwnedShip) Update() error {
	return GormDB.Save(s).Error
}

func (s *OwnedShip) Delete() error {
	return GormDB.Delete(s).Error
}

func (s *OwnedShip) ProposeShip() error {
	s.Propose = true
	return s.Update()
}

func (s *OwnedShip) SetFavorite(b uint32) error {
	var newState bool
	if b != 0 {
		newState = true
	}
	s.CommonFlag = newState
	return s.Update()
}

func (s *OwnedShip) RenameShip(newName string) error {
	if !s.Propose {
		return ErrNotProposed
	}
	// Check if the ship was renamed in the last 30 days
	if time.Since(s.ChangeNameTimestamp) < time.Hour*24*30 {
		return ErrRenameInCooldown
	}
	// XXX: We're not doing any verifications in the server-side
	s.CustomName = newName
	s.ChangeNameTimestamp = time.Now().Add(time.Hour * 24 * 30)
	return s.Update()
}
