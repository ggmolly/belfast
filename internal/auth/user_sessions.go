package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

func CreateUserSession(userID string, ip string, userAgent string, cfg config.AuthConfig) (*orm.UserSession, error) {
	csrfToken, err := NewToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	session := orm.UserSession{
		ID:            uuid.NewString(),
		UserID:        userID,
		CreatedAt:     now,
		LastSeenAt:    now,
		ExpiresAt:     now.Add(SessionTTL(cfg)),
		IPAddress:     ip,
		UserAgent:     userAgent,
		CSRFToken:     csrfToken,
		CSRFExpiresAt: now.Add(CSRFTTL(cfg)),
	}
	if err := orm.GormDB.Create(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func LoadUserSession(sessionID string) (*orm.UserSession, *orm.UserAccount, error) {
	var session orm.UserSession
	if err := orm.GormDB.First(&session, "id = ? AND revoked_at IS NULL", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrSessionNotFound
		}
		return nil, nil, err
	}
	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil, ErrSessionNotFound
	}
	var user orm.UserAccount
	if err := orm.GormDB.First(&user, "id = ?", session.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrSessionNotFound
		}
		return nil, nil, err
	}
	return &session, &user, nil
}

func TouchUserSession(sessionID string, lastSeen time.Time, expiresAt time.Time) error {
	updates := map[string]interface{}{
		"last_seen_at": lastSeen,
	}
	if !expiresAt.IsZero() {
		updates["expires_at"] = expiresAt
	}
	return orm.GormDB.Model(&orm.UserSession{}).Where("id = ?", sessionID).Updates(updates).Error
}

func RefreshUserCSRF(sessionID string, cfg config.AuthConfig) (string, time.Time, error) {
	token, err := NewToken(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(CSRFTTL(cfg))
	if err := orm.GormDB.Model(&orm.UserSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"csrf_token":      token,
		"csrf_expires_at": expiresAt,
	}).Error; err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func RevokeUserSession(sessionID string) error {
	return orm.GormDB.Model(&orm.UserSession{}).Where("id = ?", sessionID).Update("revoked_at", time.Now().UTC()).Error
}
