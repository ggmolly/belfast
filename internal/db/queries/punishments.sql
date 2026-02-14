-- Punishment queries

-- name: UpsertPunishment :one
INSERT INTO punishments (
  punished_id,
  lift_timestamp,
  is_permanent
) VALUES (
  $1, $2, $3
)
ON CONFLICT (id)
DO UPDATE SET
  punished_id = EXCLUDED.punished_id,
  lift_timestamp = EXCLUDED.lift_timestamp,
  is_permanent = EXCLUDED.is_permanent
RETURNING id;

-- name: UpdatePunishment :execresult
UPDATE punishments
SET punished_id = $2,
    lift_timestamp = $3,
    is_permanent = $4
WHERE id = $1;

-- name: GetPunishment :one
SELECT id, punished_id, lift_timestamp, is_permanent
FROM punishments
WHERE id = $1;

-- name: DeletePunishment :execresult
DELETE FROM punishments
WHERE id = $1;
