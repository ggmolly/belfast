-- Fleet queries

-- name: CreateFleet :one
INSERT INTO fleets (
  game_id,
  commander_id,
  name,
  ship_list,
  meowfficer_list
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id;

-- name: UpdateFleetName :exec
UPDATE fleets
SET name = $2
WHERE id = $1;

-- name: UpdateFleetShipList :exec
UPDATE fleets
SET ship_list = $2
WHERE id = $1;
