package orm

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
)

func ListShipsPage(params ShipQueryParams) ([]Ship, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM ships
WHERE ($1::bigint IS NULL OR rarity_id = $1)
	AND ($2::bigint IS NULL OR type = $2)
	AND ($3::bigint IS NULL OR nationality = $3)
	AND ($4::text = '' OR LOWER(name) LIKE '%' || LOWER($4) || '%')
`, pgInt8FromUint32Ptr(params.RarityID), pgInt8FromUint32Ptr(params.TypeID), pgInt8FromUint32Ptr(params.NationalityID), params.Name).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT template_id, name, english_name, rarity_id, star, type, nationality, build_time, pool_id
FROM ships
WHERE ($1::bigint IS NULL OR rarity_id = $1)
	AND ($2::bigint IS NULL OR type = $2)
	AND ($3::bigint IS NULL OR nationality = $3)
	AND ($4::text = '' OR LOWER(name) LIKE '%' || LOWER($4) || '%')
ORDER BY template_id ASC
OFFSET $5
LIMIT $6
`, pgInt8FromUint32Ptr(params.RarityID), pgInt8FromUint32Ptr(params.TypeID), pgInt8FromUint32Ptr(params.NationalityID), params.Name, int64(params.Offset), int64(params.Limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	ships := make([]Ship, 0)
	for rows.Next() {
		var ship Ship
		var poolID pgtype.Int8
		if err := rows.Scan(
			&ship.TemplateID,
			&ship.Name,
			&ship.EnglishName,
			&ship.RarityID,
			&ship.Star,
			&ship.Type,
			&ship.Nationality,
			&ship.BuildTime,
			&poolID,
		); err != nil {
			return nil, 0, err
		}
		ship.PoolID = pgInt8PtrToUint32Ptr(poolID)
		ships = append(ships, ship)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return ships, total, nil
}

func ListEquipmentPage(offset int, limit int) ([]Equipment, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM equipments`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, base, destroy_gold, destroy_item, equip_limit, "group", important, level, next, prev, restore_gold, restore_item, ship_type_forbidden, trans_use_gold, trans_use_item, type, upgrade_formula_id
FROM equipments
ORDER BY id ASC
OFFSET $1
LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	equipment := make([]Equipment, 0)
	for rows.Next() {
		row, err := scanEquipment(rows)
		if err != nil {
			return nil, 0, err
		}
		equipment = append(equipment, *row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return equipment, total, nil
}

func GetEquipmentByID(id uint32) (*Equipment, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, base, destroy_gold, destroy_item, equip_limit, "group", important, level, next, prev, restore_gold, restore_item, ship_type_forbidden, trans_use_gold, trans_use_item, type, upgrade_formula_id
FROM equipments
WHERE id = $1
`, int64(id))
	equipment, err := scanEquipment(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return equipment, nil
}

func CreateEquipmentRecord(equipment *Equipment) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
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
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
`,
		int64(equipment.ID),
		pgInt8FromUint32Ptr(equipment.Base),
		int64(equipment.DestroyGold),
		equipment.DestroyItem,
		equipment.EquipLimit,
		int64(equipment.Group),
		int64(equipment.Important),
		int64(equipment.Level),
		equipment.Next,
		equipment.Prev,
		int64(equipment.RestoreGold),
		equipment.RestoreItem,
		equipment.ShipTypeForbidden,
		int64(equipment.TransUseGold),
		equipment.TransUseItem,
		int64(equipment.Type),
		equipment.UpgradeFormulaID,
	)
	return err
}

func UpdateEquipmentRecord(equipment *Equipment) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE equipments
SET base = $2,
	destroy_gold = $3,
	destroy_item = $4,
	equip_limit = $5,
	"group" = $6,
	important = $7,
	level = $8,
	next = $9,
	prev = $10,
	restore_gold = $11,
	restore_item = $12,
	ship_type_forbidden = $13,
	trans_use_gold = $14,
	trans_use_item = $15,
	type = $16,
	upgrade_formula_id = $17
WHERE id = $1
`,
		int64(equipment.ID),
		pgInt8FromUint32Ptr(equipment.Base),
		int64(equipment.DestroyGold),
		equipment.DestroyItem,
		equipment.EquipLimit,
		int64(equipment.Group),
		int64(equipment.Important),
		int64(equipment.Level),
		equipment.Next,
		equipment.Prev,
		int64(equipment.RestoreGold),
		equipment.RestoreItem,
		equipment.ShipTypeForbidden,
		int64(equipment.TransUseGold),
		equipment.TransUseItem,
		int64(equipment.Type),
		equipment.UpgradeFormulaID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteEquipmentRecord(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM equipments WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListWeaponsPage(offset int, limit int) ([]Weapon, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM weapons`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, action_index, aim_type, angle, attack_attribute, attack_attribute_ratio, auto_aftercast, axis_angle, barrage_id, bullet_id, charge_param, corrected, damage, effect_move, expose, fire_fx, fire_fx_loop_type, fire_sfx, initial_over_heat, min_range, oxy_type, precast_param, queue, range, recover_time, reload_max, search_condition, search_type, shake_screen, spawn_bound, suppress, torpedo_ammo, type
FROM weapons
ORDER BY id ASC
OFFSET $1
LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	weapons := make([]Weapon, 0)
	for rows.Next() {
		row, err := scanWeapon(rows)
		if err != nil {
			return nil, 0, err
		}
		weapons = append(weapons, *row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return weapons, total, nil
}

func GetWeaponByID(id uint32) (*Weapon, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, action_index, aim_type, angle, attack_attribute, attack_attribute_ratio, auto_aftercast, axis_angle, barrage_id, bullet_id, charge_param, corrected, damage, effect_move, expose, fire_fx, fire_fx_loop_type, fire_sfx, initial_over_heat, min_range, oxy_type, precast_param, queue, range, recover_time, reload_max, search_condition, search_type, shake_screen, spawn_bound, suppress, torpedo_ammo, type
FROM weapons
WHERE id = $1
`, int64(id))
	weapon, err := scanWeapon(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return weapon, nil
}

func CreateWeaponRecord(weapon *Weapon) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
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
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33)
`,
		int64(weapon.ID),
		weapon.ActionIndex,
		weapon.AimType,
		weapon.Angle,
		weapon.AttackAttribute,
		weapon.AttackAttributeRatio,
		weapon.AutoAftercast,
		weapon.AxisAngle,
		weapon.BarrageID,
		weapon.BulletID,
		weapon.ChargeParam,
		weapon.Corrected,
		weapon.Damage,
		weapon.EffectMove,
		weapon.Expose,
		weapon.FireFX,
		weapon.FireFXLoopType,
		weapon.FireSFX,
		weapon.InitialOverHeat,
		weapon.MinRange,
		weapon.OxyType,
		weapon.PrecastParam,
		weapon.Queue,
		weapon.Range,
		weapon.RecoverTime,
		weapon.ReloadMax,
		weapon.SearchCondition,
		weapon.SearchType,
		weapon.ShakeScreen,
		weapon.SpawnBound,
		weapon.Suppress,
		weapon.TorpedoAmmo,
		weapon.Type,
	)
	return err
}

func UpdateWeaponRecord(weapon *Weapon) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE weapons
SET action_index = $2,
	aim_type = $3,
	angle = $4,
	attack_attribute = $5,
	attack_attribute_ratio = $6,
	auto_aftercast = $7,
	axis_angle = $8,
	barrage_id = $9,
	bullet_id = $10,
	charge_param = $11,
	corrected = $12,
	damage = $13,
	effect_move = $14,
	expose = $15,
	fire_fx = $16,
	fire_fx_loop_type = $17,
	fire_sfx = $18,
	initial_over_heat = $19,
	min_range = $20,
	oxy_type = $21,
	precast_param = $22,
	queue = $23,
	range = $24,
	recover_time = $25,
	reload_max = $26,
	search_condition = $27,
	search_type = $28,
	shake_screen = $29,
	spawn_bound = $30,
	suppress = $31,
	torpedo_ammo = $32,
	type = $33
WHERE id = $1
`,
		int64(weapon.ID),
		weapon.ActionIndex,
		weapon.AimType,
		weapon.Angle,
		weapon.AttackAttribute,
		weapon.AttackAttributeRatio,
		weapon.AutoAftercast,
		weapon.AxisAngle,
		weapon.BarrageID,
		weapon.BulletID,
		weapon.ChargeParam,
		weapon.Corrected,
		weapon.Damage,
		weapon.EffectMove,
		weapon.Expose,
		weapon.FireFX,
		weapon.FireFXLoopType,
		weapon.FireSFX,
		weapon.InitialOverHeat,
		weapon.MinRange,
		weapon.OxyType,
		weapon.PrecastParam,
		weapon.Queue,
		weapon.Range,
		weapon.RecoverTime,
		weapon.ReloadMax,
		weapon.SearchCondition,
		weapon.SearchType,
		weapon.ShakeScreen,
		weapon.SpawnBound,
		weapon.Suppress,
		weapon.TorpedoAmmo,
		weapon.Type,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteWeaponRecord(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM weapons WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListSkillsPage(offset int, limit int) ([]Skill, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM skills`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, name, "desc", cd, painting, picture, ani_effect, ui_effect, effect_list
FROM skills
ORDER BY id ASC
OFFSET $1
LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	skills := make([]Skill, 0)
	for rows.Next() {
		row, err := scanSkill(rows)
		if err != nil {
			return nil, 0, err
		}
		skills = append(skills, *row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return skills, total, nil
}

func GetSkillByID(id uint32) (*Skill, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, name, "desc", cd, painting, picture, ani_effect, ui_effect, effect_list
FROM skills
WHERE id = $1
`, int64(id))
	skill, err := scanSkill(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return skill, nil
}

func CreateSkillRecord(skill *Skill) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO skills (id, name, "desc", cd, painting, picture, ani_effect, ui_effect, effect_list)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, int64(skill.ID), skill.Name, skill.Desc, int64(skill.CD), skill.Painting, skill.Picture, skill.AniEffect, skill.UIEffect, skill.EffectList)
	return err
}

func UpdateSkillRecord(skill *Skill) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE skills
SET name = $2,
	"desc" = $3,
	cd = $4,
	painting = $5,
	picture = $6,
	ani_effect = $7,
	ui_effect = $8,
	effect_list = $9
WHERE id = $1
`, int64(skill.ID), skill.Name, skill.Desc, int64(skill.CD), skill.Painting, skill.Picture, skill.AniEffect, skill.UIEffect, skill.EffectList)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteSkillRecord(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM skills WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListBuffsPage(offset int, limit int) ([]Buff, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM buffs`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, name, description, max_time, benefit_type
FROM buffs
ORDER BY id ASC
OFFSET $1
LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	buffs := make([]Buff, 0)
	for rows.Next() {
		var buff Buff
		if err := rows.Scan(&buff.ID, &buff.Name, &buff.Description, &buff.MaxTime, &buff.BenefitType); err != nil {
			return nil, 0, err
		}
		buffs = append(buffs, buff)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return buffs, total, nil
}

func GetBuffByID(id uint32) (*Buff, error) {
	ctx := context.Background()
	var buff Buff
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, name, description, max_time, benefit_type
FROM buffs
WHERE id = $1
`, int64(id)).Scan(&buff.ID, &buff.Name, &buff.Description, &buff.MaxTime, &buff.BenefitType)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &buff, nil
}

func CreateBuffRecord(buff *Buff) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO buffs (id, name, description, max_time, benefit_type)
VALUES ($1, $2, $3, $4, $5)
`, int64(buff.ID), buff.Name, buff.Description, buff.MaxTime, buff.BenefitType)
	return err
}

