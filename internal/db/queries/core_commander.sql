-- Commander aggregate load queries (Milestone 4)

-- name: CreateCommander :exec
INSERT INTO commanders (
  commander_id,
  account_id,
  level,
  exp,
  name,
  last_login,
  guide_index,
  new_guide_index,
  name_change_cooldown,
  room_id,
  exchange_count,
  draw_count1,
  draw_count10,
  support_requisition_count,
  support_requisition_month,
  collect_attack_count,
  acc_pay_lv,
  living_area_cover_id,
  selected_icon_frame_id,
  selected_chat_frame_id,
  selected_battle_ui_id,
  display_icon_id,
  display_skin_id,
  display_icon_theme_id,
  manifesto,
  dorm_name,
  random_ship_mode,
  random_flag_ship_enabled
) VALUES (
  $1, $2, 1, 0, $3, now(), $4, $5, '1970-01-01 00:00:00+00',
  0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false
);

-- name: GetCommanderByAccountID :one
SELECT
  commander_id,
  account_id,
  level,
  exp,
  name,
  last_login,
  guide_index,
  new_guide_index,
  name_change_cooldown,
  room_id,
  exchange_count,
  draw_count1,
  draw_count10,
  support_requisition_count,
  support_requisition_month,
  collect_attack_count,
  acc_pay_lv,
  living_area_cover_id,
  selected_icon_frame_id,
  selected_chat_frame_id,
  selected_battle_ui_id,
  display_icon_id,
  display_skin_id,
  display_icon_theme_id,
  manifesto,
  dorm_name,
  random_ship_mode,
  random_flag_ship_enabled
FROM commanders
WHERE account_id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: CountCommandersByName :one
SELECT COUNT(*)::bigint
FROM commanders
WHERE name = $1
  AND deleted_at IS NULL;

-- name: GetCommanderByID :one
SELECT
  commander_id,
  account_id,
  level,
  exp,
  name,
  last_login,
  guide_index,
  new_guide_index,
  name_change_cooldown,
  room_id,
  exchange_count,
  draw_count1,
  draw_count10,
  support_requisition_count,
  support_requisition_month,
  collect_attack_count,
  acc_pay_lv,
  living_area_cover_id,
  selected_icon_frame_id,
  selected_chat_frame_id,
  selected_battle_ui_id,
  display_icon_id,
  display_skin_id,
  display_icon_theme_id,
  manifesto,
  dorm_name,
  random_ship_mode,
  random_flag_ship_enabled
FROM commanders
WHERE commander_id = $1
  AND deleted_at IS NULL;

-- name: ListOwnedShipsWithShipByOwnerID :many
SELECT
  os.owner_id,
  os.ship_id,
  os.id,
  os.level,
  os.exp,
  os.surplus_exp,
  os.max_level,
  os.intimacy,
  os.is_locked,
  os.propose,
  os.common_flag,
  os.blueprint_flag,
  os.proficiency,
  os.activity_npc,
  os.custom_name,
  os.change_name_timestamp,
  os.create_time,
  os.energy,
  os.state,
  os.state_info1,
  os.state_info2,
  os.state_info3,
  os.state_info4,
  os.skin_id,
  os.is_secretary,
  os.secretary_position,
  os.secretary_phantom_id,
  s.template_id AS ship_template_id,
  s.name AS ship_name,
  s.english_name AS ship_english_name,
  s.rarity_id AS ship_rarity_id,
  s.star AS ship_star,
  s.type AS ship_type,
  s.nationality AS ship_nationality,
  s.build_time AS ship_build_time,
  s.pool_id AS ship_pool_id
FROM owned_ships AS os
JOIN ships AS s ON s.template_id = os.ship_id
WHERE os.owner_id = $1
  AND os.deleted_at IS NULL
ORDER BY os.id ASC;

-- name: ListOwnedShipEquipmentsByOwnerID :many
SELECT
  owner_id,
  ship_id,
  pos,
  equip_id,
  skin_id
FROM owned_ship_equipments
WHERE owner_id = $1
ORDER BY ship_id ASC, pos ASC;

