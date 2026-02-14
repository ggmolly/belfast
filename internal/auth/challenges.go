package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
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
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	var user pgtype.Text
	if userID != nil {
		user = pgtype.Text{String: *userID, Valid: true}
	}
	row, err := db.DefaultStore.Queries.CreateAuthChallenge(ctx, gen.CreateAuthChallengeParams{
		ID:        entry.ID,
		UserID:    user,
		Type:      entry.Type,
		Challenge: entry.Challenge,
		ExpiresAt: pgtype.Timestamptz{Time: entry.ExpiresAt, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: entry.CreatedAt, Valid: true},
		Metadata:  entry.Metadata,
	})
	if err != nil {
		return nil, err
	}
	stored := orm.AuthChallenge{
		ID:        row.ID,
		Type:      row.Type,
		Challenge: row.Challenge,
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
		Metadata:  row.Metadata,
	}
	if row.UserID.Valid {
		v := row.UserID.String
		stored.UserID = &v
	}
	return &stored, nil
}

func LoadChallengeByUser(userID string, challengeType string) (*orm.AuthChallenge, *webauthn.SessionData, error) {
	if db.DefaultStore == nil {
		return nil, nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetLatestAuthChallengeByUser(ctx, gen.GetLatestAuthChallengeByUserParams{UserID: pgtype.Text{String: userID, Valid: true}, Type: challengeType})
	err = db.MapNotFound(err)
	if db.IsNotFound(err) {
		return nil, nil, ErrChallengeNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	entry := orm.AuthChallenge{
		ID:        row.ID,
		Type:      row.Type,
		Challenge: row.Challenge,
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
		Metadata:  row.Metadata,
	}
	if row.UserID.Valid {
		v := row.UserID.String
		entry.UserID = &v
	}
	return loadChallengeSession(&entry)
}

func LoadChallengeByChallenge(challenge string, challengeType string) (*orm.AuthChallenge, *webauthn.SessionData, error) {
	if db.DefaultStore == nil {
		return nil, nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetLatestAuthChallengeByChallenge(ctx, gen.GetLatestAuthChallengeByChallengeParams{Challenge: challenge, Type: challengeType})
	err = db.MapNotFound(err)
	if db.IsNotFound(err) {
		return nil, nil, ErrChallengeNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	entry := orm.AuthChallenge{
		ID:        row.ID,
		Type:      row.Type,
		Challenge: row.Challenge,
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
		Metadata:  row.Metadata,
	}
	if row.UserID.Valid {
		v := row.UserID.String
		entry.UserID = &v
	}
	return loadChallengeSession(&entry)
}

func DeleteChallenge(id string) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.DeleteAuthChallenge(ctx, id)
}

func loadChallengeSession(entry *orm.AuthChallenge) (*orm.AuthChallenge, *webauthn.SessionData, error) {
	var session webauthn.SessionData
	if err := json.Unmarshal(entry.Metadata, &session); err != nil {
		return nil, nil, err
	}
	return entry, &session, nil
}
