-- Commander furniture queries

-- name: ListCommanderFurnitures :many
SELECT commander_id, furniture_id, count, get_time
FROM commander_furnitures
WHERE commander_id = $1
ORDER BY furniture_id ASC;

-- name: AddCommanderFurniture :exec
INSERT INTO commander_furnitures (
  commander_id,
  furniture_id,
  count,
  get_time
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (commander_id, furniture_id)
DO UPDATE SET
  count = commander_furnitures.count + EXCLUDED.count,
  get_time = EXCLUDED.get_time;
