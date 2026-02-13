-- 0002_missing_tables.sql
-- Milestone 6: add tables previously created via Gorm AutoMigrate.

CREATE TABLE IF NOT EXISTS punishments (
  id bigserial PRIMARY KEY,
  punished_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  lift_timestamp timestamptz,
  is_permanent boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_punishments_punished_id ON punishments (punished_id);

CREATE TABLE IF NOT EXISTS likes (
  group_id bigint NOT NULL,
  liker_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  PRIMARY KEY (group_id, liker_id)
);

CREATE TABLE IF NOT EXISTS messages (
  id bigserial PRIMARY KEY,
  sender_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  room_id bigint NOT NULL,
  sent_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  content varchar(512) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_room_id_sent_at ON messages (room_id, sent_at DESC);

CREATE TABLE IF NOT EXISTS commander_attires (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  type bigint NOT NULL,
  attire_id bigint NOT NULL,
  expires_at timestamptz,
  is_new boolean NOT NULL DEFAULT false,
  PRIMARY KEY (commander_id, type, attire_id)
);

CREATE TABLE IF NOT EXISTS commander_buffs (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  buff_id bigint NOT NULL,
  expires_at timestamptz NOT NULL,
  PRIMARY KEY (commander_id, buff_id)
);

CREATE INDEX IF NOT EXISTS idx_commander_buffs_expires_at ON commander_buffs (expires_at);

CREATE TABLE IF NOT EXISTS commander_dorm_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  level bigint NOT NULL DEFAULT 1,
  food bigint NOT NULL DEFAULT 0,
  food_max_increase_count bigint NOT NULL DEFAULT 0,
  food_max_increase bigint NOT NULL DEFAULT 0,
  floor_num bigint NOT NULL DEFAULT 1,
  exp_pos bigint NOT NULL DEFAULT 2,
  next_timestamp bigint NOT NULL DEFAULT 0,
  load_exp bigint NOT NULL DEFAULT 0,
  load_food bigint NOT NULL DEFAULT 0,
  load_time bigint NOT NULL DEFAULT 0,
  updated_at_unix_timestamp bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS commander_appreciation_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  music_no bigint NOT NULL DEFAULT 0,
  music_mode bigint NOT NULL DEFAULT 0,
  cartoon_read_mark text NOT NULL DEFAULT '[]',
  cartoon_collect_mark text NOT NULL DEFAULT '[]',
  gallery_unlocks text NOT NULL DEFAULT '[]',
  gallery_favor_ids text NOT NULL DEFAULT '[]',
  music_favor_ids text NOT NULL DEFAULT '[]'
);
