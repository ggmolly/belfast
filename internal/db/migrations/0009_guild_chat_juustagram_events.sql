-- 0009_guild_chat_juustagram_events.sql
-- Milestone 6: tables required for packet handlers.

CREATE TABLE IF NOT EXISTS guild_chat_messages (
  id bigserial PRIMARY KEY,
  guild_id bigint NOT NULL,
  sender_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  sent_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  content text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_guild_chat_time ON guild_chat_messages(guild_id, sent_at DESC);

CREATE TABLE IF NOT EXISTS juustagram_groups (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  group_id bigint NOT NULL,
  skin_id bigint NOT NULL DEFAULT 0,
  favorite bigint NOT NULL DEFAULT 0,
  cur_chat_group bigint NOT NULL DEFAULT 0,
  UNIQUE (commander_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_juus_group_commander ON juustagram_groups(commander_id);

CREATE TABLE IF NOT EXISTS juustagram_chat_groups (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  group_record_id bigint NOT NULL REFERENCES juustagram_groups(id) ON DELETE CASCADE,
  chat_group_id bigint NOT NULL,
  op_time bigint NOT NULL DEFAULT 0,
  read_flag bigint NOT NULL DEFAULT 0,
  UNIQUE (commander_id, chat_group_id)
);

CREATE INDEX IF NOT EXISTS idx_juus_chat_group_commander ON juustagram_chat_groups(commander_id);
CREATE INDEX IF NOT EXISTS idx_juus_chat_group_group_record_id ON juustagram_chat_groups(group_record_id);

CREATE TABLE IF NOT EXISTS juustagram_replies (
  id bigserial PRIMARY KEY,
  chat_group_record_id bigint NOT NULL REFERENCES juustagram_chat_groups(id) ON DELETE CASCADE,
  sequence bigint NOT NULL,
  key bigint NOT NULL,
  value bigint NOT NULL,
  UNIQUE (chat_group_record_id, sequence)
);

CREATE INDEX IF NOT EXISTS idx_juus_reply_chat_group_record_id ON juustagram_replies(chat_group_record_id);

CREATE TABLE IF NOT EXISTS event_collections (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  collection_id bigint NOT NULL,
  start_time bigint NOT NULL DEFAULT 0,
  finish_time bigint NOT NULL DEFAULT 0,
  ship_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
  PRIMARY KEY (commander_id, collection_id)
);
