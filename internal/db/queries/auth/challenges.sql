-- name: CreateAuthChallenge :one
INSERT INTO auth_challenges (
  id,
  user_id,
  type,
  challenge,
  expires_at,
  created_at,
  metadata
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING
  id,
  user_id,
  type,
  challenge,
  expires_at,
  created_at,
  metadata;

-- name: GetLatestAuthChallengeByUser :one
SELECT
  id,
  user_id,
  type,
  challenge,
  expires_at,
  created_at,
  metadata
FROM auth_challenges
WHERE user_id = $1
  AND type = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLatestAuthChallengeByChallenge :one
SELECT
  id,
  user_id,
  type,
  challenge,
  expires_at,
  created_at,
  metadata
FROM auth_challenges
WHERE challenge = $1
  AND type = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteAuthChallenge :exec
DELETE FROM auth_challenges
WHERE id = $1;
