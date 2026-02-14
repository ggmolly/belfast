-- 0001_init.sql
--
-- TODO(M1): Replace this placeholder with a full Postgres schema snapshot.
--
-- Milestone 1 introduces an embedded migration runner and switches the startup
-- path (when enabled) to apply migrations instead of relying on Gorm
-- AutoMigrate.
--
-- Until the full schema is captured here, running with
-- `database.migrations_enabled = true` will fail at runtime once the server
-- attempts to access missing tables.

--
-- Milestone 3: auth/session/audit/authz foundation tables
--
-- sqlc requires a schema to type-check queries. We define the minimum set of
-- tables/constraints needed for Milestone 3 sqlc queries here. The production
-- Postgres schema is expected to already exist in practice.

CREATE TABLE IF NOT EXISTS accounts (
  id text PRIMARY KEY,
  username text,
  username_normalized text UNIQUE,
  commander_id bigint UNIQUE,

  password_hash text NOT NULL,
  password_algo text NOT NULL,
  password_updated_at timestamptz NOT NULL,

  is_admin boolean NOT NULL DEFAULT false,
  disabled_at timestamptz,
  last_login_at timestamptz,

  web_authn_user_handle bytea UNIQUE,

  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
  id text PRIMARY KEY,
  account_id text NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,

  created_at timestamptz NOT NULL,
  last_seen_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,

  ip_address text NOT NULL DEFAULT '',
  user_agent text NOT NULL DEFAULT '',
  revoked_at timestamptz,

  csrf_token text NOT NULL,
  csrf_expires_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_challenges (
  id text PRIMARY KEY,
  user_id text REFERENCES accounts(id) ON DELETE CASCADE,
  type text NOT NULL,
  challenge text NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  metadata jsonb
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id text PRIMARY KEY,

  actor_account_id text REFERENCES accounts(id) ON DELETE SET NULL,
  actor_commander_id bigint,

  method text NOT NULL,
  path text NOT NULL,
  status_code integer NOT NULL,

  permission_key text,
  permission_op text,
  action text,

  metadata jsonb,
  created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
  id text PRIMARY KEY,
  name text NOT NULL UNIQUE,
  description text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  updated_by text
);

CREATE TABLE IF NOT EXISTS permissions (
  id text PRIMARY KEY,
  key text NOT NULL UNIQUE,
  description text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id text NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id text NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

  can_read_self boolean NOT NULL DEFAULT false,
  can_read_any boolean NOT NULL DEFAULT false,
  can_write_self boolean NOT NULL DEFAULT false,
  can_write_any boolean NOT NULL DEFAULT false,

  updated_at timestamptz NOT NULL,

  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS account_roles (
  account_id text NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  role_id text NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL,
  PRIMARY KEY (account_id, role_id)
);

CREATE TABLE IF NOT EXISTS account_permission_overrides (
  account_id text NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  permission_id text NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

  mode text NOT NULL,

  can_read_self boolean NOT NULL DEFAULT false,
  can_read_any boolean NOT NULL DEFAULT false,
  can_write_self boolean NOT NULL DEFAULT false,
  can_write_any boolean NOT NULL DEFAULT false,

  updated_at timestamptz NOT NULL,

  PRIMARY KEY (account_id, permission_id)
);

--
-- Milestone 4: core commander aggregate + inventory/resources/items/mail
--
-- This is a minimal schema slice to support sqlc query type-checking and to
-- allow migrations-enabled Postgres deployments to start up on an empty schema.
-- It is intentionally incomplete; additional tables will be added in later
-- milestones as they are ported.

CREATE TABLE IF NOT EXISTS commanders (
  commander_id bigint PRIMARY KEY,
  account_id bigint NOT NULL,
  level integer NOT NULL DEFAULT 1,
  exp integer NOT NULL DEFAULT 0,
  name text NOT NULL,
  last_login timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  guide_index bigint NOT NULL DEFAULT 0,
  new_guide_index bigint NOT NULL DEFAULT 0,
  name_change_cooldown timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
  room_id bigint NOT NULL DEFAULT 0,
  exchange_count bigint NOT NULL DEFAULT 0,
  draw_count1 bigint NOT NULL DEFAULT 0,
  draw_count10 bigint NOT NULL DEFAULT 0,
  support_requisition_count bigint NOT NULL DEFAULT 0,
  support_requisition_month bigint NOT NULL DEFAULT 0,
  collect_attack_count bigint NOT NULL DEFAULT 0,
  acc_pay_lv bigint NOT NULL DEFAULT 0,
  living_area_cover_id bigint NOT NULL DEFAULT 0,
  selected_icon_frame_id bigint NOT NULL DEFAULT 0,
  selected_chat_frame_id bigint NOT NULL DEFAULT 0,
  selected_battle_ui_id bigint NOT NULL DEFAULT 0,
  display_icon_id bigint NOT NULL DEFAULT 0,
  display_skin_id bigint NOT NULL DEFAULT 0,
  display_icon_theme_id bigint NOT NULL DEFAULT 0,
  manifesto text NOT NULL DEFAULT '',
  dorm_name text NOT NULL DEFAULT '',
  random_ship_mode bigint NOT NULL DEFAULT 0,
  random_flag_ship_enabled boolean NOT NULL DEFAULT false,
  deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS items (
  id bigint PRIMARY KEY,
  name text NOT NULL,
  rarity integer NOT NULL,
  shop_id integer NOT NULL DEFAULT -2,
  type integer NOT NULL,
  virtual_type integer NOT NULL
);

CREATE TABLE IF NOT EXISTS resources (
  id bigint PRIMARY KEY,
  item_id bigint NOT NULL DEFAULT 0,
  name text NOT NULL
);

CREATE TABLE IF NOT EXISTS commander_items (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  item_id bigint NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  count bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (commander_id, item_id)
);

-- Matches the current Gorm model tags (Data participates in the primary key).
CREATE TABLE IF NOT EXISTS commander_misc_items (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  item_id bigint NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  data bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (commander_id, item_id, data)
);

CREATE TABLE IF NOT EXISTS owned_resources (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  resource_id bigint NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  amount bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (commander_id, resource_id)
);

CREATE TABLE IF NOT EXISTS mails (
  id bigint PRIMARY KEY,
  receiver_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  read boolean NOT NULL DEFAULT false,
  date timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  title text NOT NULL,
  body text NOT NULL,
  attachments_collected boolean NOT NULL DEFAULT false,
  is_important boolean NOT NULL DEFAULT false,
  custom_sender text,
  is_archived boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS mail_attachments (
  id bigint PRIMARY KEY,
  mail_id bigint NOT NULL REFERENCES mails(id) ON DELETE CASCADE,
  type bigint NOT NULL,
  item_id bigint NOT NULL,
  quantity bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS owned_skins (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  skin_id bigint NOT NULL,
  expires_at timestamptz,
  PRIMARY KEY (commander_id, skin_id)
);

CREATE TABLE IF NOT EXISTS owned_equipments (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  equipment_id bigint NOT NULL,
  count bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (commander_id, equipment_id)
);

CREATE TABLE IF NOT EXISTS owned_spweapons (
  owner_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  id bigint PRIMARY KEY,
  template_id bigint NOT NULL,
  attr_1 bigint NOT NULL DEFAULT 0,
  attr_2 bigint NOT NULL DEFAULT 0,
  attr_temp_1 bigint NOT NULL DEFAULT 0,
  attr_temp_2 bigint NOT NULL DEFAULT 0,
  effect bigint NOT NULL DEFAULT 0,
  pt bigint NOT NULL DEFAULT 0,
  equipped_ship_id bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ships (
  template_id bigint PRIMARY KEY,
  name text NOT NULL,
  english_name text NOT NULL,
  rarity_id bigint NOT NULL,
  star bigint NOT NULL,
  type bigint NOT NULL,
  nationality bigint NOT NULL,
  build_time bigint NOT NULL,
  pool_id bigint
);

CREATE TABLE IF NOT EXISTS owned_ships (
  owner_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_id bigint NOT NULL REFERENCES ships(template_id) ON DELETE CASCADE,
  id bigint PRIMARY KEY,
  level bigint NOT NULL DEFAULT 1,
  exp bigint NOT NULL DEFAULT 0,
  surplus_exp bigint NOT NULL DEFAULT 0,
  max_level bigint NOT NULL DEFAULT 50,
  intimacy bigint NOT NULL DEFAULT 5000,
  is_locked boolean NOT NULL DEFAULT false,
  propose boolean NOT NULL DEFAULT false,
  common_flag boolean NOT NULL DEFAULT false,
  blueprint_flag boolean NOT NULL DEFAULT false,
  proficiency boolean NOT NULL DEFAULT false,
  activity_npc bigint NOT NULL DEFAULT 0,
  custom_name text NOT NULL DEFAULT '',
  change_name_timestamp timestamptz NOT NULL DEFAULT '1970-01-01 01:00:00+00',
  create_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  energy bigint NOT NULL DEFAULT 150,
  state bigint NOT NULL DEFAULT 0,
  state_info1 bigint NOT NULL DEFAULT 0,
  state_info2 bigint NOT NULL DEFAULT 0,
  state_info3 bigint NOT NULL DEFAULT 0,
  state_info4 bigint NOT NULL DEFAULT 0,
  skin_id bigint NOT NULL DEFAULT 0,
  is_secretary boolean NOT NULL DEFAULT false,
  secretary_position bigint,
  secretary_phantom_id bigint NOT NULL DEFAULT 0,
  deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS owned_ship_equipments (
  owner_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_id bigint NOT NULL REFERENCES owned_ships(id) ON DELETE CASCADE,
  pos bigint NOT NULL,
  equip_id bigint NOT NULL,
  skin_id bigint NOT NULL,
  PRIMARY KEY (owner_id, ship_id, pos)
);

CREATE TABLE IF NOT EXISTS owned_ship_strengths (
  owner_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_id bigint NOT NULL REFERENCES owned_ships(id) ON DELETE CASCADE,
  strength_id bigint NOT NULL,
  exp bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (owner_id, ship_id, strength_id)
);

CREATE TABLE IF NOT EXISTS fleets (
  id bigint PRIMARY KEY,
  game_id bigint NOT NULL,
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  name text NOT NULL,
  ship_list jsonb NOT NULL DEFAULT '[]'::jsonb,
  meowfficer_list jsonb NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS builds (
  id bigint PRIMARY KEY,
  builder_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_id bigint NOT NULL REFERENCES ships(template_id) ON DELETE CASCADE,
  pool_id bigint NOT NULL,
  finishes_at timestamptz NOT NULL
);

--
-- Milestone 5: bulk game-data importer tables (items/ships already above)
--

CREATE TABLE IF NOT EXISTS buffs (
  id bigint PRIMARY KEY,
  name text NOT NULL,
  description text NOT NULL,
  max_time integer NOT NULL DEFAULT 0,
  benefit_type text NOT NULL
);

CREATE TABLE IF NOT EXISTS skins (
  id bigint PRIMARY KEY,
  name text NOT NULL,
  ship_group bigint NOT NULL,
  "desc" text,
  bg text,
  bg_sp text,
  bgm text,
  painting text,
  prefab text,
  change_skin jsonb,
  show_skin text,
  skeleton_skin text,
  ship_l2_d_id jsonb,
  l2_d_animations jsonb,
  l2_d_drag_rate jsonb,
  l2_d_para_range jsonb,
  l2_dse jsonb,
  l2_d_voice_calib jsonb,
  part_scale text,
  main_ui_fx text,
  spine_offset jsonb,
  spine_profile jsonb,
  tag jsonb,
  time jsonb,
  get_showing jsonb,
  purchase_offset jsonb,
  shop_offset jsonb,
  rarity_bg text,
  special_effects jsonb,
  group_index integer,
  gyro integer,
  hand_id integer,
  illustrator integer,
  illustrator2 integer,
  voice_actor integer,
  voice_actor2 integer,
  double_char integer,
  lip_smoothing integer,
  lip_sync_gain integer,
  l2_d_ignore_drag integer,
  skin_type integer,
  shop_id integer,
  shop_type_id integer,
  shop_dynamic_hx integer,
  spine_action jsonb,
  spine_use_live2_d integer,
  live2_d_offset jsonb,
  live2_d_profile jsonb,
  fx_container jsonb,
  bound_bone jsonb,
  smoke jsonb
);

CREATE TABLE IF NOT EXISTS weapons (
  id bigint PRIMARY KEY,
  action_index text NOT NULL,
  aim_type integer NOT NULL,
  angle integer NOT NULL,
  attack_attribute integer NOT NULL,
  attack_attribute_ratio integer NOT NULL,
  auto_aftercast jsonb,
  axis_angle integer NOT NULL,
  barrage_id jsonb,
  bullet_id jsonb,
  charge_param jsonb,
  corrected integer NOT NULL,
  damage integer NOT NULL,
  effect_move integer NOT NULL,
  expose integer NOT NULL,
  fire_fx text,
  fire_fx_loop_type integer NOT NULL,
  fire_sfx text,
  initial_over_heat integer NOT NULL,
  min_range integer NOT NULL,
  oxy_type jsonb,
  precast_param jsonb,
  queue integer NOT NULL,
  range integer NOT NULL,
  recover_time jsonb,
  reload_max integer NOT NULL,
  search_condition jsonb,
  search_type integer NOT NULL,
  shake_screen integer NOT NULL,
  spawn_bound jsonb,
  suppress integer NOT NULL,
  torpedo_ammo integer NOT NULL,
  type integer NOT NULL
);

CREATE TABLE IF NOT EXISTS equipments (
  id bigint PRIMARY KEY,
  base bigint,
  destroy_gold bigint NOT NULL,
  destroy_item jsonb,
  equip_limit integer NOT NULL,
  "group" bigint NOT NULL,
  important bigint NOT NULL,
  level bigint NOT NULL,
  next bigint NOT NULL,
  prev bigint NOT NULL,
  restore_gold bigint NOT NULL,
  restore_item jsonb,
  ship_type_forbidden jsonb,
  trans_use_gold bigint NOT NULL,
  trans_use_item jsonb,
  type bigint NOT NULL,
  upgrade_formula_id jsonb
);

CREATE TABLE IF NOT EXISTS skills (
  id bigint PRIMARY KEY,
  name text NOT NULL,
  "desc" text,
  cd bigint NOT NULL,
  painting jsonb,
  picture text,
  ani_effect jsonb,
  ui_effect text,
  effect_list jsonb
);

CREATE TABLE IF NOT EXISTS config_entries (
  id bigserial PRIMARY KEY,
  category text NOT NULL,
  key text NOT NULL,
  data jsonb NOT NULL,
  UNIQUE (category, key)
);

CREATE TABLE IF NOT EXISTS requisition_ships (
  ship_id bigint PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS shop_offers (
  id bigint PRIMARY KEY,
  effects jsonb NOT NULL,
  effect_args jsonb,
  number integer NOT NULL,
  resource_number integer NOT NULL,
  resource_id bigint NOT NULL,
  type bigint NOT NULL,
  genre text NOT NULL,
  discount integer NOT NULL
);

CREATE TABLE IF NOT EXISTS juustagram_templates (
  id bigint PRIMARY KEY,
  group_id bigint NOT NULL,
  ship_group bigint NOT NULL,
  name text NOT NULL,
  sculpture text NOT NULL,
  picture_persist text NOT NULL,
  message_persist text NOT NULL,
  is_active bigint NOT NULL DEFAULT 0,
  npc_discuss_persist text NOT NULL DEFAULT '[]',
  time text NOT NULL DEFAULT '[]',
  time_persist text NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS juustagram_npc_templates (
  id bigint PRIMARY KEY,
  ship_group bigint NOT NULL,
  message_persist text NOT NULL,
  npc_reply_persist text NOT NULL DEFAULT '[]',
  time_persist text NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS juustagram_languages (
  key text PRIMARY KEY,
  value text NOT NULL
);

CREATE TABLE IF NOT EXISTS juustagram_ship_group_templates (
  ship_group bigint PRIMARY KEY,
  name text NOT NULL,
  background text NOT NULL,
  sculpture text NOT NULL,
  sculpture_ii text NOT NULL,
  nationality bigint NOT NULL,
  type bigint NOT NULL
);
