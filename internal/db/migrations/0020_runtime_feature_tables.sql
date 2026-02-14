-- 0020_runtime_feature_tables.sql

CREATE TABLE IF NOT EXISTS activity_fleets (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  activity_id bigint NOT NULL,
  group_list jsonb NOT NULL DEFAULT '[]'::jsonb,
  PRIMARY KEY (commander_id, activity_id)
);

CREATE TABLE IF NOT EXISTS backyard_custom_theme_templates (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  pos bigint NOT NULL,
  name text NOT NULL,
  furniture_put_list jsonb NOT NULL DEFAULT '[]'::jsonb,
  icon_image_md5 text NOT NULL DEFAULT '',
  image_md5 text NOT NULL DEFAULT '',
  upload_time bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (commander_id, pos)
);

CREATE TABLE IF NOT EXISTS backyard_published_theme_versions (
  theme_id text NOT NULL,
  upload_time bigint NOT NULL,
  owner_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  pos bigint NOT NULL,
  name text NOT NULL,
  furniture_put_list jsonb NOT NULL DEFAULT '[]'::jsonb,
  icon_image_md5 text NOT NULL DEFAULT '',
  image_md5 text NOT NULL DEFAULT '',
  like_count bigint NOT NULL DEFAULT 0,
  fav_count bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (theme_id, upload_time)
);
CREATE INDEX IF NOT EXISTS idx_backyard_published_theme_versions_theme_id ON backyard_published_theme_versions(theme_id);

CREATE TABLE IF NOT EXISTS backyard_theme_collections (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  theme_id text NOT NULL,
  upload_time bigint NOT NULL,
  PRIMARY KEY (commander_id, theme_id, upload_time)
);
CREATE INDEX IF NOT EXISTS idx_backyard_theme_collections_commander_upload ON backyard_theme_collections(commander_id, upload_time DESC);

CREATE TABLE IF NOT EXISTS backyard_theme_likes (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  theme_id text NOT NULL,
  upload_time bigint NOT NULL,
  PRIMARY KEY (commander_id, theme_id, upload_time)
);

