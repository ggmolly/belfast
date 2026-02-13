-- name: ListPermissions :many
SELECT id, key, description, created_at, updated_at
FROM permissions
ORDER BY key ASC;

-- name: GetPermissionByKey :one
SELECT id, key, description, created_at, updated_at
FROM permissions
WHERE key = $1;

-- name: CreatePermission :exec
INSERT INTO permissions (id, key, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);
