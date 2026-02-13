-- Commander common flag queries

-- name: ListCommanderCommonFlags :many
SELECT flag_id
FROM commander_common_flags
WHERE commander_id = $1
ORDER BY flag_id ASC;

-- name: CreateCommanderCommonFlag :exec
INSERT INTO commander_common_flags (commander_id, flag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteCommanderCommonFlag :exec
DELETE FROM commander_common_flags
WHERE commander_id = $1
  AND flag_id = $2;
