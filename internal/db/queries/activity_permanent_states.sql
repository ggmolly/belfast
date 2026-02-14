-- Activity permanent state queries

-- name: GetActivityPermanentState :one
SELECT commander_id, current_activity_id, finished_activity_ids
FROM activity_permanent_states
WHERE commander_id = $1;

-- name: CreateActivityPermanentState :one
INSERT INTO activity_permanent_states (
  commander_id,
  current_activity_id,
  finished_activity_ids
) VALUES (
  $1,
  0,
  '[]'::jsonb
)
RETURNING commander_id, current_activity_id, finished_activity_ids;

-- name: UpsertActivityPermanentState :exec
INSERT INTO activity_permanent_states (
  commander_id,
  current_activity_id,
  finished_activity_ids
) VALUES (
  $1,
  $2,
  $3
)
ON CONFLICT (commander_id)
DO UPDATE SET
  current_activity_id = EXCLUDED.current_activity_id,
  finished_activity_ids = EXCLUDED.finished_activity_ids;
