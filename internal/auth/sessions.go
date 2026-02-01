package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

var ErrSessionNotFound = errors.New("session not found")

func CreateSession(userID string, ip string, userAgent string, cfg config.AuthConfig) (*orm.AdminSession, error) {
	csrfToken, err := NewToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	session := orm.AdminSession{
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

func LoadSession(sessionID string) (*orm.AdminSession, *orm.AdminUser, error) {
	var session orm.AdminSession
	if err := orm.GormDB.First(&session, "id = ? AND revoked_at IS NULL", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrSessionNotFound
		}
		return nil, nil, err
	}
	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil, ErrSessionNotFound
	}
	var user orm.AdminUser
	if err := orm.GormDB.First(&user, "id = ?", session.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrSessionNotFound
		}
		return nil, nil, err
	}
	return &session, &user, nil
}

func TouchSession(sessionID string, lastSeen time.Time, expiresAt time.Time) error {
	updates := map[string]interface{}{
		"last_seen_at": lastSeen,
	}
	if !expiresAt.IsZero() {
		updates["expires_at"] = expiresAt
	}
	return orm.GormDB.Model(&orm.AdminSession{}).Where("id = ?", sessionID).Updates(updates).Error
}

func RefreshCSRF(sessionID string, cfg config.AuthConfig) (string, time.Time, error) {
	token, err := NewToken(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(CSRFTTL(cfg))
	if err := orm.GormDB.Model(&orm.AdminSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"csrf_token":      token,
		"csrf_expires_at": expiresAt,
	}).Error; err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func RevokeSession(sessionID string) error {
	return orm.GormDB.Model(&orm.AdminSession{}).Where("id = ?", sessionID).Update("revoked_at", time.Now().UTC()).Error
}

func RevokeUserSessions(userID string, exceptSessionID string) error {
	query := orm.GormDB.Model(&orm.AdminSession{}).Where("user_id = ? AND revoked_at IS NULL", userID)
	if exceptSessionID != "" {
		query = query.Where("id <> ?", exceptSessionID)
	}
	return query.Update("revoked_at", time.Now().UTC()).Error
}
