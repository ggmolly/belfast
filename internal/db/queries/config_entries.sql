-- Config entries lookup queries

-- name: GetConfigEntry :one
SELECT
  id,
  category,
  key,
  data
FROM config_entries
WHERE category = $1
  AND key = $2;

-- name: ListConfigEntriesByCategory :many
SELECT
  id,
  category,
  key,
  data
FROM config_entries
WHERE category = $1
ORDER BY key ASC;
