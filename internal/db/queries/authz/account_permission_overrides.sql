-- name: ListAccountOverrides :many
SELECT
  permissions.key AS key,
  account_permission_overrides.mode AS mode,
  account_permission_overrides.can_read_self AS can_read_self,
  account_permission_overrides.can_read_any AS can_read_any,
  account_permission_overrides.can_write_self AS can_write_self,
  account_permission_overrides.can_write_any AS can_write_any
FROM account_permission_overrides
JOIN permissions ON permissions.id = account_permission_overrides.permission_id
WHERE account_permission_overrides.account_id = $1
ORDER BY permissions.key ASC;

-- name: DeleteAccountOverridesByAccountID :exec
DELETE FROM account_permission_overrides
WHERE account_id = $1;

-- name: CreateAccountPermissionOverride :exec
INSERT INTO account_permission_overrides (
  account_id,
  permission_id,
  mode,
  can_read_self,
  can_read_any,
  can_write_self,
  can_write_any,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: ListAccountOverrideRows :many
SELECT
  permissions.key AS key,
  account_permission_overrides.mode AS mode,
  account_permission_overrides.can_read_self AS can_read_self,
  account_permission_overrides.can_read_any AS can_read_any,
  account_permission_overrides.can_write_self AS can_write_self,
  account_permission_overrides.can_write_any AS can_write_any
FROM account_permission_overrides
JOIN permissions ON permissions.id = account_permission_overrides.permission_id
WHERE account_permission_overrides.account_id = $1;
