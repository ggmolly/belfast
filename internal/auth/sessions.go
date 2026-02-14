package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
	"github.com/ggmolly/belfast/internal/orm"
)

var ErrSessionNotFound = errors.New("session not found")

func CreateSession(accountID string, ip string, userAgent string, cfg config.AuthConfig) (*orm.Session, error) {
	csrfToken, err := NewToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	session := orm.Session{
		ID:            uuid.NewString(),
		AccountID:     accountID,
		CreatedAt:     now,
		LastSeenAt:    now,
		ExpiresAt:     now.Add(SessionTTL(cfg)),
		IPAddress:     ip,
		UserAgent:     userAgent,
		CSRFToken:     csrfToken,
		CSRFExpiresAt: now.Add(CSRFTTL(cfg)),
	}
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.CreateSession(ctx, gen.CreateSessionParams{
		ID:            session.ID,
		AccountID:     session.AccountID,
		CreatedAt:     pgtype.Timestamptz{Time: session.CreatedAt, Valid: true},
		LastSeenAt:    pgtype.Timestamptz{Time: session.LastSeenAt, Valid: true},
		ExpiresAt:     pgtype.Timestamptz{Time: session.ExpiresAt, Valid: true},
		IpAddress:     session.IPAddress,
		UserAgent:     session.UserAgent,
		RevokedAt:     pgtype.Timestamptz{},
		CsrfToken:     session.CSRFToken,
		CsrfExpiresAt: pgtype.Timestamptz{Time: session.CSRFExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	stored := orm.Session{
		ID:            row.ID,
		AccountID:     row.AccountID,
		CreatedAt:     row.CreatedAt.Time,
		LastSeenAt:    row.LastSeenAt.Time,
		ExpiresAt:     row.ExpiresAt.Time,
		IPAddress:     row.IpAddress,
		UserAgent:     row.UserAgent,
		CSRFToken:     row.CsrfToken,
		CSRFExpiresAt: row.CsrfExpiresAt.Time,
	}
	if row.RevokedAt.Valid {
		v := row.RevokedAt.Time
		stored.RevokedAt = &v
	}
	return &stored, nil
}

func LoadSession(sessionID string) (*orm.Session, *orm.Account, error) {
	var session orm.Session
	if db.DefaultStore == nil {
		return nil, nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetSessionByIDActive(ctx, sessionID)
	err = db.MapNotFound(err)
	if db.IsNotFound(err) {
		return nil, nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	session = orm.Session{
		ID:            row.ID,
		AccountID:     row.AccountID,
		CreatedAt:     row.CreatedAt.Time,
		LastSeenAt:    row.LastSeenAt.Time,
		ExpiresAt:     row.ExpiresAt.Time,
		IPAddress:     row.IpAddress,
		UserAgent:     row.UserAgent,
		CSRFToken:     row.CsrfToken,
		CSRFExpiresAt: row.CsrfExpiresAt.Time,
	}
	if row.RevokedAt.Valid {
		v := row.RevokedAt.Time
		session.RevokedAt = &v
	}
	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil, ErrSessionNotFound
	}
	var account orm.Account
	rowAccount, err := db.DefaultStore.Queries.GetAccountByID(ctx, session.AccountID)
	err = db.MapNotFound(err)
	if db.IsNotFound(err) {
		return nil, nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	account = orm.Account{
		ID:                 rowAccount.ID,
		PasswordHash:       rowAccount.PasswordHash,
		PasswordAlgo:       rowAccount.PasswordAlgo,
		PasswordUpdatedAt:  rowAccount.PasswordUpdatedAt.Time,
		IsAdmin:            rowAccount.IsAdmin,
		WebAuthnUserHandle: rowAccount.WebAuthnUserHandle,
		CreatedAt:          rowAccount.CreatedAt.Time,
		UpdatedAt:          rowAccount.UpdatedAt.Time,
	}
	if rowAccount.Username.Valid {
		v := rowAccount.Username.String
		account.Username = &v
	}
	if rowAccount.UsernameNormalized.Valid {
		v := rowAccount.UsernameNormalized.String
		account.UsernameNormalized = &v
	}
	if rowAccount.CommanderID.Valid {
		v, convErr := orm.Uint32FromInt64Checked(rowAccount.CommanderID.Int64)
		if convErr != nil {
			return nil, nil, convErr
		}
		account.CommanderID = &v
	}
	if rowAccount.DisabledAt.Valid {
		v := rowAccount.DisabledAt.Time
		account.DisabledAt = &v
	}
	if rowAccount.LastLoginAt.Valid {
		v := rowAccount.LastLoginAt.Time
		account.LastLoginAt = &v
	}
	return &session, &account, nil
}

func TouchSession(sessionID string, lastSeen time.Time, expiresAt time.Time) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	if expiresAt.IsZero() {
		return db.DefaultStore.Queries.TouchSession(ctx, gen.TouchSessionParams{ID: sessionID, LastSeenAt: pgtype.Timestamptz{Time: lastSeen, Valid: true}})
	}
	return db.DefaultStore.Queries.TouchSessionWithExpires(ctx, gen.TouchSessionWithExpiresParams{ID: sessionID, LastSeenAt: pgtype.Timestamptz{Time: lastSeen, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true}})
}

func RefreshCSRF(sessionID string, cfg config.AuthConfig) (string, time.Time, error) {
	token, err := NewToken(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(CSRFTTL(cfg))
	if db.DefaultStore == nil {
		return "", time.Time{}, errors.New("db not initialized")
	}
	ctx := context.Background()
	if err := db.DefaultStore.Queries.RefreshSessionCSRF(ctx, gen.RefreshSessionCSRFParams{ID: sessionID, CsrfToken: token, CsrfExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true}}); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func RevokeSession(sessionID string) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.RevokeSession(ctx, gen.RevokeSessionParams{ID: sessionID, RevokedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}})
}

func RevokeSessions(accountID string, exceptSessionID string) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if exceptSessionID == "" {
		return db.DefaultStore.Queries.RevokeSessionsAll(ctx, gen.RevokeSessionsAllParams{AccountID: accountID, RevokedAt: pgtype.Timestamptz{Time: now, Valid: true}})
	}
	return db.DefaultStore.Queries.RevokeSessionsExcept(ctx, gen.RevokeSessionsExceptParams{AccountID: accountID, ID: exceptSessionID, RevokedAt: pgtype.Timestamptz{Time: now, Valid: true}})
}
