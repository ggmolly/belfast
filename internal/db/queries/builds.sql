-- Build queue queries

-- name: CreateBuild :one
INSERT INTO builds (
  builder_id,
  ship_id,
  pool_id,
  finishes_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING id;

-- name: UpdateBuild :exec
UPDATE builds
SET builder_id = $2,
    ship_id = $3,
    pool_id = $4,
    finishes_at = $5
WHERE id = $1;

-- name: DeleteBuildByID :execresult
DELETE FROM builds
WHERE id = $1;

-- name: GetBuildByID :one
SELECT id, builder_id, ship_id, pool_id, finishes_at
FROM builds
WHERE id = $1;

-- name: ListBuildsRangeByBuilderID :many
SELECT id, builder_id, ship_id, pool_id, finishes_at
FROM builds
WHERE builder_id = $1
ORDER BY id ASC
OFFSET $2
LIMIT $3;
