package auth

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ggmolly/belfast/internal/orm"
)

var ErrChallengeNotFound = errors.New("challenge not found")

func StoreChallenge(userID *string, challengeType string, session webauthn.SessionData, expiresAt time.Time) (*orm.AuthChallenge, error) {
	metadata, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	entry := orm.AuthChallenge{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      challengeType,
		Challenge: session.Challenge,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
		Metadata:  metadata,
	}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func LoadChallengeByUser(userID string, challengeType string) (*orm.AuthChallenge, *webauthn.SessionData, error) {
	var entry orm.AuthChallenge
	if err := orm.GormDB.Where("user_id = ? AND type = ?", userID, challengeType).Order("created_at desc").First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrChallengeNotFound
		}
		return nil, nil, err
	}
	return loadChallengeSession(&entry)
}

func LoadChallengeByChallenge(challenge string, challengeType string) (*orm.AuthChallenge, *webauthn.SessionData, error) {
	var entry orm.AuthChallenge
	if err := orm.GormDB.Where("challenge = ? AND type = ?", challenge, challengeType).Order("created_at desc").First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrChallengeNotFound
		}
		return nil, nil, err
	}
	return loadChallengeSession(&entry)
}

func DeleteChallenge(id string) error {
	return orm.GormDB.Delete(&orm.AuthChallenge{}, "id = ?", id).Error
}

func loadChallengeSession(entry *orm.AuthChallenge) (*orm.AuthChallenge, *webauthn.SessionData, error) {
	var session webauthn.SessionData
	if err := json.Unmarshal(entry.Metadata, &session); err != nil {
		return nil, nil, err
	}
	return entry, &session, nil
}