func UpdateBuffRecord(buff *Buff) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE buffs
SET name = $2,
	description = $3,
	max_time = $4,
	benefit_type = $5
WHERE id = $1
`, int64(buff.ID), buff.Name, buff.Description, buff.MaxTime, buff.BenefitType)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteBuffRecord(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM buffs WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListSkinsPage(offset int, limit int) ([]Skin, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM skins`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, skinsSelect+`
ORDER BY id ASC
OFFSET $1
LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	skins := make([]Skin, 0)
	for rows.Next() {
		row, err := scanSkin(rows)
		if err != nil {
			return nil, 0, err
		}
		skins = append(skins, *row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return skins, total, nil
}

func ListSkinsByShipGroupPage(shipGroup uint32, offset int, limit int) ([]Skin, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM skins WHERE ship_group = $1`, int64(shipGroup)).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, skinsSelect+`
WHERE ship_group = $1
ORDER BY id ASC
OFFSET $2
LIMIT $3
`, int64(shipGroup), int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	skins := make([]Skin, 0)
	for rows.Next() {
		row, err := scanSkin(rows)
		if err != nil {
			return nil, 0, err
		}
		skins = append(skins, *row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return skins, total, nil
}

func GetSkinByID(id uint32) (*Skin, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, skinsSelect+`
WHERE id = $1
`, int64(id))
	skin, err := scanSkin(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return skin, nil
}

func CreateSkinRecord(skin *Skin) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO skins (
	id, name, ship_group, "desc", bg, bg_sp, bgm, painting, prefab, change_skin,
	show_skin, skeleton_skin, ship_l2_d_id, l2_d_animations, l2_d_drag_rate, l2_d_para_range,
	l2_dse, l2_d_voice_calib, part_scale, main_ui_fx, spine_offset, spine_profile, tag, time,
	get_showing, purchase_offset, shop_offset, rarity_bg, special_effects, group_index, gyro,
	hand_id, illustrator, illustrator2, voice_actor, voice_actor2, double_char, lip_smoothing,
	lip_sync_gain, l2_d_ignore_drag, skin_type, shop_id, shop_type_id, shop_dynamic_hx,
	spine_action, spine_use_live2_d, live2_d_offset, live2_d_profile, fx_container, bound_bone, smoke
)
VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
	$11, $12, $13, $14, $15, $16,
	$17, $18, $19, $20, $21, $22, $23, $24,
	$25, $26, $27, $28, $29, $30, $31,
	$32, $33, $34, $35, $36, $37, $38,
	$39, $40, $41, $42, $43, $44,
	$45, $46, $47, $48, $49, $50, $51
)
`,
		int64(skin.ID), skin.Name, skin.ShipGroup, skin.Desc, skin.BG, skin.BGSp, skin.BGM, skin.Painting, skin.Prefab, skin.ChangeSkin,
		skin.ShowSkin, skin.SkeletonSkin, skin.ShipL2DID, skin.L2DAnimations, skin.L2DDragRate, skin.L2DParaRange,
		skin.L2DSE, skin.L2DVoiceCalib, skin.PartScale, skin.MainUIFX, skin.SpineOffset, skin.SpineProfile, skin.Tag, skin.Time,
		skin.GetShowing, skin.PurchaseOffset, skin.ShopOffset, skin.RarityBG, skin.SpecialEffects, pgInt8FromIntPtr(skin.GroupIndex), pgInt8FromIntPtr(skin.Gyro),
		pgInt8FromIntPtr(skin.HandID), pgInt8FromIntPtr(skin.Illustrator), pgInt8FromIntPtr(skin.Illustrator2), pgInt8FromIntPtr(skin.VoiceActor), pgInt8FromIntPtr(skin.VoiceActor2), pgInt8FromIntPtr(skin.DoubleChar), pgInt8FromIntPtr(skin.LipSmoothing),
		pgInt8FromIntPtr(skin.LipSyncGain), pgInt8FromIntPtr(skin.L2DIgnoreDrag), pgInt8FromIntPtr(skin.SkinType), pgInt8FromIntPtr(skin.ShopID), pgInt8FromIntPtr(skin.ShopTypeID), pgInt8FromIntPtr(skin.ShopDynamicHX),
		skin.SpineAction, pgInt8FromIntPtr(skin.SpineUseLive2D), skin.Live2DOffset, skin.Live2DProfile, skin.FXContainer, skin.BoundBone, skin.Smoke,
	)
	return err
}

func UpdateSkinRecord(skin *Skin) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE skins
SET name = $2,
	ship_group = $3,
	"desc" = $4,
	bg = $5,
	bg_sp = $6,
	bgm = $7,
	painting = $8,
	prefab = $9,
	change_skin = $10,
	show_skin = $11,
	skeleton_skin = $12,
	ship_l2_d_id = $13,
	l2_d_animations = $14,
	l2_d_drag_rate = $15,
	l2_d_para_range = $16,
	l2_dse = $17,
	l2_d_voice_calib = $18,
	part_scale = $19,
	main_ui_fx = $20,
	spine_offset = $21,
	spine_profile = $22,
	tag = $23,
	time = $24,
	get_showing = $25,
	purchase_offset = $26,
	shop_offset = $27,
	rarity_bg = $28,
	special_effects = $29,
	group_index = $30,
	gyro = $31,
	hand_id = $32,
	illustrator = $33,
	illustrator2 = $34,
	voice_actor = $35,
	voice_actor2 = $36,
	double_char = $37,
	lip_smoothing = $38,
	lip_sync_gain = $39,
	l2_d_ignore_drag = $40,
	skin_type = $41,
	shop_id = $42,
	shop_type_id = $43,
	shop_dynamic_hx = $44,
	spine_action = $45,
	spine_use_live2_d = $46,
	live2_d_offset = $47,
	live2_d_profile = $48,
	fx_container = $49,
	bound_bone = $50,
	smoke = $51
WHERE id = $1
`,
		int64(skin.ID), skin.Name, skin.ShipGroup, skin.Desc, skin.BG, skin.BGSp, skin.BGM, skin.Painting, skin.Prefab, skin.ChangeSkin,
		skin.ShowSkin, skin.SkeletonSkin, skin.ShipL2DID, skin.L2DAnimations, skin.L2DDragRate, skin.L2DParaRange,
		skin.L2DSE, skin.L2DVoiceCalib, skin.PartScale, skin.MainUIFX, skin.SpineOffset, skin.SpineProfile, skin.Tag, skin.Time,
		skin.GetShowing, skin.PurchaseOffset, skin.ShopOffset, skin.RarityBG, skin.SpecialEffects, pgInt8FromIntPtr(skin.GroupIndex), pgInt8FromIntPtr(skin.Gyro),
		pgInt8FromIntPtr(skin.HandID), pgInt8FromIntPtr(skin.Illustrator), pgInt8FromIntPtr(skin.Illustrator2), pgInt8FromIntPtr(skin.VoiceActor), pgInt8FromIntPtr(skin.VoiceActor2), pgInt8FromIntPtr(skin.DoubleChar), pgInt8FromIntPtr(skin.LipSmoothing),
		pgInt8FromIntPtr(skin.LipSyncGain), pgInt8FromIntPtr(skin.L2DIgnoreDrag), pgInt8FromIntPtr(skin.SkinType), pgInt8FromIntPtr(skin.ShopID), pgInt8FromIntPtr(skin.ShopTypeID), pgInt8FromIntPtr(skin.ShopDynamicHX),
		skin.SpineAction, pgInt8FromIntPtr(skin.SpineUseLive2D), skin.Live2DOffset, skin.Live2DProfile, skin.FXContainer, skin.BoundBone, skin.Smoke,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteSkinRecord(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM skins WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListGlobalSkinRestrictionsPage(offset int, limit int) ([]GlobalSkinRestriction, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM global_skin_restrictions`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT skin_id, type
FROM global_skin_restrictions
ORDER BY skin_id ASC
OFFSET $1
LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	restrictions := make([]GlobalSkinRestriction, 0)
	for rows.Next() {
		var restriction GlobalSkinRestriction
		if err := rows.Scan(&restriction.SkinID, &restriction.Type); err != nil {
			return nil, 0, err
		}
		restrictions = append(restrictions, restriction)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return restrictions, total, nil
}

func GetGlobalSkinRestrictionBySkinID(skinID uint32) (*GlobalSkinRestriction, error) {
	ctx := context.Background()
	var restriction GlobalSkinRestriction
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT skin_id, type
FROM global_skin_restrictions
WHERE skin_id = $1
`, int64(skinID)).Scan(&restriction.SkinID, &restriction.Type)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &restriction, nil
}

