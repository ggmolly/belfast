-- name: CreateSession :one
INSERT INTO sessions (
  id,
  account_id,
  created_at,
  last_seen_at,
  expires_at,
  ip_address,
  user_agent,
  revoked_at,
  csrf_token,
  csrf_expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING
  id,
  account_id,
  created_at,
  last_seen_at,
  expires_at,
  ip_address,
  user_agent,
  revoked_at,
  csrf_token,
  csrf_expires_at;

-- name: GetSessionByIDActive :one
SELECT
  id,
  account_id,
  created_at,
  last_seen_at,
  expires_at,
  ip_address,
  user_agent,
  revoked_at,
  csrf_token,
  csrf_expires_at
FROM sessions
WHERE id = $1
  AND revoked_at IS NULL;

-- name: TouchSession :exec
UPDATE sessions
SET last_seen_at = $2
WHERE id = $1;

-- name: TouchSessionWithExpires :exec
UPDATE sessions
SET last_seen_at = $2,
    expires_at = $3
WHERE id = $1;

-- name: RefreshSessionCSRF :exec
UPDATE sessions
SET csrf_token = $2,
    csrf_expires_at = $3
WHERE id = $1;

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = $2
WHERE id = $1;

-- name: RevokeSessionsAll :exec
UPDATE sessions
SET revoked_at = $2
WHERE account_id = $1
  AND revoked_at IS NULL;

-- name: RevokeSessionsExcept :exec
UPDATE sessions
SET revoked_at = $3
WHERE account_id = $1
  AND revoked_at IS NULL
  AND id <> $2;
