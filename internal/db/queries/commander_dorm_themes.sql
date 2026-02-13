-- Commander dorm theme queries

-- name: UpsertCommanderDormTheme :exec
INSERT INTO commander_dorm_themes (
  commander_id,
  theme_slot_id,
  name,
  furniture_put_list
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (commander_id, theme_slot_id)
DO UPDATE SET
  name = EXCLUDED.name,
  furniture_put_list = EXCLUDED.furniture_put_list;

-- name: DeleteCommanderDormTheme :exec
DELETE FROM commander_dorm_themes
WHERE commander_id = $1
  AND theme_slot_id = $2;

-- name: ListCommanderDormThemes :many
SELECT commander_id, theme_slot_id, name, furniture_put_list
FROM commander_dorm_themes
WHERE commander_id = $1
ORDER BY theme_slot_id ASC;

-- name: GetCommanderDormTheme :one
SELECT commander_id, theme_slot_id, name, furniture_put_list
FROM commander_dorm_themes
WHERE commander_id = $1
  AND theme_slot_id = $2;
