-- 0010_juustagram_player_state.sql
-- Milestone 6: per-player Juustagram state.

CREATE TABLE IF NOT EXISTS juustagram_message_states (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  message_id bigint NOT NULL,
  is_read bigint NOT NULL DEFAULT 0,
  is_good bigint NOT NULL DEFAULT 0,
  good_count bigint NOT NULL DEFAULT 0,
  updated_at bigint NOT NULL DEFAULT 0,
  UNIQUE (commander_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_juus_message_state_commander_id ON juustagram_message_states(commander_id);

CREATE TABLE IF NOT EXISTS juustagram_player_discusses (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  message_id bigint NOT NULL,
  discuss_id bigint NOT NULL,
  option_index bigint NOT NULL,
  npc_reply_id bigint NOT NULL DEFAULT 0,
  comment_time bigint NOT NULL DEFAULT 0,
  UNIQUE (commander_id, message_id, discuss_id)
);

CREATE INDEX IF NOT EXISTS idx_juus_discuss_state_commander_id ON juustagram_player_discusses(commander_id);
