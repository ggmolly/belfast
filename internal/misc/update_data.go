package misc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// region / file
	URL_BASE            = "https://raw.githubusercontent.com/ggmolly/belfast-data/v2/%s/%s"
	REGIONLESS_URL_BASE = "https://raw.githubusercontent.com/ggmolly/belfast-data/v2/%s"
)

var (
	dataFn = map[string]func(string, *gorm.DB) error{
		"Items":       importItems,
		"Buffs":       importBuffs,
		"Ships":       importShips,
		"Skins":       importSkins,
		"Resources":   importResources,
		"Pools":       importPools,
		"Requisition": importRequisitionShips,
		"BuildTimes":  importBuildTimes,
		"ShopOffers":  importShopOffers,
		"Weapons":     importWeapons,
		"Equipments":  importEquipments,
		"Skills":      importSkills,
	}
	// Golang maps are unordered, so we need to keep track of order of keys ourselves
	order = []string{"Items", "Buffs", "Ships", "Skins", "Resources", "Pools", "Requisition", "BuildTimes", "ShopOffers", "Weapons", "Equipments", "Skills"}
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
