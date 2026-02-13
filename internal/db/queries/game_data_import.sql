-- Game data bulk importer queries (Milestone 5)

-- name: UpsertItem :exec
INSERT INTO items (
  id,
  name,
  rarity,
  shop_id,
  type,
  virtual_type
) VALUES (
  $1, $2, $3, $4, $5, $6
)
ON CONFLICT (id)
DO UPDATE SET
  name = EXCLUDED.name,
  rarity = EXCLUDED.rarity,
  shop_id = EXCLUDED.shop_id,
  type = EXCLUDED.type,
  virtual_type = EXCLUDED.virtual_type;

-- name: UpsertBuff :exec
INSERT INTO buffs (
  id,
  name,
  description,
  max_time,
  benefit_type
) VALUES (
  $1, $2, $3, $4, $5
)
ON CONFLICT (id)
DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  max_time = EXCLUDED.max_time,
  benefit_type = EXCLUDED.benefit_type;

-- name: UpsertShip :exec
INSERT INTO ships (
  template_id,
  name,
  english_name,
  rarity_id,
  star,
  type,
  nationality,
  build_time,
  pool_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (template_id)
DO UPDATE SET
  name = EXCLUDED.name,
  english_name = EXCLUDED.english_name,
  rarity_id = EXCLUDED.rarity_id,
  star = EXCLUDED.star,
  type = EXCLUDED.type,
  nationality = EXCLUDED.nationality,
  build_time = EXCLUDED.build_time,
  pool_id = EXCLUDED.pool_id;

-- name: SetShipPoolID :execresult
UPDATE ships
SET pool_id = $2
WHERE template_id = $1;

-- name: SetShipBuildTime :execresult
UPDATE ships
SET build_time = $2
WHERE template_id = $1;

-- name: UpsertSkin :exec
INSERT INTO skins (
  id,
  name,
  ship_group,
  "desc",
  bg,
  bg_sp,
  bgm,
  painting,
  prefab,
  change_skin,
  show_skin,
  skeleton_skin,
  ship_l2_d_id,
  l2_d_animations,
  l2_d_drag_rate,
  l2_d_para_range,
  l2_dse,
  l2_d_voice_calib,
  part_scale,
  main_ui_fx,
  spine_offset,
  spine_profile,
  tag,
  "time",
  get_showing,
  purchase_offset,
  shop_offset,
  rarity_bg,
  special_effects,
  group_index,
  gyro,
  hand_id,
  illustrator,
  illustrator2,
  voice_actor,
  voice_actor2,
  double_char,
  lip_smoothing,
  lip_sync_gain,
  l2_d_ignore_drag,
  skin_type,
  shop_id,
  shop_type_id,
  shop_dynamic_hx,
  spine_action,
  spine_use_live2_d,
  live2_d_offset,
  live2_d_profile,
  fx_container,
  bound_bone,
  smoke
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
  $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
  $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
  $31, $32, $33, $34, $35, $36, $37, $38, $39, $40,
  $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51
)
ON CONFLICT (id)
DO UPDATE SET
  name = EXCLUDED.name,
  ship_group = EXCLUDED.ship_group,
  "desc" = EXCLUDED."desc",
  bg = EXCLUDED.bg,
  bg_sp = EXCLUDED.bg_sp,
  bgm = EXCLUDED.bgm,
  painting = EXCLUDED.painting,
  prefab = EXCLUDED.prefab,
  change_skin = EXCLUDED.change_skin,
  show_skin = EXCLUDED.show_skin,
  skeleton_skin = EXCLUDED.skeleton_skin,
  ship_l2_d_id = EXCLUDED.ship_l2_d_id,
  l2_d_animations = EXCLUDED.l2_d_animations,
  l2_d_drag_rate = EXCLUDED.l2_d_drag_rate,
  l2_d_para_range = EXCLUDED.l2_d_para_range,
  l2_dse = EXCLUDED.l2_dse,
  l2_d_voice_calib = EXCLUDED.l2_d_voice_calib,
  part_scale = EXCLUDED.part_scale,
  main_ui_fx = EXCLUDED.main_ui_fx,
  spine_offset = EXCLUDED.spine_offset,
  spine_profile = EXCLUDED.spine_profile,
  tag = EXCLUDED.tag,
  "time" = EXCLUDED."time",
  get_showing = EXCLUDED.get_showing,
  purchase_offset = EXCLUDED.purchase_offset,
  shop_offset = EXCLUDED.shop_offset,
  rarity_bg = EXCLUDED.rarity_bg,
  special_effects = EXCLUDED.special_effects,
  group_index = EXCLUDED.group_index,
  gyro = EXCLUDED.gyro,
  hand_id = EXCLUDED.hand_id,
  illustrator = EXCLUDED.illustrator,
  illustrator2 = EXCLUDED.illustrator2,
  voice_actor = EXCLUDED.voice_actor,
  voice_actor2 = EXCLUDED.voice_actor2,
  double_char = EXCLUDED.double_char,
  lip_smoothing = EXCLUDED.lip_smoothing,
  lip_sync_gain = EXCLUDED.lip_sync_gain,
  l2_d_ignore_drag = EXCLUDED.l2_d_ignore_drag,
  skin_type = EXCLUDED.skin_type,
  shop_id = EXCLUDED.shop_id,
  shop_type_id = EXCLUDED.shop_type_id,
  shop_dynamic_hx = EXCLUDED.shop_dynamic_hx,
  spine_action = EXCLUDED.spine_action,
  spine_use_live2_d = EXCLUDED.spine_use_live2_d,
  live2_d_offset = EXCLUDED.live2_d_offset,
  live2_d_profile = EXCLUDED.live2_d_profile,
  fx_container = EXCLUDED.fx_container,
  bound_bone = EXCLUDED.bound_bone,
  smoke = EXCLUDED.smoke;

-- name: UpsertResource :exec
INSERT INTO resources (
  id,
  item_id,
  name
) VALUES (
  $1, $2, $3
)
ON CONFLICT (id)
DO UPDATE SET
  item_id = EXCLUDED.item_id,
  name = EXCLUDED.name;

-- name: UpsertRequisitionShip :exec
INSERT INTO requisition_ships (
  ship_id
) VALUES (
  $1
)
ON CONFLICT (ship_id)
DO NOTHING;

-- name: UpsertShopOffer :exec
INSERT INTO shop_offers (
  id,
  effects,
  effect_args,
  number,
  resource_number,
  resource_id,
  type,
  genre,
  discount
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (id)
DO UPDATE SET
  effects = EXCLUDED.effects,
  effect_args = EXCLUDED.effect_args,
  number = EXCLUDED.number,
  resource_number = EXCLUDED.resource_number,
  resource_id = EXCLUDED.resource_id,
  type = EXCLUDED.type,
  genre = EXCLUDED.genre,
  discount = EXCLUDED.discount;

-- name: UpsertWeapon :exec
INSERT INTO weapons (
  id,
  action_index,
  aim_type,
  angle,
  attack_attribute,
  attack_attribute_ratio,
  auto_aftercast,
  axis_angle,
  barrage_id,
  bullet_id,
  charge_param,
  corrected,
  damage,
  effect_move,
  expose,
  fire_fx,
  fire_fx_loop_type,
  fire_sfx,
  initial_over_heat,
  min_range,
  oxy_type,
  precast_param,
  queue,
  range,
  recover_time,
  reload_max,
  search_condition,
  search_type,
  shake_screen,
  spawn_bound,
  suppress,
  torpedo_ammo,
  type
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
  $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22,
  $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33
)
ON CONFLICT (id)
DO UPDATE SET
  action_index = EXCLUDED.action_index,
  aim_type = EXCLUDED.aim_type,
  angle = EXCLUDED.angle,
  attack_attribute = EXCLUDED.attack_attribute,
  attack_attribute_ratio = EXCLUDED.attack_attribute_ratio,
  auto_aftercast = EXCLUDED.auto_aftercast,
  axis_angle = EXCLUDED.axis_angle,
  barrage_id = EXCLUDED.barrage_id,
  bullet_id = EXCLUDED.bullet_id,
  charge_param = EXCLUDED.charge_param,
  corrected = EXCLUDED.corrected,
  damage = EXCLUDED.damage,
  effect_move = EXCLUDED.effect_move,
  expose = EXCLUDED.expose,
  fire_fx = EXCLUDED.fire_fx,
  fire_fx_loop_type = EXCLUDED.fire_fx_loop_type,
  fire_sfx = EXCLUDED.fire_sfx,
  initial_over_heat = EXCLUDED.initial_over_heat,
  min_range = EXCLUDED.min_range,
  oxy_type = EXCLUDED.oxy_type,
  precast_param = EXCLUDED.precast_param,
  queue = EXCLUDED.queue,
  range = EXCLUDED.range,
  recover_time = EXCLUDED.recover_time,
  reload_max = EXCLUDED.reload_max,
  search_condition = EXCLUDED.search_condition,
  search_type = EXCLUDED.search_type,
  shake_screen = EXCLUDED.shake_screen,
  spawn_bound = EXCLUDED.spawn_bound,
  suppress = EXCLUDED.suppress,
  torpedo_ammo = EXCLUDED.torpedo_ammo,
  type = EXCLUDED.type;

-- name: UpsertEquipment :exec
INSERT INTO equipments (
  id,
  base,
  destroy_gold,
  destroy_item,
  equip_limit,
  "group",
  important,
  level,
  next,
  prev,
  restore_gold,
  restore_item,
  ship_type_forbidden,
  trans_use_gold,
  trans_use_item,
  type,
  upgrade_formula_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
ON CONFLICT (id)
DO UPDATE SET
  base = EXCLUDED.base,
  destroy_gold = EXCLUDED.destroy_gold,
  destroy_item = EXCLUDED.destroy_item,
  equip_limit = EXCLUDED.equip_limit,
  "group" = EXCLUDED."group",
  important = EXCLUDED.important,
  level = EXCLUDED.level,
  next = EXCLUDED.next,
  prev = EXCLUDED.prev,
  restore_gold = EXCLUDED.restore_gold,
  restore_item = EXCLUDED.restore_item,
  ship_type_forbidden = EXCLUDED.ship_type_forbidden,
  trans_use_gold = EXCLUDED.trans_use_gold,
  trans_use_item = EXCLUDED.trans_use_item,
  type = EXCLUDED.type,
  upgrade_formula_id = EXCLUDED.upgrade_formula_id;

-- name: UpsertSkill :exec
INSERT INTO skills (
  id,
  name,
  "desc",
  cd,
  painting,
  picture,
  ani_effect,
  ui_effect,
  effect_list
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (id)
DO UPDATE SET
  name = EXCLUDED.name,
  "desc" = EXCLUDED."desc",
  cd = EXCLUDED.cd,
  painting = EXCLUDED.painting,
  picture = EXCLUDED.picture,
  ani_effect = EXCLUDED.ani_effect,
  ui_effect = EXCLUDED.ui_effect,
  effect_list = EXCLUDED.effect_list;

-- name: UpsertConfigEntry :exec
INSERT INTO config_entries (
  category,
  key,
  data
) VALUES (
  $1, $2, $3
)
ON CONFLICT (category, key)
DO UPDATE SET
  data = EXCLUDED.data;

-- name: UpsertJuustagramTemplate :exec
INSERT INTO juustagram_templates (
  id,
  group_id,
  ship_group,
  name,
  sculpture,
  picture_persist,
  message_persist,
  is_active,
  npc_discuss_persist,
  time,
  time_persist
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (id)
DO UPDATE SET
  group_id = EXCLUDED.group_id,
  ship_group = EXCLUDED.ship_group,
  name = EXCLUDED.name,
  sculpture = EXCLUDED.sculpture,
  picture_persist = EXCLUDED.picture_persist,
  message_persist = EXCLUDED.message_persist,
  is_active = EXCLUDED.is_active,
  npc_discuss_persist = EXCLUDED.npc_discuss_persist,
  time = EXCLUDED.time,
  time_persist = EXCLUDED.time_persist;

-- name: UpsertJuustagramNpcTemplate :exec
INSERT INTO juustagram_npc_templates (
  id,
  ship_group,
  message_persist,
  npc_reply_persist,
  time_persist
) VALUES (
  $1, $2, $3, $4, $5
)
ON CONFLICT (id)
DO UPDATE SET
  ship_group = EXCLUDED.ship_group,
  message_persist = EXCLUDED.message_persist,
  npc_reply_persist = EXCLUDED.npc_reply_persist,
  time_persist = EXCLUDED.time_persist;

-- name: UpsertJuustagramLanguage :exec
INSERT INTO juustagram_languages (
  key,
  value
) VALUES (
  $1, $2
)
ON CONFLICT (key)
DO UPDATE SET
  value = EXCLUDED.value;

-- name: UpsertJuustagramShipGroupTemplate :exec
INSERT INTO juustagram_ship_group_templates (
  ship_group,
  name,
  background,
  sculpture,
  sculpture_ii,
  nationality,
  type
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (ship_group)
DO UPDATE SET
  name = EXCLUDED.name,
  background = EXCLUDED.background,
  sculpture = EXCLUDED.sculpture,
  sculpture_ii = EXCLUDED.sculpture_ii,
  nationality = EXCLUDED.nationality,
  type = EXCLUDED.type;
