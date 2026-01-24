package misc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// region / file
	URL_BASE            = "https://raw.githubusercontent.com/ggmolly/belfast-data/main/%s/%s"
	REGIONLESS_URL_BASE = "https://raw.githubusercontent.com/ggmolly/belfast-data/main/%s"
)

var (
	dataFn = map[string]func(string, *gorm.DB) error{
		"Items":                  importItems,
		"Buffs":                  importBuffs,
		"Ships":                  importShips,
		"Skins":                  importSkins,
		"Resources":              importResources,
		"Pools":                  importPools,
		"Requisition":            importRequisitionShips,
		"BuildTimes":             importBuildTimes,
		"ShopOffers":             importShopOffers,
		"Weapons":                importWeapons,
		"Equipments":             importEquipments,
		"Skills":                 importSkills,
		"Configs":                importConfigEntries,
		"JuustagramTemplates":    importJuustagramTemplates,
		"JuustagramNpcTemplates": importJuustagramNpcTemplates,
		"JuustagramLanguage":     importJuustagramLanguage,
		"JuustagramShipGroups":   importJuustagramShipGroups,
	}
	// Golang maps are unordered, so we need to keep track of order of keys ourselves
	order = []string{"Items", "Buffs", "Ships", "Skins", "Resources", "Pools", "Requisition", "BuildTimes", "ShopOffers", "Weapons", "Equipments", "Skills", "Configs", "JuustagramTemplates", "JuustagramNpcTemplates", "JuustagramLanguage", "JuustagramShipGroups"}
)

func getBelfastData(region string, file string) (*json.Decoder, *http.Response, error) {
	var url string
	if region == "" {
		url = fmt.Sprintf(REGIONLESS_URL_BASE, file)
	} else {
		url = fmt.Sprintf(URL_BASE, region, file)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to fetch data from %s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body), resp, nil
}

type githubContent struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func listBelfastDataFiles(region string, directory string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/ggmolly/belfast-data/contents/%s/%s", region, directory)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list data from %s: %s", url, resp.Status)
	}
	var contents []githubContent
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&contents); err != nil {
		return nil, err
	}
	files := make([]string, 0, len(contents))
	for _, entry := range contents {
		if entry.Type != "file" || !strings.HasSuffix(entry.Name, ".json") {
			continue
		}
		files = append(files, fmt.Sprintf("%s/%s", directory, entry.Name))
	}
	return files, nil
}

func importConfigEntries(region string, tx *gorm.DB) error {
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
			"ShareCfg/item_data_frame.json",
			"ShareCfg/item_data_chat.json",
			"ShareCfg/item_data_battleui.json",
			"ShareCfg/livingarea_cover.json",
			"ShareCfg/oilfield_template.json",
			"ShareCfg/class_upgrade_template.json",
			"ShareCfg/ship_data_blueprint.json",
			"ShareCfg/ship_strengthen_blueprint.json",
			"ShareCfg/ship_strengthen_meta.json",
			"ShareCfg/month_shop_template.json",
			"ShareCfg/newserver_shop_template.json",
			"ShareCfg/blackfriday_shop_template.json",
			"ShareCfg/guild_store.json",
			"ShareCfg/guildset.json",
			"ShareCfg/shop_template.json",
			"ShareCfg/quota_shop_template.json",
			"ShareCfg/recommend_shop.json",
			"ShareCfg/shop_banner_template.json",
			"ShareCfg/shop_discount_coupon_template.json",
		},
	)
	gameCfgFiles = filterConfigFiles(gameCfgFiles, nil, []string{
		"GameCfg/dorm.json",
	})

	for _, file := range append(shareCfgFiles, gameCfgFiles...) {
		if err := importConfigEntriesFromFile(region, file, tx); err != nil {
			return err
		}
	}
	return nil
}

func filterConfigFiles(files []string, prefixes []string, include []string) []string {
	allowed := make(map[string]struct{}, len(files))
	for _, file := range files {
		name := strings.TrimPrefix(file, "ShareCfg/")
		name = strings.TrimPrefix(name, "GameCfg/")
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				allowed[file] = struct{}{}
				break
			}
		}
	}
	for _, file := range include {
		allowed[file] = struct{}{}
	}
	filtered := make([]string, 0, len(allowed))
	for file := range allowed {
		filtered = append(filtered, file)
	}
	return filtered
}

