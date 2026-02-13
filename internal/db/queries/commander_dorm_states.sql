-- Commander dorm state queries

-- name: GetCommanderDormStateByCommanderID :one
SELECT
  commander_id,
  level,
  food,
  food_max_increase_count,
  food_max_increase,
  floor_num,
  exp_pos,
  next_timestamp,
  load_exp,
  load_food,
  load_time,
  updated_at_unix_timestamp
FROM commander_dorm_states
WHERE commander_id = $1;

-- name: CreateCommanderDormState :exec
INSERT INTO commander_dorm_states (
  commander_id,
  level,
  food,
  food_max_increase_count,
  food_max_increase,
  floor_num,
  exp_pos,
  next_timestamp,
  load_exp,
  load_food,
  load_time,
  updated_at_unix_timestamp
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12
);
