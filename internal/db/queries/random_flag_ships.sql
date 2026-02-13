-- Random flag ship queries

-- name: UpsertRandomFlagShip :exec
INSERT INTO random_flag_ships (
  commander_id,
  ship_id,
  phantom_id,
  enabled
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (commander_id, ship_id, phantom_id)
DO UPDATE SET enabled = EXCLUDED.enabled;

-- name: DeleteRandomFlagShip :exec
DELETE FROM random_flag_ships
WHERE commander_id = $1
  AND ship_id = $2
  AND phantom_id = $3;

-- name: ListEnabledRandomFlagShipsByCommander :many
SELECT commander_id, ship_id, phantom_id, enabled
FROM random_flag_ships
WHERE commander_id = $1
  AND enabled = true;

-- name: ListEnabledRandomFlagShipsByCommanderAndShips :many
SELECT commander_id, ship_id, phantom_id, enabled
FROM random_flag_ships
WHERE commander_id = $1
  AND enabled = true
  AND ship_id = ANY($2::bigint[]);