CREATE TABLE IF NOT EXISTS backyard_theme_informs (
  id bigserial PRIMARY KEY,
  reporter_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  target_id bigint NOT NULL,
  target_name text NOT NULL,
  theme_id text NOT NULL,
  theme_name text NOT NULL,
  reason bigint NOT NULL,
  created_at bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS battle_sessions (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  system bigint NOT NULL,
  stage_id bigint NOT NULL,
  key bigint NOT NULL,
  ship_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chapter_drops (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  chapter_id bigint NOT NULL,
  ship_id bigint NOT NULL,
  PRIMARY KEY (commander_id, chapter_id, ship_id)
);

CREATE TABLE IF NOT EXISTS chapter_progress (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  chapter_id bigint NOT NULL,
  progress bigint NOT NULL DEFAULT 0,
  kill_boss_count bigint NOT NULL DEFAULT 0,
  kill_enemy_count bigint NOT NULL DEFAULT 0,
  take_box_count bigint NOT NULL DEFAULT 0,
  defeat_count bigint NOT NULL DEFAULT 0,
  today_defeat_count bigint NOT NULL DEFAULT 0,
  pass_count bigint NOT NULL DEFAULT 0,
  updated_at bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (commander_id, chapter_id)
);

CREATE TABLE IF NOT EXISTS chapter_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  chapter_id bigint NOT NULL,
  state bytea NOT NULL,
  updated_at bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS commander_tbs (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  state bytea NOT NULL,
  permanent bytea NOT NULL
);

CREATE TABLE IF NOT EXISTS commander_trophy_progresses (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  trophy_id bigint NOT NULL,
  progress bigint NOT NULL DEFAULT 0,
  timestamp bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (commander_id, trophy_id)
);

CREATE TABLE IF NOT EXISTS equip_code_shares (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_group_id bigint NOT NULL,
  share_day bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (commander_id, ship_group_id, share_day)
);
CREATE INDEX IF NOT EXISTS idx_equip_code_shares_commander_day ON equip_code_shares(commander_id, share_day);

CREATE TABLE IF NOT EXISTS equip_code_likes (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_group_id bigint NOT NULL,
  share_id bigint NOT NULL,
  like_day bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (commander_id, ship_group_id, share_id, like_day)
);
CREATE INDEX IF NOT EXISTS idx_equip_code_likes_share_id ON equip_code_likes(share_id);

CREATE TABLE IF NOT EXISTS equip_code_reports (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  share_id bigint NOT NULL,
  report_day bigint NOT NULL,
  ship_group_id bigint NOT NULL,
  report_type bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (commander_id, share_id, report_day)
);
CREATE INDEX IF NOT EXISTS idx_equip_code_reports_share_id ON equip_code_reports(share_id);

CREATE TABLE IF NOT EXISTS exchange_codes (
  id bigserial PRIMARY KEY,
  code text NOT NULL UNIQUE,
  platform text NOT NULL DEFAULT '',
  quota bigint NOT NULL DEFAULT -1,
  rewards jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS exchange_code_redeems (
  exchange_code_id bigint NOT NULL REFERENCES exchange_codes(id) ON DELETE CASCADE,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  redeemed_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (exchange_code_id, commander_id)
);
CREATE INDEX IF NOT EXISTS idx_exchange_code_redeems_exchange_code_id_redeemed_at ON exchange_code_redeems(exchange_code_id, redeemed_at DESC);

CREATE TABLE IF NOT EXISTS exercise_fleets (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  vanguard_ship_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
  main_ship_ids jsonb NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS global_skin_restrictions (
  skin_id bigint PRIMARY KEY REFERENCES skins(id) ON DELETE CASCADE,
  type bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS global_skin_restriction_windows (
  id bigint PRIMARY KEY,
  skin_id bigint NOT NULL REFERENCES skins(id) ON DELETE CASCADE,
  type bigint NOT NULL,
  start_time bigint NOT NULL,
  stop_time bigint NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_global_skin_restriction_windows_skin_id ON global_skin_restriction_windows(skin_id);

CREATE TABLE IF NOT EXISTS local_accounts (
  arg2 bigint PRIMARY KEY,
  account text NOT NULL UNIQUE,
  password text NOT NULL,
  mail_box text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS owned_ship_transforms (
  owner_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_id bigint NOT NULL REFERENCES owned_ships(id) ON DELETE CASCADE,
  transform_id bigint NOT NULL,
  level bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (owner_id, ship_id, transform_id)
);

CREATE TABLE IF NOT EXISTS rarities (
  id bigint PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE IF NOT EXISTS reflux_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  active bigint NOT NULL DEFAULT 0,
  return_lv bigint NOT NULL DEFAULT 0,
  return_time bigint NOT NULL DEFAULT 0,
  ship_number bigint NOT NULL DEFAULT 0,
  last_offline_time bigint NOT NULL DEFAULT 0,
  pt bigint NOT NULL DEFAULT 0,
  sign_cnt bigint NOT NULL DEFAULT 0,
  sign_last_time bigint NOT NULL DEFAULT 0,
  pt_stage bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS remaster_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ticket_count bigint NOT NULL DEFAULT 0,
  active_chapter_id bigint NOT NULL DEFAULT 0,
  daily_count bigint NOT NULL DEFAULT 0,
  last_daily_reset_at timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS remaster_progresses (
  id bigserial PRIMARY KEY,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  chapter_id bigint NOT NULL,
  pos bigint NOT NULL,
  count bigint NOT NULL DEFAULT 0,
  received boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (commander_id, chapter_id, pos)
);

CREATE TABLE IF NOT EXISTS secondary_password_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  password_hash text NOT NULL DEFAULT '',
  notice text NOT NULL DEFAULT '',
  system_list jsonb NOT NULL DEFAULT '[]'::jsonb,
  state bigint NOT NULL DEFAULT 0,
  fail_count bigint NOT NULL DEFAULT 0,
  fail_cd bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ship_types (
  id bigint PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE IF NOT EXISTS submarine_expedition_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  last_refresh_time bigint NOT NULL DEFAULT 0,
  weekly_refresh_count bigint NOT NULL DEFAULT 0,
  active_chapter_id bigint NOT NULL DEFAULT 0,
  overall_progress bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS survey_states (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  survey_id bigint NOT NULL DEFAULT 0,
  completed_at timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS web_authn_credentials (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  credential_id text NOT NULL UNIQUE,
  public_key bytea NOT NULL,
  sign_count bigint NOT NULL DEFAULT 0,
  transports jsonb NOT NULL DEFAULT '[]'::jsonb,
  aaguid text NOT NULL DEFAULT '',
  attestation_fmt text NOT NULL DEFAULT '',
  resident_key text NOT NULL DEFAULT '',
  backup_eligible boolean,
  backup_state boolean,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at timestamptz,
  label text,
  rp_id text NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_web_authn_credentials_user_id ON web_authn_credentials(user_id);
