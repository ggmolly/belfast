-- 0017_escort_states.sql

CREATE TABLE IF NOT EXISTS escort_states (
  id bigserial PRIMARY KEY,
  account_id bigint NOT NULL,
  line_id bigint NOT NULL,
  award_timestamp bigint NOT NULL DEFAULT 0,
  flash_timestamp bigint NOT NULL DEFAULT 0,
  map_positions jsonb,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (account_id, line_id)
);
