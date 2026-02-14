-- name: ListAccountRoleNames :many
SELECT roles.name
FROM account_roles
JOIN roles ON roles.id = account_roles.role_id
WHERE account_roles.account_id = $1
ORDER BY roles.name ASC;

-- name: DeleteAccountRolesByAccountID :exec
DELETE FROM account_roles
WHERE account_id = $1;

-- name: CreateAccountRoleLink :exec
INSERT INTO account_roles (account_id, role_id, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (account_id, role_id) DO NOTHING;

-- name: CountAccountsWithRoleExceptAccount :one
SELECT COUNT(*)::bigint
FROM account_roles
WHERE role_id = $1
  AND ($2::text = '' OR account_id <> $2);

-- name: ListAccountRoleIDs :many
SELECT role_id
FROM account_roles
WHERE account_id = $1;