func CreateGlobalSkinRestriction(restriction *GlobalSkinRestriction) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO global_skin_restrictions (skin_id, type)
VALUES ($1, $2)
`, int64(restriction.SkinID), int64(restriction.Type))
	return err
}

func UpdateGlobalSkinRestriction(restriction *GlobalSkinRestriction) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE global_skin_restrictions
SET type = $2
WHERE skin_id = $1
`, int64(restriction.SkinID), int64(restriction.Type))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteGlobalSkinRestriction(skinID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM global_skin_restrictions WHERE skin_id = $1`, int64(skinID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListGlobalSkinRestrictionWindowsPage(offset int, limit int) ([]GlobalSkinRestrictionWindow, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM global_skin_restriction_windows`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, skin_id, type, start_time, stop_time
FROM global_skin_restriction_windows
ORDER BY id ASC
OFFSET $1
LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	windows := make([]GlobalSkinRestrictionWindow, 0)
	for rows.Next() {
		var window GlobalSkinRestrictionWindow
		if err := rows.Scan(&window.ID, &window.SkinID, &window.Type, &window.StartTime, &window.StopTime); err != nil {
			return nil, 0, err
		}
		windows = append(windows, window)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return windows, total, nil
}

func GetGlobalSkinRestrictionWindowByID(id uint32) (*GlobalSkinRestrictionWindow, error) {
	ctx := context.Background()
	var window GlobalSkinRestrictionWindow
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, skin_id, type, start_time, stop_time
FROM global_skin_restriction_windows
WHERE id = $1
`, int64(id)).Scan(&window.ID, &window.SkinID, &window.Type, &window.StartTime, &window.StopTime)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &window, nil
}

func CreateGlobalSkinRestrictionWindow(window *GlobalSkinRestrictionWindow) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO global_skin_restriction_windows (id, skin_id, type, start_time, stop_time)
VALUES ($1, $2, $3, $4, $5)
`, int64(window.ID), int64(window.SkinID), int64(window.Type), int64(window.StartTime), int64(window.StopTime))
	return err
}

