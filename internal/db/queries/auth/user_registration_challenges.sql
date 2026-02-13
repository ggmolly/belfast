-- User registration challenge queries

-- name: GetLatestPendingRegistrationChallengeByCommander :one
SELECT id, commander_id, pin, password_hash, password_algo, status, expires_at, consumed_at, created_at
FROM user_registration_challenges
WHERE commander_id = $1
  AND status = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: GetPendingRegistrationChallengeByPin :one
SELECT id, commander_id, pin, password_hash, password_algo, status, expires_at, consumed_at, created_at
FROM user_registration_challenges
WHERE pin = $1
  AND status = $2
  AND expires_at > $3
LIMIT 1;

-- name: GetPendingRegistrationChallengeByPinForUpdate :one
SELECT id, commander_id, pin, password_hash, password_algo, status, expires_at, consumed_at, created_at
FROM user_registration_challenges
WHERE pin = $1
  AND status = $2
FOR UPDATE;

-- name: GetRegistrationChallengeByIDForUpdate :one
SELECT id, commander_id, pin, password_hash, password_algo, status, expires_at, consumed_at, created_at
FROM user_registration_challenges
WHERE id = $1
FOR UPDATE;

-- name: GetRegistrationChallengeByID :one
SELECT id, commander_id, pin, password_hash, password_algo, status, expires_at, consumed_at, created_at
FROM user_registration_challenges
WHERE id = $1;

-- name: CreateRegistrationChallenge :exec
INSERT INTO user_registration_challenges (
  id,
  commander_id,
  pin,
  password_hash,
  password_algo,
  status,
  expires_at,
  consumed_at,
  created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: UpdateRegistrationChallengeStatus :exec
UPDATE user_registration_challenges
SET status = $2,
    consumed_at = $3
WHERE id = $1;
