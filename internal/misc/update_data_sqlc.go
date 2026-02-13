package misc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
)

var dataFnSQLC = map[string]func(context.Context, string, *gen.Queries) error{
	"Items":                  importItemsSQLC,
	"Buffs":                  importBuffsSQLC,
	"Ships":                  importShipsSQLC,
	"Skins":                  importSkinsSQLC,
	"Resources":              importResourcesSQLC,
	"Pools":                  importPoolsSQLC,
	"Requisition":            importRequisitionShipsSQLC,
	"BuildTimes":             importBuildTimesSQLC,
	"ShopOffers":             importShopOffersSQLC,
	"Weapons":                importWeaponsSQLC,
	"Equipments":             importEquipmentsSQLC,
	"Skills":                 importSkillsSQLC,
	"Configs":                importConfigEntriesSQLC,
	"JuustagramTemplates":    importJuustagramTemplatesSQLC,
	"JuustagramNpcTemplates": importJuustagramNpcTemplatesSQLC,
	"JuustagramLanguage":     importJuustagramLanguageSQLC,
	"JuustagramShipGroups":   importJuustagramShipGroupsSQLC,
}

func updateAllDataSQLC(region string) {
	ctx := context.Background()
	err := db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		for _, key := range order {
			fn := dataFnSQLC[key]
			if fn == nil {
				return fmt.Errorf("missing sqlc importer for %s", key)
			}
			logger.LogEvent("GameData", "Updating", fmt.Sprintf("Updating %s (region=%s)", key, region), logger.LOG_LEVEL_INFO)
			if err := fn(ctx, region, q); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.LogEvent("GameData", "Updating", fmt.Sprintf("failed to update game data: %s", err.Error()), logger.LOG_LEVEL_ERROR)
	}
}

func importItemsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "sharecfgdata/item_data_statistics.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var item orm.Item
		if err := decoder.Decode(&item); err != nil {
			return err
		}
		if err := q.UpsertItem(ctx, gen.UpsertItemParams{
			ID:          int64(item.ID),
			Name:        item.Name,
			Rarity:      int32(item.Rarity),
			ShopID:      int32(item.ShopID),
			Type:        int32(item.Type),
			VirtualType: int32(item.VirtualType),
		}); err != nil {
			return err
		}
	}
	return nil
}

func importBuffsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/benefit_buff_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var buff orm.Buff
		if err := decoder.Decode(&buff); err != nil {
			return err
		}
		if err := q.UpsertBuff(ctx, gen.UpsertBuffParams{
			ID:          int64(buff.ID),
			Name:        buff.Name,
			Description: buff.Description,
			MaxTime:     int32(buff.MaxTime),
			BenefitType: buff.BenefitType,
		}); err != nil {
			return err
		}
	}
	return nil
}

func importShipsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "sharecfgdata/ship_data_statistics.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var ship orm.Ship
		if err := decoder.Decode(&ship); err != nil {
			return err
		}
		if err := q.UpsertShip(ctx, gen.UpsertShipParams{
			TemplateID:  int64(ship.TemplateID),
			Name:        ship.Name,
			EnglishName: ship.EnglishName,
			RarityID:    int64(ship.RarityID),
			Star:        int64(ship.Star),
			Type:        int64(ship.Type),
			Nationality: int64(ship.Nationality),
			BuildTime:   int64(ship.BuildTime),
			PoolID:      pgInt8FromUint32Ptr(ship.PoolID),
		}); err != nil {
			return err
		}
	}
	return nil
}

func importSkinsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/ship_skin_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var skin orm.Skin
		if err := decoder.Decode(&skin); err != nil {
			return err
		}
		if err := q.UpsertSkin(ctx, gen.UpsertSkinParams{
			ID:             int64(skin.ID),
			Name:           skin.Name,
			ShipGroup:      int64(skin.ShipGroup),
			Desc:           pgText(skin.Desc),
			Bg:             pgText(skin.BG),
			BgSp:           pgText(skin.BGSp),
			Bgm:            pgText(skin.BGM),
			Painting:       pgText(skin.Painting),
			Prefab:         pgText(skin.Prefab),
			ChangeSkin:     skin.ChangeSkin,
			ShowSkin:       pgText(skin.ShowSkin),
			SkeletonSkin:   pgText(skin.SkeletonSkin),
			ShipL2DID:      skin.ShipL2DID,
			L2DAnimations:  skin.L2DAnimations,
			L2DDragRate:    skin.L2DDragRate,
			L2DParaRange:   skin.L2DParaRange,
			L2Dse:          skin.L2DSE,
			L2DVoiceCalib:  skin.L2DVoiceCalib,
			PartScale:      pgText(skin.PartScale),
			MainUiFx:       pgText(skin.MainUIFX),
			SpineOffset:    skin.SpineOffset,
			SpineProfile:   skin.SpineProfile,
			Tag:            skin.Tag,
			Time:           skin.Time,
			GetShowing:     skin.GetShowing,
			PurchaseOffset: skin.PurchaseOffset,
			ShopOffset:     skin.ShopOffset,
			RarityBg:       pgText(skin.RarityBG),
			SpecialEffects: skin.SpecialEffects,
			GroupIndex:     pgInt4FromPtr(skin.GroupIndex),
			Gyro:           pgInt4FromPtr(skin.Gyro),
			HandID:         pgInt4FromPtr(skin.HandID),
			Illustrator:    pgInt4FromPtr(skin.Illustrator),
			Illustrator2:   pgInt4FromPtr(skin.Illustrator2),
			VoiceActor:     pgInt4FromPtr(skin.VoiceActor),
			VoiceActor2:    pgInt4FromPtr(skin.VoiceActor2),
			DoubleChar:     pgInt4FromPtr(skin.DoubleChar),
			LipSmoothing:   pgInt4FromPtr(skin.LipSmoothing),
			LipSyncGain:    pgInt4FromPtr(skin.LipSyncGain),
			L2DIgnoreDrag:  pgInt4FromPtr(skin.L2DIgnoreDrag),
			SkinType:       pgInt4FromPtr(skin.SkinType),
			ShopID:         pgInt4FromPtr(skin.ShopID),
			ShopTypeID:     pgInt4FromPtr(skin.ShopTypeID),
			ShopDynamicHx:  pgInt4FromPtr(skin.ShopDynamicHX),
			SpineAction:    skin.SpineAction,
			SpineUseLive2D: pgInt4FromPtr(skin.SpineUseLive2D),
			Live2DOffset:   skin.Live2DOffset,
			Live2DProfile:  skin.Live2DProfile,
			FxContainer:    skin.FXContainer,
			BoundBone:      skin.BoundBone,
			Smoke:          skin.Smoke,
		}); err != nil {
			return err
		}
	}
	return nil
}

func importResourcesSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/player_resource.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var resource orm.Resource
		if err := decoder.Decode(&resource); err != nil {
			return err
		}
		if err := q.UpsertResource(ctx, gen.UpsertResourceParams{
			ID:     int64(resource.ID),
			ItemID: int64(resource.ItemID),
			Name:   resource.Name,
		}); err != nil {
			return err
		}
	}
	return nil
}

func importPoolsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData("", "build_pools.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var pool struct {
			ID   uint32 `json:"id"`
			Pool uint32 `json:"pool"`
		}
		if err := decoder.Decode(&pool); err != nil {
			return err
		}
		tag, err := q.SetShipPoolID(ctx, gen.SetShipPoolIDParams{
			TemplateID: int64(pool.ID),
			PoolID:     pgtype.Int8{Int64: int64(pool.Pool), Valid: true},
		})
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("ship not found for template_id=%d", pool.ID)
		}
	}
	return nil
}

func importRequisitionShipsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData("", "requisition_ships.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var shipIDs []uint32
	if err := decoder.Decode(&shipIDs); err != nil {
		return err
	}
	for _, shipID := range shipIDs {
		if err := q.UpsertRequisitionShip(ctx, int64(shipID)); err != nil {
			return err
		}
	}
	return nil
}

func importBuildTimesSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData("", "build_times.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buildTimes map[string]uint32
	if err := decoder.Decode(&buildTimes); err != nil {
		return err
	}
	for id, timeValue := range buildTimes {
		parsed, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
		tag, err := q.SetShipBuildTime(ctx, gen.SetShipBuildTimeParams{
			TemplateID: parsed,
			BuildTime:  int64(timeValue),
		})
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("ship not found for template_id=%s", id)
		}
	}
	return nil
}

func importShopOffersSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "sharecfgdata/shop_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var offer orm.ShopOffer
		if err := decoder.Decode(&offer); err != nil {
			return err
		}
		var effects []uint32
		if err := json.Unmarshal(offer.EffectArgs, &effects); err == nil {
			offer.Effects = orm.ToInt64List(effects)
		}
		effectsJSON, err := json.Marshal(offer.Effects)
		if err != nil {
			return err
		}
		if err := q.UpsertShopOffer(ctx, gen.UpsertShopOfferParams{
			ID:             int64(offer.ID),
			Effects:        effectsJSON,
			EffectArgs:     offer.EffectArgs,
			Number:         int32(offer.Number),
			ResourceNumber: int32(offer.ResourceNumber),
			ResourceID:     int64(offer.ResourceID),
			Type:           int64(offer.Type),
			Genre:          offer.Genre,
			Discount:       int32(offer.Discount),
		}); err != nil {
			return err
		}
	}
	return nil
}

func importWeaponsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "sharecfgdata/weapon_property.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var weapon orm.Weapon
		if err := decoder.Decode(&weapon); err != nil {
			return err
		}
		if err := q.UpsertWeapon(ctx, gen.UpsertWeaponParams{
			ID:                   int64(weapon.ID),
			ActionIndex:          weapon.ActionIndex,
			AimType:              int32(weapon.AimType),
			Angle:                int32(weapon.Angle),
			AttackAttribute:      int32(weapon.AttackAttribute),
			AttackAttributeRatio: int32(weapon.AttackAttributeRatio),
			AutoAftercast:        weapon.AutoAftercast,
			AxisAngle:            int32(weapon.AxisAngle),
			BarrageID:            weapon.BarrageID,
			BulletID:             weapon.BulletID,
			ChargeParam:          weapon.ChargeParam,
			Corrected:            int32(weapon.Corrected),
			Damage:               int32(weapon.Damage),
			EffectMove:           int32(weapon.EffectMove),
			Expose:               int32(weapon.Expose),
			FireFx:               pgText(weapon.FireFX),
			FireFxLoopType:       int32(weapon.FireFXLoopType),
			FireSfx:              pgText(weapon.FireSFX),
			InitialOverHeat:      int32(weapon.InitialOverHeat),
			MinRange:             int32(weapon.MinRange),
			OxyType:              weapon.OxyType,
			PrecastParam:         weapon.PrecastParam,
			Queue:                int32(weapon.Queue),
			Range:                int32(weapon.Range),
			RecoverTime:          weapon.RecoverTime,
			ReloadMax:            int32(weapon.ReloadMax),
			SearchCondition:      weapon.SearchCondition,
			SearchType:           int32(weapon.SearchType),
			ShakeScreen:          int32(weapon.ShakeScreen),
			SpawnBound:           weapon.SpawnBound,
			Suppress:             int32(weapon.Suppress),
			TorpedoAmmo:          int32(weapon.TorpedoAmmo),
			Type:                 int32(weapon.Type),
		}); err != nil {
			return err
		}
	}
	return nil
}

func importEquipmentsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "sharecfgdata/equip_data_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var equip orm.Equipment
		if err := decoder.Decode(&equip); err != nil {
			return err
		}
		if err := q.UpsertEquipment(ctx, gen.UpsertEquipmentParams{
			ID:                int64(equip.ID),
			Base:              pgInt8FromUint32Ptr(equip.Base),
			DestroyGold:       int64(equip.DestroyGold),
			DestroyItem:       equip.DestroyItem,
			EquipLimit:        int32(equip.EquipLimit),
			Group:             int64(equip.Group),
			Important:         int64(equip.Important),
			Level:             int64(equip.Level),
			Next:              int64(equip.Next),
			Prev:              int64(equip.Prev),
			RestoreGold:       int64(equip.RestoreGold),
			RestoreItem:       equip.RestoreItem,
			ShipTypeForbidden: equip.ShipTypeForbidden,
			TransUseGold:      int64(equip.TransUseGold),
			TransUseItem:      equip.TransUseItem,
			Type:              int64(equip.Type),
			UpgradeFormulaID:  equip.UpgradeFormulaID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func importSkillsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "GameCfg/skill.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var skillMap map[string]orm.Skill
	if err := decoder.Decode(&skillMap); err != nil {
		return err
	}
	for _, skill := range skillMap {
		if err := q.UpsertSkill(ctx, gen.UpsertSkillParams{
			ID:         int64(skill.ID),
			Name:       skill.Name,
			Desc:       pgText(skill.Desc),
			Cd:         int64(skill.CD),
			Painting:   skill.Painting,
			Picture:    pgText(skill.Picture),
			AniEffect:  skill.AniEffect,
			UiEffect:   pgText(skill.UIEffect),
			EffectList: skill.EffectList,
		}); err != nil {
			return err
		}
	}
	return nil
}

func importConfigEntriesSQLC(ctx context.Context, region string, q *gen.Queries) error {
	shareCfgFiles, err := listBelfastDataFiles(region, "ShareCfg")
	if err != nil {
		return err
	}
	gameCfgFiles, err := listBelfastDataFiles(region, "GameCfg")
	if err != nil {
		return err
	}
	shareCfgFiles = filterConfigFiles(shareCfgFiles,
		[]string{
			"activity_",
			"child_",
			"child2_",
			"dorm_",
			"furniture_",
			"navalacademy_",
			"spweapon_",
			"equip_skin_",
			"fleet_tech_",
			"ship_meta_",
			"technology_",
			"shop_",
		},
		[]string{
			"ShareCfg/tutorial_handbook.json",
			"ShareCfg/tutorial_handbook_task.json",
			"ShareCfg/game_room_template.json",
			"ShareCfg/gameroom_shop_template.json",
			"ShareCfg/backyard_theme_template.json",
			"ShareCfg/gameset.json",
			"ShareCfg/benefit_buff_template.json",
			"sharecfgdata/item_data_statistics.json",
			"sharecfgdata/item_virtual_data_statistics.json",
			"sharecfgdata/chapter_template.json",
			"sharecfgdata/chapter_template_loop.json",
			"sharecfgdata/ship_data_template.json",
			"ShareCfg/item_data_frame.json",
			"ShareCfg/item_data_chat.json",
			"ShareCfg/item_data_battleui.json",
			"ShareCfg/drop_data_restore.json",
			"ShareCfg/livingarea_cover.json",
			"ShareCfg/oilfield_template.json",
			"ShareCfg/class_upgrade_template.json",
			"ShareCfg/ship_data_blueprint.json",
			"ShareCfg/ship_data_strengthen.json",
			"ShareCfg/ship_strengthen_blueprint.json",
			"ShareCfg/ship_strengthen_meta.json",
			"ShareCfg/transform_data_template.json",
			"ShareCfg/compose_data_template.json",
			"ShareCfg/equip_upgrade_data.json",
			"ShareCfg/month_shop_template.json",
			"ShareCfg/medal_template.json",
			"ShareCfg/newserver_shop_template.json",
			"ShareCfg/blackfriday_shop_template.json",
			"ShareCfg/guild_store.json",
			"ShareCfg/guildset.json",
			"ShareCfg/shop_template.json",
			"ShareCfg/quota_shop_template.json",
			"ShareCfg/recommend_shop.json",
			"ShareCfg/re_map_template.json",
			"ShareCfg/escort_template.json",
			"ShareCfg/escort_map_template.json",
			"ShareCfg/shop_banner_template.json",
			"ShareCfg/shop_discount_coupon_template.json",
			"ShareCfg/emoji_template.json",
			"sharecfgdata/expedition_data_template.json",
			"ShareCfg/ship_level.json",
			"ShareCfg/user_level.json",
		},
	)
	gameCfgFiles = filterConfigFiles(gameCfgFiles, nil, []string{
		"GameCfg/dorm.json",
	})
	for _, file := range append(shareCfgFiles, gameCfgFiles...) {
		if err := importConfigEntriesFromFileSQLC(ctx, region, file, q); err != nil {
			return err
		}
	}
	return nil
}

func importConfigEntriesFromFileSQLC(ctx context.Context, region string, file string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, file)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	firstToken, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := firstToken.(json.Delim)
	if !ok {
		return fmt.Errorf("unexpected json token in %s", file)
	}
	if delim == '[' {
		index := 0
		for decoder.More() {
			var raw json.RawMessage
			if err := decoder.Decode(&raw); err != nil {
				return err
			}
			key := configEntryKey(raw, index)
			if err := q.UpsertConfigEntry(ctx, gen.UpsertConfigEntryParams{Category: file, Key: key, Data: raw}); err != nil {
				return err
			}
			index++
		}
		return nil
	}
	if delim != '{' {
		return fmt.Errorf("unexpected json delimiter in %s", file)
	}
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("unexpected key in %s", file)
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return err
		}
		if err := q.UpsertConfigEntry(ctx, gen.UpsertConfigEntryParams{Category: file, Key: key, Data: raw}); err != nil {
			return err
		}
	}
	return nil
}

func importJuustagramTemplatesSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/activity_ins_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var template orm.JuustagramTemplate
		if err := decoder.Decode(&template); err != nil {
			return err
		}
		npcDiscussPersist, err := json.Marshal(template.NpcDiscussPersist)
		if err != nil {
			return err
		}
		timeValue, err := json.Marshal(template.Time)
		if err != nil {
			return err
		}
		timePersistValue, err := json.Marshal(template.TimePersist)
		if err != nil {
			return err
		}
		if err := q.UpsertJuustagramTemplate(ctx, gen.UpsertJuustagramTemplateParams{
			ID:                int64(template.ID),
			GroupID:           int64(template.GroupID),
			ShipGroup:         int64(template.ShipGroup),
			Name:              template.Name,
			Sculpture:         template.Sculpture,
			PicturePersist:    template.PicturePersist,
			MessagePersist:    template.MessagePersist,
			IsActive:          int64(template.IsActive),
			NpcDiscussPersist: string(npcDiscussPersist),
			Time:              string(timeValue),
			TimePersist:       string(timePersistValue),
		}); err != nil {
			return err
		}
	}
	return nil
}

func importJuustagramNpcTemplatesSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/activity_ins_npc_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var template orm.JuustagramNpcTemplate
		if err := decoder.Decode(&template); err != nil {
			return err
		}
		npcReplyPersist, err := json.Marshal(template.NpcReplyPersist)
		if err != nil {
			return err
		}
		timePersist, err := json.Marshal(template.TimePersist)
		if err != nil {
			return err
		}
		if err := q.UpsertJuustagramNpcTemplate(ctx, gen.UpsertJuustagramNpcTemplateParams{
			ID:              int64(template.ID),
			ShipGroup:       int64(template.ShipGroup),
			MessagePersist:  template.MessagePersist,
			NpcReplyPersist: string(npcReplyPersist),
			TimePersist:     string(timePersist),
		}); err != nil {
			return err
		}
	}
	return nil
}

func importJuustagramLanguageSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/activity_ins_language.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var entries map[string]struct {
		Value string `json:"value"`
	}
	if err := decoder.Decode(&entries); err != nil {
		return err
	}
	for key, entry := range entries {
		if err := q.UpsertJuustagramLanguage(ctx, gen.UpsertJuustagramLanguageParams{Key: key, Value: entry.Value}); err != nil {
			return err
		}
	}
	return nil
}

func importJuustagramShipGroupsSQLC(ctx context.Context, region string, q *gen.Queries) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/activity_ins_ship_group_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token()
	for decoder.More() {
		var template orm.JuustagramShipGroupTemplate
		if err := decoder.Decode(&template); err != nil {
			return err
		}
		if err := q.UpsertJuustagramShipGroupTemplate(ctx, gen.UpsertJuustagramShipGroupTemplateParams{
			ShipGroup:   int64(template.ShipGroup),
			Name:        template.Name,
			Background:  template.Background,
			Sculpture:   template.Sculpture,
			SculptureIi: template.SculptureII,
			Nationality: int64(template.Nationality),
			Type:        int64(template.Type),
		}); err != nil {
			return err
		}
	}
	return nil
}

func pgText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func pgInt4FromPtr(value *int) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*value), Valid: true}
}

func pgInt8FromUint32Ptr(value *uint32) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(*value), Valid: true}
}
