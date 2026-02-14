-- 0004_commander_misc_tables.sql
-- Milestone 6: commander-scoped tables used by packet handlers.

CREATE TABLE IF NOT EXISTS commander_common_flags (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  flag_id bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (commander_id, flag_id)
);

CREATE TABLE IF NOT EXISTS commander_dorm_floor_layouts (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  floor bigint NOT NULL,
  furniture_put_list jsonb NOT NULL,
  PRIMARY KEY (commander_id, floor)
);

CREATE TABLE IF NOT EXISTS commander_dorm_themes (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  theme_slot_id bigint NOT NULL,
  name text NOT NULL DEFAULT '',
  furniture_put_list jsonb NOT NULL,
  PRIMARY KEY (commander_id, theme_slot_id)
);

CREATE TABLE IF NOT EXISTS commander_furnitures (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  furniture_id bigint NOT NULL,
  count bigint NOT NULL,
  get_time bigint NOT NULL,
  PRIMARY KEY (commander_id, furniture_id)
);

CREATE TABLE IF NOT EXISTS commander_living_area_covers (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  cover_id bigint NOT NULL,
  unlocked_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_new boolean NOT NULL DEFAULT false,
  PRIMARY KEY (commander_id, cover_id)
);

CREATE TABLE IF NOT EXISTS commander_medal_displays (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  position bigint NOT NULL,
  medal_id bigint NOT NULL,
  PRIMARY KEY (commander_id, position)
);

CREATE TABLE IF NOT EXISTS commander_soundstories (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  story_id bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (commander_id, story_id)
);

CREATE TABLE IF NOT EXISTS commander_stories (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  story_id bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (commander_id, story_id)
);
