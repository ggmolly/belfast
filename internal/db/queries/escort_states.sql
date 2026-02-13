-- Escort state queries

-- name: ListEscortStatesByAccountID :many
SELECT id, account_id, line_id, award_timestamp, flash_timestamp, map_positions, created_at, updated_at
FROM escort_states
WHERE account_id = $1
ORDER BY line_id ASC;

-- name: UpsertEscortStateByAccountLine :exec
INSERT INTO escort_states (
  account_id,
  line_id,
  award_timestamp,
  flash_timestamp,
  map_positions,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, now(), now()
)
ON CONFLICT (account_id, line_id)
DO UPDATE SET
  award_timestamp = EXCLUDED.award_timestamp,
  flash_timestamp = EXCLUDED.flash_timestamp,
  map_positions = EXCLUDED.map_positions,
  updated_at = now();
