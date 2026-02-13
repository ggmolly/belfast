-- Ship queries

-- name: UpsertShipRecord :exec
INSERT INTO ships (
  template_id,
  name,
  english_name,
  rarity_id,
  star,
  type,
  nationality,
  build_time,
  pool_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (template_id)
DO UPDATE SET
  name = EXCLUDED.name,
  english_name = EXCLUDED.english_name,
  rarity_id = EXCLUDED.rarity_id,
  star = EXCLUDED.star,
  type = EXCLUDED.type,
  nationality = EXCLUDED.nationality,
  build_time = EXCLUDED.build_time,
  pool_id = EXCLUDED.pool_id;

-- name: GetShip :one
SELECT template_id, name, english_name, rarity_id, star, type, nationality, build_time, pool_id
FROM ships
WHERE template_id = $1;

-- name: DeleteShip :execresult
DELETE FROM ships
WHERE template_id = $1;

-- name: CountShipByTemplateID :one
SELECT COUNT(*)::bigint
FROM ships
WHERE template_id = $1;

-- name: GetRandomPoolShip :one
SELECT template_id, name, english_name, rarity_id, star, type, nationality, build_time, pool_id
FROM ships
WHERE pool_id = $1
  AND rarity_id = $2
ORDER BY RANDOM()
LIMIT 1;
