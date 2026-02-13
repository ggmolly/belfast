-- Commander dorm floor layout queries

-- name: UpsertCommanderDormFloorLayout :exec
INSERT INTO commander_dorm_floor_layouts (
  commander_id,
  floor,
  furniture_put_list
) VALUES (
  $1, $2, $3
)
ON CONFLICT (commander_id, floor)
DO UPDATE SET furniture_put_list = EXCLUDED.furniture_put_list;

-- name: ListCommanderDormFloorLayouts :many
SELECT commander_id, floor, furniture_put_list
FROM commander_dorm_floor_layouts
WHERE commander_id = $1
ORDER BY floor ASC;
