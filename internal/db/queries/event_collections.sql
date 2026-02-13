-- Event collection queries

-- name: GetEventCollection :one
SELECT commander_id, collection_id, start_time, finish_time, ship_ids
FROM event_collections
WHERE commander_id = $1
  AND collection_id = $2;

-- name: CreateEventCollection :exec
INSERT INTO event_collections (
  commander_id,
  collection_id,
  start_time,
  finish_time,
  ship_ids
) VALUES (
  $1, $2, $3, $4, $5
)
ON CONFLICT DO NOTHING;

-- name: UpdateEventCollection :exec
UPDATE event_collections
SET start_time = $3,
    finish_time = $4,
    ship_ids = $5
WHERE commander_id = $1
  AND collection_id = $2;

-- name: DeleteEventCollection :exec
DELETE FROM event_collections
WHERE commander_id = $1
  AND collection_id = $2;

-- name: CountActiveEventCollections :one
SELECT COUNT(*)::bigint
FROM event_collections
WHERE commander_id = $1
  AND finish_time > 0;

-- name: ListBusyEventCollections :many
SELECT commander_id, collection_id, start_time, finish_time, ship_ids
FROM event_collections
WHERE commander_id = $1
  AND finish_time > 0;
