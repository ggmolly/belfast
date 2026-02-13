-- name: ListRoles :many
SELECT id, name, description, created_at, updated_at, updated_by
FROM roles
ORDER BY name ASC;

-- name: GetRoleByName :one
SELECT id, name, description, created_at, updated_at, updated_by
FROM roles
WHERE name = $1;

-- name: CreateRole :exec
INSERT INTO roles (id, name, description, created_at, updated_at, updated_by)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateRoleDescription :exec
UPDATE roles
SET description = $2
WHERE id = $1;

-- name: UpdateRoleUpdatedBy :exec
UPDATE roles
SET updated_by = $2,
    updated_at = $3
WHERE id = $1;
