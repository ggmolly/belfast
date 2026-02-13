-- 0005_commander_sound_story_table.sql
-- Fix table naming to match legacy Gorm naming strategy.

DO $$
BEGIN
  IF to_regclass('commander_soundstories') IS NOT NULL AND to_regclass('commander_sound_stories') IS NULL THEN
    ALTER TABLE commander_soundstories RENAME TO commander_sound_stories;
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS commander_sound_stories (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  story_id bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (commander_id, story_id)
);

DO $$
BEGIN
  IF to_regclass('commander_soundstories') IS NOT NULL AND to_regclass('commander_sound_stories') IS NOT NULL THEN
    INSERT INTO commander_sound_stories (commander_id, story_id, created_at)
    SELECT commander_id, story_id, created_at
    FROM commander_soundstories
    ON CONFLICT DO NOTHING;
  END IF;
END $$;
