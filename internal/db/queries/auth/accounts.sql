-- name: GetAccountByID :one
SELECT
  id,
  username,
  username_normalized,
  commander_id,
  password_hash,
  password_algo,
  password_updated_at,
  is_admin,
  disabled_at,
  last_login_at,
  web_authn_user_handle,
  created_at,
  updated_at
FROM accounts
WHERE id = $1;

-- name: CountAccountsByCommanderID :one
SELECT COUNT(*)::bigint
FROM accounts
WHERE commander_id = $1;

-- name: GetAccountByCommanderID :one
SELECT
  id,
  username,
  username_normalized,
  commander_id,
  password_hash,
  password_algo,
  password_updated_at,
  is_admin,
  disabled_at,
  last_login_at,
  web_authn_user_handle,
  created_at,
  updated_at
FROM accounts
WHERE commander_id = $1;

-- name: CreateAccount :exec
INSERT INTO accounts (
  id,
  commander_id,
  password_hash,
  password_algo,
  password_updated_at,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
);

-- name: UpdateAccountWebAuthnUserHandle :exec
UPDATE accounts
SET web_authn_user_handle = $2,
    updated_at = $3
WHERE id = $1;