func importConfigEntriesFromFile(region string, file string, tx *gorm.DB) error {
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
			entry := orm.ConfigEntry{Category: file, Key: key, Data: raw}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "category"}, {Name: "key"}},
				UpdateAll: true,
			}).Create(&entry).Error; err != nil {
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
		entry := orm.ConfigEntry{Category: file, Key: key, Data: raw}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "category"}, {Name: "key"}},
			UpdateAll: true,
		}).Create(&entry).Error; err != nil {
			return err
		}
	}
	return nil
}

func configEntryKey(raw json.RawMessage, index int) string {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return strconv.Itoa(index)
	}
	for _, field := range []string{"id", "ID", "Id"} {
		value, ok := object[field]
		if !ok {
			continue
		}
		if key, ok := parseConfigKey(value); ok {
			return key
		}
	}
	return strconv.Itoa(index)
}

func parseConfigKey(value json.RawMessage) (string, bool) {
	var number uint64
	if err := json.Unmarshal(value, &number); err == nil {
		return strconv.FormatUint(number, 10), true
	}
	var text string
	if err := json.Unmarshal(value, &text); err == nil {
		return text, text != ""
	}
	var floatValue float64
	if err := json.Unmarshal(value, &floatValue); err == nil {
		return strconv.FormatUint(uint64(floatValue), 10), true
	}
	return "", false
}

// TODO: A lot of code duplication here, could be refactored

func importItems(region string, tx *gorm.DB) error {
	decoder, resp, err := getBelfastData(region, "sharecfgdata/item_data_statistics.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token() // Consume the start of the array '['

	// Decode each elements
	for decoder.More() {
		var item orm.Item
		if err := decoder.Decode(&item); err != nil {
			return err
		} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func importBuffs(region string, tx *gorm.DB) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/benefit_buff_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token() // Consume the start of the array '['

	// Decode each elements
	for decoder.More() {
		var buff orm.Buff
		if err := decoder.Decode(&buff); err != nil {
			return err
		} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&buff).Error; err != nil {
			return err
		}
	}
	return nil
}

func importShips(region string, tx *gorm.DB) error {
	decoder, resp, err := getBelfastData(region, "sharecfgdata/ship_data_statistics.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token() // Consume the start of the array '['

	// Decode each elements
	for decoder.More() {
		var ship orm.Ship
		if err := decoder.Decode(&ship); err != nil {
			return err
		} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&ship).Error; err != nil {
			return err
		}
	}
	return nil
}

func importSkins(region string, tx *gorm.DB) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/ship_skin_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token() // Consume the start of the array '['

	// Decode each elements
	for decoder.More() {
		var skin orm.Skin
		if err := decoder.Decode(&skin); err != nil {
			return err
		} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&skin).Error; err != nil {
			return err
		}
	}
	return nil
}

func importResources(region string, tx *gorm.DB) error {
	decoder, resp, err := getBelfastData(region, "ShareCfg/player_resource.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token() // Consume the start of the array '['

	// Decode each elements
	for decoder.More() {
		var resource orm.Resource
		if err := decoder.Decode(&resource); err != nil {
			return err
		} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&resource).Error; err != nil {
			return err
		}
	}
	return nil
}

func importPools(region string, tx *gorm.DB) error {
	decoder, resp, err := getBelfastData("", "build_pools.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// [{"id": 101451, "pool": 2}, {"id": 702011, "pool": 2}, {"id": 101491, "pool": 2}, {"id": 702031, "pool": 2}...]
	decoder.Token() // Consume the start of the array '['

	// Decode each elements
	for decoder.More() {
		var pool struct {
			ID   uint32 `json:"id"`
			Pool uint32 `json:"pool"`
		}
		if err := decoder.Decode(&pool); err != nil {
			return err
		}

		// Update each ship with the pool
		var ship orm.Ship
		if err := tx.Where("template_id = ?", pool.ID).First(&ship).Error; err != nil {
			return err
		}
		ship.PoolID = proto.Uint32(pool.Pool)
		if err := tx.Save(&ship).Error; err != nil {
			return err
		}
	}
	return nil
}

func importRequisitionShips(region string, tx *gorm.DB) error {
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
		entry := orm.RequisitionShip{ShipID: shipID}
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&entry).Error; err != nil {
			return err
		}
	}
	return nil
}

