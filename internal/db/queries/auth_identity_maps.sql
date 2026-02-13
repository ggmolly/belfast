-- Account identity map queries

-- name: CreateYostarusMap :exec
INSERT INTO yostarus_maps (arg2, account_id)
VALUES ($1, $2);

-- name: GetYostarusMapByArg2 :one
SELECT arg2, account_id
FROM yostarus_maps
WHERE arg2 = $1;

-- name: UpsertDeviceAuthMap :exec
INSERT INTO device_auth_maps (device_id, arg2, account_id, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (device_id)
DO UPDATE SET
  arg2 = EXCLUDED.arg2,
  account_id = EXCLUDED.account_id,
  updated_at = now();

-- name: GetDeviceAuthMapByDeviceID :one
SELECT device_id, arg2, account_id, updated_at
FROM device_auth_maps
WHERE device_id = $1;