func UpdateGlobalSkinRestrictionWindow(window *GlobalSkinRestrictionWindow) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE global_skin_restriction_windows
SET skin_id = $2,
	type = $3,
	start_time = $4,
	stop_time = $5
WHERE id = $1
`, int64(window.ID), int64(window.SkinID), int64(window.Type), int64(window.StartTime), int64(window.StopTime))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteGlobalSkinRestrictionWindow(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM global_skin_restriction_windows WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListConfigEntriesFiltered(category string, key string) ([]ConfigEntry, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, category, "key", data
FROM config_entries
WHERE ($1::text = '' OR category = $1)
	AND ($2::text = '' OR "key" = $2)
ORDER BY category ASC, "key" ASC
`, category, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]ConfigEntry, 0)
	for rows.Next() {
		var entry ConfigEntry
		if err := rows.Scan(&entry.ID, &entry.Category, &entry.Key, &entry.Data); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func GetConfigEntryByID(id uint64) (*ConfigEntry, error) {
	ctx := context.Background()
	var entry ConfigEntry
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, category, "key", data
FROM config_entries
WHERE id = $1
`, int64(id)).Scan(&entry.ID, &entry.Category, &entry.Key, &entry.Data)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func CreateConfigEntryRecord(entry *ConfigEntry) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO config_entries (category, "key", data)
VALUES ($1, $2, $3)
`, entry.Category, entry.Key, entry.Data)
	return err
}