func importBuildTimes(region string, tx *gorm.DB) error {
	decoder, resp, err := getBelfastData("", "build_times.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// {"101031": 1380, "101041": 1380, "101061": 1500, "101071": 1500, ...}
	var buildTimes map[string]uint32
	if err := decoder.Decode(&buildTimes); err != nil {
		return err
	}

	// Update each ship with the build time
	for id, time := range buildTimes {
		var ship orm.Ship
		if err := tx.Where("template_id = ?", id).First(&ship).Error; err != nil {
			return err
		}
		ship.BuildTime = time
		if err := tx.Save(&ship).Error; err != nil {
			return err
		}
	}
	return nil
}

func importShopOffers(region string, tx *gorm.DB) error {
	decoder, resp, err := getBelfastData(region, "sharecfgdata/shop_template.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder.Token() // Consume of start of array '['

	// Decode each elements
	for decoder.More() {
		var offer orm.ShopOffer
		if err := decoder.Decode(&offer); err != nil {
			return err
		}
		var effects []uint32
		if err := json.Unmarshal(offer.EffectArgs, &effects); err == nil {
			offer.Effects = make([]int64, len(effects))
			for i, effect := range effects {
				offer.Effects[i] = int64(effect)
			}
		}
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&offer).Error; err != nil {
			return err
		}
	}
	return nil
}

func importWeapons(region string, tx *gorm.DB) error {
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
		} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&weapon).Error; err != nil {
			return err
		}
	}
	return nil
}

func importEquipments(region string, tx *gorm.DB) error {
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
		} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&equip).Error; err != nil {
			return err
		}
	}
	return nil
}

func importSkills(region string, tx *gorm.DB) error {
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
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&skill).Error; err != nil {
			return err
		}
	}
	return nil
}

func importJuustagramTemplates(region string, tx *gorm.DB) error {
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
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&template).Error; err != nil {
			return err
		}
	}
	return nil
}

func importJuustagramNpcTemplates(region string, tx *gorm.DB) error {
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
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&template).Error; err != nil {
			return err
		}
	}
	return nil
}

func importJuustagramLanguage(region string, tx *gorm.DB) error {
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
		item := orm.JuustagramLanguage{
			Key:   key,
			Value: entry.Value,
		}
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func importJuustagramShipGroups(region string, tx *gorm.DB) error {
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
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&template).Error; err != nil {
			return err
		}
	}
	return nil
}

// XXX: The database can end in a limbo state if an error occurs while updating data (e.g. network error, invalid JSON, etc.)
// upon restarting Belfast, database won't be re-populated because some tables were already populated
// this could be fixed by passing a single transaction to all the data import functions, but requires some refactoring
// due to way data is being initialized (mix of 'misc' and 'orm' packages)
func UpdateAllData(region string) {
	logger.LogEvent("GameData", "Updating", "Updating all game data.. this may take a while.", logger.LOG_LEVEL_INFO)
	tx := orm.GormDB.Begin()
	for _, key := range order {
		fn := dataFn[key]
		logger.LogEvent("GameData", "Updating", fmt.Sprintf("Updating %s (region=%s)", key, region), logger.LOG_LEVEL_INFO)
		// defer func() {
		// 	if r := recover(); r != nil {
		// 		logger.LogEvent("GameData", "Updating", fmt.Sprintf("panic occurred while updating %s: %v", key, r), logger.LOG_LEVEL_ERROR)
		// 		tx.Rollback()
		// 	}
		// }()
		if err := fn(region, tx); err != nil {
			logger.LogEvent("GameData", "Updating", fmt.Sprintf("failed to update %s: %s", key, err.Error()), logger.LOG_LEVEL_ERROR)
			tx.Rollback()
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		logger.LogEvent("GameData", "Updating", fmt.Sprintf("failed to commit transaction: %s", err.Error()), logger.LOG_LEVEL_ERROR)
	}
}
