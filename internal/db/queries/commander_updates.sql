-- Commander update queries

-- name: UpdateCommanderRandomShipMode :exec
UPDATE commanders
SET random_ship_mode = $2
WHERE commander_id = $1;

-- name: UpdateCommanderRoom :exec
UPDATE commanders
SET room_id = $2
WHERE commander_id = $1;

-- name: UpdateCommanderLastLogin :exec
UPDATE commanders
SET last_login = $2
WHERE commander_id = $1;

-- name: UpdateCommanderExchangeCount :exec
UPDATE commanders
SET exchange_count = $2
WHERE commander_id = $1;

-- name: UpdateCommanderDrawCounts :exec
UPDATE commanders
SET draw_count1 = $2,
    draw_count10 = $3
WHERE commander_id = $1;

-- name: UpdateCommanderCore :exec
UPDATE commanders
SET
  account_id = $2,
  level = $3,
  exp = $4,
  name = $5,
  last_login = $6,
  guide_index = $7,
  new_guide_index = $8,
  name_change_cooldown = $9,
  room_id = $10,
  exchange_count = $11,
  draw_count1 = $12,
  draw_count10 = $13,
  support_requisition_count = $14,
  support_requisition_month = $15,
  collect_attack_count = $16,
  acc_pay_lv = $17,
  living_area_cover_id = $18,
  selected_icon_frame_id = $19,
  selected_chat_frame_id = $20,
  selected_battle_ui_id = $21,
  display_icon_id = $22,
  display_skin_id = $23,
  display_icon_theme_id = $24,
  manifesto = $25,
  dorm_name = $26,
  random_ship_mode = $27,
  random_flag_ship_enabled = $28
WHERE commander_id = $1;
