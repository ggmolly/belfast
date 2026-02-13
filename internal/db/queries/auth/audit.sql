-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
  id,
  actor_account_id,
  actor_commander_id,
  method,
  path,
  status_code,
  permission_key,
  permission_op,
  action,
  metadata,
  created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
);
