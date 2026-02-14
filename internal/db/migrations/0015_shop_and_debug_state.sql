-- 0015_shop_and_debug_state.sql

CREATE TABLE IF NOT EXISTS arena_shop_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  flash_count bigint NOT NULL,
  last_refresh_time bigint NOT NULL,
  next_flash_time bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS shopping_street_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  level bigint NOT NULL,
  next_flash_time bigint NOT NULL,
  level_up_time bigint NOT NULL,
  flash_count bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS shopping_street_goods (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  goods_id bigint NOT NULL,
  discount bigint NOT NULL,
  buy_count bigint NOT NULL,
  PRIMARY KEY (commander_id, goods_id)
);

CREATE TABLE IF NOT EXISTS medal_shop_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  next_refresh_time bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS medal_shop_goods (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  index bigint NOT NULL,
  goods_id bigint NOT NULL,
  count bigint NOT NULL,
  PRIMARY KEY (commander_id, index)
);

CREATE TABLE IF NOT EXISTS mini_game_shop_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  next_refresh_time bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS mini_game_shop_goods (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  goods_id bigint NOT NULL,
  count bigint NOT NULL,
  PRIMARY KEY (commander_id, goods_id)
);

CREATE TABLE IF NOT EXISTS guild_shop_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  refresh_count bigint NOT NULL,
  next_refresh_time bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS guild_shop_goods (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  index bigint NOT NULL,
  goods_id bigint NOT NULL,
  count bigint NOT NULL,
  PRIMARY KEY (commander_id, index)
);

CREATE TABLE IF NOT EXISTS debugs (
  frame_id bigserial PRIMARY KEY,
  packet_size bigint NOT NULL,
  packet_id bigint NOT NULL,
  data bytea NOT NULL,
  logged_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
