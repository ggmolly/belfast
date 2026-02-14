-- name: UpsertRolePermission :exec
INSERT INTO role_permissions (
  role_id,
  permission_id,
  can_read_self,
  can_read_any,
  can_write_self,
  can_write_any,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (role_id, permission_id) DO UPDATE SET
  can_read_self = EXCLUDED.can_read_self,
  can_read_any = EXCLUDED.can_read_any,
  can_write_self = EXCLUDED.can_write_self,
  can_write_any = EXCLUDED.can_write_any,
  updated_at = EXCLUDED.updated_at;

-- name: ListRolePolicyRows :many
SELECT
  permissions.key AS key,
  role_permissions.can_read_self AS can_read_self,
  role_permissions.can_read_any AS can_read_any,
  role_permissions.can_write_self AS can_write_self,
  role_permissions.can_write_any AS can_write_any
FROM role_permissions
JOIN permissions ON permissions.id = role_permissions.permission_id
WHERE role_permissions.role_id = $1
  AND permissions.key = ANY($2::text[]);

-- name: ListEffectivePermissionRows :many
SELECT
  permissions.key AS key,
  role_permissions.can_read_self AS can_read_self,
  role_permissions.can_read_any AS can_read_any,
  role_permissions.can_write_self AS can_write_self,
  role_permissions.can_write_any AS can_write_any
FROM role_permissions
JOIN permissions ON permissions.id = role_permissions.permission_id
WHERE role_permissions.role_id = ANY($1::text[]);
