-- 0023_love_letter_states.sql

CREATE TABLE IF NOT EXISTS commander_love_letter_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  medals jsonb NOT NULL DEFAULT '[]'::jsonb,
  manual_letters jsonb NOT NULL DEFAULT '[]'::jsonb,
  converted_items jsonb NOT NULL DEFAULT '[]'::jsonb,
  rewarded_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
  letter_contents jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