func UpdateConfigEntryRecord(entry *ConfigEntry) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE config_entries
SET category = $2,
	"key" = $3,
	data = $4
WHERE id = $1
`, int64(entry.ID), entry.Category, entry.Key, entry.Data)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteConfigEntryByID(id uint64) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM config_entries WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

type equipmentScanner interface {
	Scan(dest ...any) error
}

func scanEquipment(scanner equipmentScanner) (*Equipment, error) {
	var equipment Equipment
	var base pgtype.Int8
	err := scanner.Scan(
		&equipment.ID,
		&base,
		&equipment.DestroyGold,
		&equipment.DestroyItem,
		&equipment.EquipLimit,
		&equipment.Group,
		&equipment.Important,
		&equipment.Level,
		&equipment.Next,
		&equipment.Prev,
		&equipment.RestoreGold,
		&equipment.RestoreItem,
		&equipment.ShipTypeForbidden,
		&equipment.TransUseGold,
		&equipment.TransUseItem,
		&equipment.Type,
		&equipment.UpgradeFormulaID,
	)
	if err != nil {
		return nil, err
	}
	equipment.Base = pgInt8PtrToUint32Ptr(base)
	return &equipment, nil
}

type weaponScanner interface {
	Scan(dest ...any) error
}

func scanWeapon(scanner weaponScanner) (*Weapon, error) {
	var weapon Weapon
	err := scanner.Scan(
		&weapon.ID,
		&weapon.ActionIndex,
		&weapon.AimType,
		&weapon.Angle,
		&weapon.AttackAttribute,
		&weapon.AttackAttributeRatio,
		&weapon.AutoAftercast,
		&weapon.AxisAngle,
		&weapon.BarrageID,
		&weapon.BulletID,
		&weapon.ChargeParam,
		&weapon.Corrected,
		&weapon.Damage,
		&weapon.EffectMove,
		&weapon.Expose,
		&weapon.FireFX,
		&weapon.FireFXLoopType,
		&weapon.FireSFX,
		&weapon.InitialOverHeat,
		&weapon.MinRange,
		&weapon.OxyType,
		&weapon.PrecastParam,
		&weapon.Queue,
		&weapon.Range,
		&weapon.RecoverTime,
		&weapon.ReloadMax,
		&weapon.SearchCondition,
		&weapon.SearchType,
		&weapon.ShakeScreen,
		&weapon.SpawnBound,
		&weapon.Suppress,
		&weapon.TorpedoAmmo,
		&weapon.Type,
	)
	if err != nil {
		return nil, err
	}
	return &weapon, nil
}

type skillScanner interface {
	Scan(dest ...any) error
}

func scanSkill(scanner skillScanner) (*Skill, error) {
	var skill Skill
	err := scanner.Scan(
		&skill.ID,
		&skill.Name,
		&skill.Desc,
		&skill.CD,
		&skill.Painting,
		&skill.Picture,
		&skill.AniEffect,
		&skill.UIEffect,
		&skill.EffectList,
	)
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

const skinsSelect = `
SELECT id, name, ship_group, "desc", bg, bg_sp, bgm, painting, prefab, change_skin,
	show_skin, skeleton_skin, ship_l2_d_id, l2_d_animations, l2_d_drag_rate, l2_d_para_range,
	l2_dse, l2_d_voice_calib, part_scale, main_ui_fx, spine_offset, spine_profile, tag, time,
	get_showing, purchase_offset, shop_offset, rarity_bg, special_effects, group_index, gyro,
	hand_id, illustrator, illustrator2, voice_actor, voice_actor2, double_char, lip_smoothing,
	lip_sync_gain, l2_d_ignore_drag, skin_type, shop_id, shop_type_id, shop_dynamic_hx,
	spine_action, spine_use_live2_d, live2_d_offset, live2_d_profile, fx_container, bound_bone, smoke
