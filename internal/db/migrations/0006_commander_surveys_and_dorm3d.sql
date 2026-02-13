-- 0006_commander_surveys_and_dorm3d.sql
-- Milestone 6: additional commander-scoped state tables.

CREATE TABLE IF NOT EXISTS commander_surveys (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  survey_id bigint NOT NULL,
  completed_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (commander_id, survey_id)
);

CREATE TABLE IF NOT EXISTS dorm3d_apartments (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  daily_vigor_max bigint NOT NULL DEFAULT 0,
  gifts jsonb NOT NULL DEFAULT '[]'::jsonb,
  ships jsonb NOT NULL DEFAULT '[]'::jsonb,
  gift_daily jsonb NOT NULL DEFAULT '[]'::jsonb,
  gift_permanent jsonb NOT NULL DEFAULT '[]'::jsonb,
  furniture_daily jsonb NOT NULL DEFAULT '[]'::jsonb,
  furniture_permanent jsonb NOT NULL DEFAULT '[]'::jsonb,
  rooms jsonb NOT NULL DEFAULT '[]'::jsonb,
  ins jsonb NOT NULL DEFAULT '[]'::jsonb
);
