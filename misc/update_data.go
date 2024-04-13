package misc

import (
	"fmt"
	"os/exec"

	"github.com/ggmolly/belfast/logger"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	ITEMS_PATH = "/home/molly/Documents/al-zero/AzurLaneData/EN/sharecfgdata/item_data_statistics.json"
	BUFF_PATH  = "/home/molly/Documents/al-zero/AzurLaneData/EN/ShareCfg/benefit_buff_template.json"
	SHIP_PATH  = "/home/molly/Documents/al-zero/AzurLaneData/EN/sharecfgdata/ship_data_statistics.json"
	SKIN_PATH  = "/home/molly/Documents/al-zero/AzurLaneData/EN/ShareCfg/ship_skin_template.json"
)

var (
	tablesUpdate = []string{
		"items",
		"buffs",
		"ships",
		"skins",
		"resources",
		"pools",
		"shop_offers",
	}
	caser = cases.Title(language.AmericanEnglish)
)

func UpdateAllData() {
	logger.LogEvent("GameData", "Updating", "Updating all game data.. this may take a while.", logger.LOG_LEVEL_INFO)
	go func() {
		for _, table := range tablesUpdate {
			if err := exec.Command("sh", "_tools/import.sh", table).Run(); err != nil {
				logger.LogEvent("GameData", caser.String(table), fmt.Sprintf("error importing %s : %s", table, err.Error()), logger.LOG_LEVEL_ERROR)
			} else {
				logger.LogEvent("GameData", caser.String(table), fmt.Sprintf("successfully imported %s!", table), logger.LOG_LEVEL_INFO)
			}
		}
	}()
}