-- name: ListOwnedShipStrengthsByOwnerID :many
SELECT
  owner_id,
  ship_id,
  strength_id,
  exp
FROM owned_ship_strengths
WHERE owner_id = $1
ORDER BY ship_id ASC, strength_id ASC;

-- name: ListCommanderItemsWithItemByCommanderID :many
SELECT
  ci.commander_id,
  ci.item_id,
  ci.count,
  i.id AS item_id_full,
  i.name AS item_name,
  i.rarity AS item_rarity,
  i.shop_id AS item_shop_id,
  i.type AS item_type,
  i.virtual_type AS item_virtual_type
FROM commander_items AS ci
JOIN items AS i ON i.id = ci.item_id
WHERE ci.commander_id = $1
ORDER BY ci.item_id ASC;

-- name: ListCommanderMiscItemsWithItemByCommanderID :many
SELECT
  cmi.commander_id,
  cmi.item_id,
  cmi.data,
  i.id AS item_id_full,
  i.name AS item_name,
  i.rarity AS item_rarity,
  i.shop_id AS item_shop_id,
  i.type AS item_type,
  i.virtual_type AS item_virtual_type
FROM commander_misc_items AS cmi
JOIN items AS i ON i.id = cmi.item_id
WHERE cmi.commander_id = $1
ORDER BY cmi.item_id ASC;

-- name: ListOwnedResourcesWithResourceByCommanderID :many
SELECT
  o.commander_id,
  o.resource_id,
  o.amount,
  r.id AS resource_id_full,
  r.item_id AS resource_item_id,
  r.name AS resource_name
FROM owned_resources AS o
JOIN resources AS r ON r.id = o.resource_id
WHERE o.commander_id = $1
ORDER BY o.resource_id ASC;

-- name: ListBuildsWithShipByBuilderID :many
SELECT
  b.id,
  b.builder_id,
  b.ship_id,
  b.pool_id,
  b.finishes_at,
  s.template_id AS ship_template_id,
  s.name AS ship_name,
  s.english_name AS ship_english_name,
  s.rarity_id AS ship_rarity_id,
  s.star AS ship_star,
  s.type AS ship_type,
  s.nationality AS ship_nationality,
  s.build_time AS ship_build_time,
  s.pool_id AS ship_pool_id
FROM builds AS b
JOIN ships AS s ON s.template_id = b.ship_id
WHERE b.builder_id = $1
ORDER BY b.id ASC;

-- name: ListMailsByReceiverID :many
SELECT
  id,
  receiver_id,
  read,
  date,
  title,
  body,
  attachments_collected,
  is_important,
  custom_sender,
  is_archived,
  created_at
FROM mails
WHERE receiver_id = $1
ORDER BY id ASC;

-- name: ListMailAttachmentsByMailIDs :many
SELECT
  id,
  mail_id,
  type,
  item_id,
  quantity
FROM mail_attachments
WHERE mail_id = ANY($1::bigint[])
ORDER BY mail_id ASC, id ASC;

-- name: ListOwnedSkinsByCommanderID :many
SELECT
  commander_id,
  skin_id,
  expires_at
FROM owned_skins
WHERE commander_id = $1
ORDER BY skin_id ASC;

-- name: ListOwnedEquipmentsByCommanderID :many
SELECT
  commander_id,
  equipment_id,
  count
FROM owned_equipments
WHERE commander_id = $1
ORDER BY equipment_id ASC;

-- name: ListOwnedSpWeaponsByOwnerID :many
SELECT
  owner_id,
  id,
  template_id,
  attr_1,
  attr_2,
  attr_temp_1,
  attr_temp_2,
  effect,
  pt,
  equipped_ship_id
FROM owned_spweapons
WHERE owner_id = $1
ORDER BY id ASC;

-- name: ListFleetsByCommanderID :many
SELECT
  id,
  game_id,
  commander_id,
  name,
  ship_list,
  meowfficer_list
FROM fleets
WHERE commander_id = $1
ORDER BY game_id ASC;
