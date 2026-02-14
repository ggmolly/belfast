-- Commander attire queries

-- name: ListCommanderAttires :many
SELECT commander_id, type, attire_id, expires_at, is_new
FROM commander_attires
WHERE commander_id = $1
ORDER BY type ASC, attire_id ASC;

-- name: ListCommanderAttiresByType :many
SELECT commander_id, type, attire_id, expires_at, is_new
FROM commander_attires
WHERE commander_id = $1
  AND type = $2
ORDER BY attire_id ASC;

-- name: GetCommanderAttire :one
SELECT commander_id, type, attire_id, expires_at, is_new
FROM commander_attires
WHERE commander_id = $1
  AND type = $2
  AND attire_id = $3;

-- name: UpsertCommanderAttire :exec
INSERT INTO commander_attires (
  commander_id,
  type,
  attire_id,
  expires_at,
  is_new
) VALUES (
  $1, $2, $3, $4, $5
)
ON CONFLICT (commander_id, type, attire_id)
DO UPDATE SET
  expires_at = EXCLUDED.expires_at,
  is_new = EXCLUDED.is_new;
