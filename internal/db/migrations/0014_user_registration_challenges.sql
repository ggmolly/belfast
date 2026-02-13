-- 0014_user_registration_challenges.sql

CREATE TABLE IF NOT EXISTS user_registration_challenges (
  id text PRIMARY KEY,
  commander_id bigint NOT NULL,
  pin text NOT NULL,
  password_hash text NOT NULL,
  password_algo text NOT NULL,
  status text NOT NULL,
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  created_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_registration_challenges_commander_status
  ON user_registration_challenges (commander_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_registration_challenges_pin_status_expires
  ON user_registration_challenges (pin, status, expires_at);