FROM skins
`

type skinScanner interface {
	Scan(dest ...any) error
}

func scanSkin(scanner skinScanner) (*Skin, error) {
	var skin Skin
	var groupIndex pgtype.Int8
	var gyro pgtype.Int8
	var handID pgtype.Int8
	var illustrator pgtype.Int8
	var illustrator2 pgtype.Int8
	var voiceActor pgtype.Int8
	var voiceActor2 pgtype.Int8
	var doubleChar pgtype.Int8
	var lipSmoothing pgtype.Int8
	var lipSyncGain pgtype.Int8
	var l2dIgnoreDrag pgtype.Int8
	var skinType pgtype.Int8
	var shopID pgtype.Int8
	var shopTypeID pgtype.Int8
	var shopDynamicHX pgtype.Int8
	var spineUseLive2D pgtype.Int8

	err := scanner.Scan(
		&skin.ID,
		&skin.Name,
		&skin.ShipGroup,
		&skin.Desc,
		&skin.BG,
		&skin.BGSp,
		&skin.BGM,
		&skin.Painting,
		&skin.Prefab,
		&skin.ChangeSkin,
		&skin.ShowSkin,
		&skin.SkeletonSkin,
		&skin.ShipL2DID,
		&skin.L2DAnimations,
		&skin.L2DDragRate,
		&skin.L2DParaRange,
		&skin.L2DSE,
		&skin.L2DVoiceCalib,
		&skin.PartScale,
		&skin.MainUIFX,
		&skin.SpineOffset,
		&skin.SpineProfile,
		&skin.Tag,
		&skin.Time,
		&skin.GetShowing,
		&skin.PurchaseOffset,
		&skin.ShopOffset,
		&skin.RarityBG,
		&skin.SpecialEffects,
		&groupIndex,
		&gyro,
		&handID,
		&illustrator,
		&illustrator2,
		&voiceActor,
		&voiceActor2,
		&doubleChar,
		&lipSmoothing,
		&lipSyncGain,
		&l2dIgnoreDrag,
		&skinType,
		&shopID,
		&shopTypeID,
		&shopDynamicHX,
		&skin.SpineAction,
		&spineUseLive2D,
		&skin.Live2DOffset,
		&skin.Live2DProfile,
		&skin.FXContainer,
		&skin.BoundBone,
		&skin.Smoke,
	)
	if err != nil {
		return nil, err
	}

	skin.GroupIndex = pgInt8PtrToIntPtr(groupIndex)
	skin.Gyro = pgInt8PtrToIntPtr(gyro)
	skin.HandID = pgInt8PtrToIntPtr(handID)
	skin.Illustrator = pgInt8PtrToIntPtr(illustrator)
	skin.Illustrator2 = pgInt8PtrToIntPtr(illustrator2)
	skin.VoiceActor = pgInt8PtrToIntPtr(voiceActor)
	skin.VoiceActor2 = pgInt8PtrToIntPtr(voiceActor2)
	skin.DoubleChar = pgInt8PtrToIntPtr(doubleChar)
	skin.LipSmoothing = pgInt8PtrToIntPtr(lipSmoothing)
	skin.LipSyncGain = pgInt8PtrToIntPtr(lipSyncGain)
	skin.L2DIgnoreDrag = pgInt8PtrToIntPtr(l2dIgnoreDrag)
	skin.SkinType = pgInt8PtrToIntPtr(skinType)
	skin.ShopID = pgInt8PtrToIntPtr(shopID)
	skin.ShopTypeID = pgInt8PtrToIntPtr(shopTypeID)
	skin.ShopDynamicHX = pgInt8PtrToIntPtr(shopDynamicHX)
	skin.SpineUseLive2D = pgInt8PtrToIntPtr(spineUseLive2D)

	return &skin, nil
}

func pgInt8FromIntPtr(value *int) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(*value), Valid: true}
}

func pgInt8PtrToIntPtr(value pgtype.Int8) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int64)
	return &v
}
