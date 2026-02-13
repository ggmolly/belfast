-- 0016_auth_identity_maps.sql

CREATE TABLE IF NOT EXISTS yostarus_maps (
  arg2 bigint PRIMARY KEY,
  account_id bigint NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS device_auth_maps (
  device_id text PRIMARY KEY,
  arg2 bigint NOT NULL,
  account_id bigint NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
