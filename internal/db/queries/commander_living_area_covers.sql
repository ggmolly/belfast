-- Commander living area cover queries

-- name: ListCommanderLivingAreaCovers :many
SELECT commander_id, cover_id, unlocked_at, is_new
FROM commander_living_area_covers
WHERE commander_id = $1
ORDER BY cover_id ASC;

-- name: GetCommanderLivingAreaCover :one
SELECT commander_id, cover_id, unlocked_at, is_new
FROM commander_living_area_covers
WHERE commander_id = $1
  AND cover_id = $2;

-- name: UpsertCommanderLivingAreaCover :exec
INSERT INTO commander_living_area_covers (
  commander_id,
  cover_id,
  unlocked_at,
  is_new
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (commander_id, cover_id)
DO UPDATE SET
  unlocked_at = EXCLUDED.unlocked_at,
  is_new = EXCLUDED.is_new;

-- name: DeleteCommanderLivingAreaCover :exec
DELETE FROM commander_living_area_covers
WHERE commander_id = $1
  AND cover_id = $2;

-- name: UpdateCommanderLivingAreaCoverIsNew :exec
UPDATE commander_living_area_covers
SET is_new = $3
WHERE commander_id = $1
  AND cover_id = $2;
