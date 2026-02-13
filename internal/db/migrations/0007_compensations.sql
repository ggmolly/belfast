-- 0007_compensations.sql
-- Milestone 6: compensation mail-like rewards.

CREATE TABLE IF NOT EXISTS compensations (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  title text NOT NULL,
  text text NOT NULL,
  send_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at timestamptz NOT NULL,
  attach_flag boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_compensations_commander_id ON compensations(commander_id);

CREATE TABLE IF NOT EXISTS compensation_attachments (
  id bigserial PRIMARY KEY,
  compensation_id bigint NOT NULL REFERENCES compensations(id) ON DELETE CASCADE,
  type bigint NOT NULL,
  item_id bigint NOT NULL,
  quantity bigint NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_compensation_attachments_compensation_id ON compensation_attachments(compensation_id);
