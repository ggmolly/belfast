package orm

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ggmolly/belfast/internal/authz"
)

const (
	UserRegistrationStatusPending  = "pending"
	UserRegistrationStatusConsumed = "consumed"
	UserRegistrationStatusExpired  = "expired"
)

var (
	ErrUserAccountExists                = errors.New("user account exists")
	ErrRegistrationChallengeExists      = errors.New("registration challenge exists")
	ErrRegistrationChallengeNotFound    = errors.New("registration challenge not found")
	ErrRegistrationChallengeConsumed    = errors.New("registration challenge consumed")
	ErrRegistrationChallengeExpired     = errors.New("registration challenge expired")
	ErrRegistrationChallengeMismatch    = errors.New("registration challenge mismatch")
	ErrRegistrationChallengePinMismatch = errors.New("registration challenge pin mismatch")
	ErrRegistrationPinExists            = errors.New("registration pin exists")
)

func CreateUserRegistrationChallenge(commanderID uint32, pin string, passwordHash string, passwordAlgo string, expiresAt time.Time, now time.Time) (*UserRegistrationChallenge, error) {
	var existingCount int64
	if err := GormDB.Model(&Account{}).Where("commander_id = ?", commanderID).Count(&existingCount).Error; err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, ErrUserAccountExists
	}

	var pending UserRegistrationChallenge
	if err := GormDB.Where("commander_id = ? AND status = ?", commanderID, UserRegistrationStatusPending).Order("created_at desc").First(&pending).Error; err == nil {
		_ = GormDB.Model(&UserRegistrationChallenge{}).Where("id = ?", pending.ID).Update("status", UserRegistrationStatusExpired).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var pinMatch UserRegistrationChallenge
	if err := GormDB.Where("pin = ? AND status = ? AND expires_at > ?", pin, UserRegistrationStatusPending, now).First(&pinMatch).Error; err == nil {
		return nil, ErrRegistrationPinExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	entry := UserRegistrationChallenge{
		ID:           uuid.NewString(),
		CommanderID:  commanderID,
		Pin:          pin,
		PasswordHash: passwordHash,
		PasswordAlgo: passwordAlgo,
		Status:       UserRegistrationStatusPending,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
	}
	if err := GormDB.Create(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func ConsumeUserRegistrationChallenge(commanderID uint32, pin string, now time.Time) (*Account, error) {
	var account *Account
	err := GormDB.Transaction(func(tx *gorm.DB) error {
		var role Role
		if err := tx.First(&role, "name = ?", authz.RolePlayer).Error; err != nil {
			return err
		}
		var challenge UserRegistrationChallenge
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("pin = ? AND status = ?", pin, UserRegistrationStatusPending).First(&challenge).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRegistrationChallengeNotFound
			}
			return err
		}
		if challenge.CommanderID != commanderID {
			return ErrRegistrationChallengeMismatch
		}
		if !challenge.ExpiresAt.After(now) {
			_ = tx.Model(&UserRegistrationChallenge{}).Where("id = ?", challenge.ID).Update("status", UserRegistrationStatusExpired).Error
			return ErrRegistrationChallengeExpired
		}

		var existing Account
		if err := tx.First(&existing, "commander_id = ?", commanderID).Error; err == nil {
			return ErrUserAccountExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		created := Account{
			ID:                uuid.NewString(),
			CommanderID:       &commanderID,
			PasswordHash:      challenge.PasswordHash,
			PasswordAlgo:      challenge.PasswordAlgo,
			PasswordUpdatedAt: now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		link := AccountRole{AccountID: created.ID, RoleID: role.ID, CreatedAt: now}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"status":      UserRegistrationStatusConsumed,
			"consumed_at": now,
		}
		if err := tx.Model(&UserRegistrationChallenge{}).Where("id = ?", challenge.ID).Updates(updates).Error; err != nil {
			return err
		}
		account = &created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func ConsumeUserRegistrationChallengeByID(id string, pin string, now time.Time) (*Account, error) {
	var account *Account
	err := GormDB.Transaction(func(tx *gorm.DB) error {
		var role Role
		if err := tx.First(&role, "name = ?", authz.RolePlayer).Error; err != nil {
			return err
		}
		var challenge UserRegistrationChallenge
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&challenge, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRegistrationChallengeNotFound
			}
			return err
		}
		switch challenge.Status {
		case UserRegistrationStatusConsumed:
			return ErrRegistrationChallengeConsumed
		case UserRegistrationStatusExpired:
			return ErrRegistrationChallengeExpired
		}
		if !challenge.ExpiresAt.After(now) {
			_ = tx.Model(&UserRegistrationChallenge{}).Where("id = ?", challenge.ID).Update("status", UserRegistrationStatusExpired).Error
			return ErrRegistrationChallengeExpired
		}
		if challenge.Pin != pin {
			return ErrRegistrationChallengePinMismatch
		}

		var existing Account
		if err := tx.First(&existing, "commander_id = ?", challenge.CommanderID).Error; err == nil {
			return ErrUserAccountExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		commanderID := challenge.CommanderID
		created := Account{
			ID:                uuid.NewString(),
			CommanderID:       &commanderID,
			PasswordHash:      challenge.PasswordHash,
			PasswordAlgo:      challenge.PasswordAlgo,
			PasswordUpdatedAt: now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		link := AccountRole{AccountID: created.ID, RoleID: role.ID, CreatedAt: now}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"status":      UserRegistrationStatusConsumed,
			"consumed_at": now,
		}
		if err := tx.Model(&UserRegistrationChallenge{}).Where("id = ?", challenge.ID).Updates(updates).Error; err != nil {
			return err
		}
		account = &created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func GetUserRegistrationChallenge(id string) (*UserRegistrationChallenge, error) {
	var challenge UserRegistrationChallenge
	if err := GormDB.First(&challenge, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}
