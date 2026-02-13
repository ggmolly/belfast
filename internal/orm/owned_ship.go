package orm

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type OwnedShip struct {
	OwnerID             uint32    `gorm:"type:int;not_null"`
	ShipID              uint32    `gorm:"not_null"`
	ID                  uint32    `gorm:"primary_key"`
	Level               uint32    `gorm:"default:1;not_null"`
	Exp                 uint32    `gorm:"default:0;not_null"`
	SurplusExp          uint32    `gorm:"default:0;not_null"`
	MaxLevel            uint32    `gorm:"default:50;not_null"`
	Intimacy            uint32    `gorm:"default:5000;not_null"`
	IsLocked            bool      `gorm:"default:false;not_null"`
	Propose             bool      `gorm:"default:false;not_null"`
	CommonFlag          bool      `gorm:"default:false;not_null"`
	BlueprintFlag       bool      `gorm:"default:false;not_null"`
	Proficiency         bool      `gorm:"default:false;not_null"`
	ActivityNPC         uint32    `gorm:"default:0;not_null"`
	CustomName          string    `gorm:"size:30;default:'';not_null"`
	ChangeNameTimestamp time.Time `gorm:"type:timestamp;default:'1970-01-01 01:00:00';not_null"`
	CreateTime          time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	Energy              uint32    `gorm:"default:150;not_null"`
	// Ship state fields used for dorm/backyard and other contexts.
	// Maps to protobuf.SHIPSTATE.
	State              uint32     `gorm:"default:0;not_null"`
	StateInfo1         uint32     `gorm:"default:0;not_null"`
	StateInfo2         uint32     `gorm:"default:0;not_null"`
	StateInfo3         uint32     `gorm:"default:0;not_null"`
	StateInfo4         uint32     `gorm:"default:0;not_null"`
	SkinID             uint32     `gorm:"default:0;not_null"`
	IsSecretary        bool       `gorm:"default:false;not_null"`
	SecretaryPosition  *uint32    `gorm:"default:999;not_null"`
	SecretaryPhantomID uint32     `gorm:"default:0;not_null"`
	DeletedAt          *time.Time `gorm:"index"`

	Ship       Ship                 `gorm:"foreignKey:ShipID;references:TemplateID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Commander  Commander            `gorm:"foreignKey:OwnerID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Equipments []OwnedShipEquipment `gorm:"foreignKey:ShipID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Transforms []OwnedShipTransform `gorm:"foreignKey:ShipID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Strengths  []OwnedShipStrength  `gorm:"foreignKey:ShipID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

var (
	ErrRenameInCooldown = errors.New("renaming is still in cooldown")
	ErrNotProposed      = errors.New("commander hasn't proposed this ship")
)

func (s *OwnedShip) Create() error {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.CreateOwnedShip(ctx, gen.CreateOwnedShipParams{OwnerID: int64(s.OwnerID), ShipID: int64(s.ShipID)})
	if err != nil {
		return err
	}
	s.ID = uint32(row.ID)
	s.CreateTime = row.CreateTime.Time
	s.ChangeNameTimestamp = row.ChangeNameTimestamp.Time
	return nil
}

func (s *OwnedShip) Update() error {
	ctx := context.Background()
	var deletedAt pgtype.Timestamptz
	if s.DeletedAt != nil {
		deletedAt = pgtype.Timestamptz{Time: *s.DeletedAt, Valid: true}
	}
	var secretaryPos pgtype.Int8
	if s.SecretaryPosition != nil {
		secretaryPos = pgtype.Int8{Int64: int64(*s.SecretaryPosition), Valid: true}
	}
	tag, err := db.DefaultStore.Queries.UpdateOwnedShip(ctx, gen.UpdateOwnedShipParams{
		OwnerID:             int64(s.OwnerID),
		ID:                  int64(s.ID),
		Level:               int64(s.Level),
		Exp:                 int64(s.Exp),
		SurplusExp:          int64(s.SurplusExp),
		MaxLevel:            int64(s.MaxLevel),
		Intimacy:            int64(s.Intimacy),
		IsLocked:            s.IsLocked,
		Propose:             s.Propose,
		CommonFlag:          s.CommonFlag,
		BlueprintFlag:       s.BlueprintFlag,
		Proficiency:         s.Proficiency,
		ActivityNpc:         int64(s.ActivityNPC),
		CustomName:          s.CustomName,
		ChangeNameTimestamp: pgtype.Timestamptz{Time: s.ChangeNameTimestamp, Valid: true},
		CreateTime:          pgtype.Timestamptz{Time: s.CreateTime, Valid: true},
		Energy:              int64(s.Energy),
		State:               int64(s.State),
		StateInfo1:          int64(s.StateInfo1),
		StateInfo2:          int64(s.StateInfo2),
		StateInfo3:          int64(s.StateInfo3),
		StateInfo4:          int64(s.StateInfo4),
		SkinID:              int64(s.SkinID),
		IsSecretary:         s.IsSecretary,
		SecretaryPosition:   secretaryPos,
		SecretaryPhantomID:  int64(s.SecretaryPhantomID),
		DeletedAt:           deletedAt,
	})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func (s *OwnedShip) Delete() error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Queries.SoftDeleteOwnedShip(ctx, gen.SoftDeleteOwnedShipParams{OwnerID: int64(s.OwnerID), ID: int64(s.ID)})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	now := time.Now()
	s.DeletedAt = &now
	return nil
}

func GetOwnedShipByOwnerAndID(ownerID uint32, ownedID uint32) (*OwnedShip, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT owner_id, ship_id, id, level, exp, surplus_exp, max_level, intimacy,
       is_locked, propose, common_flag, blueprint_flag, proficiency,
       activity_npc, custom_name, change_name_timestamp, create_time, energy,
       state, state_info1, state_info2, state_info3, state_info4, skin_id,
       is_secretary, secretary_position, secretary_phantom_id, deleted_at
FROM owned_ships
WHERE owner_id = $1
  AND id = $2
  AND deleted_at IS NULL
`, int64(ownerID), int64(ownedID))

	ship := OwnedShip{}
	var secretaryPosition *int64
	var deletedAt *time.Time
	err := row.Scan(
		&ship.OwnerID,
		&ship.ShipID,
		&ship.ID,
		&ship.Level,
		&ship.Exp,
		&ship.SurplusExp,
		&ship.MaxLevel,
		&ship.Intimacy,
		&ship.IsLocked,
		&ship.Propose,
		&ship.CommonFlag,
		&ship.BlueprintFlag,
		&ship.Proficiency,
		&ship.ActivityNPC,
		&ship.CustomName,
		&ship.ChangeNameTimestamp,
		&ship.CreateTime,
		&ship.Energy,
		&ship.State,
		&ship.StateInfo1,
		&ship.StateInfo2,
		&ship.StateInfo3,
		&ship.StateInfo4,
		&ship.SkinID,
		&ship.IsSecretary,
		&secretaryPosition,
		&ship.SecretaryPhantomID,
		&deletedAt,
	)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	if secretaryPosition != nil {
		pos := uint32(*secretaryPosition)
		ship.SecretaryPosition = &pos
	}
	ship.DeletedAt = deletedAt
	return &ship, nil
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
