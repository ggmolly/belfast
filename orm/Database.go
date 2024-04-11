package orm

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bettercallmolly/belfast/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	GormDB *gorm.DB
)

func InitDatabase() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Europe/Paris",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASS"),
		os.Getenv("POSTGRES_DB"),
	)
	var err error
	GormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic("failed to connect database " + err.Error())
	}

	err = GormDB.AutoMigrate(
		// Types
		&Item{},
		&Skin{},
		&Notice{},
		&Build{},
		&ShipType{},
		&Ship{},
		&Rarity{},
		&Buff{},
		&Resource{},
		&Mail{},
		&MailAttachment{},
		&ShopOffer{},
		&Message{},
		// Servers
		&ServerState{},
		&Server{},
		// Debug stuff
		&DebugName{},
		&Debug{},
		// Commander related stuff
		&YostarusMap{},
		&OwnedShip{},
		&OwnedSkin{},
		&Punishment{},
		&Commander{},
		&CommanderItem{},
		// &CommanderLimitItem{},
		&CommanderMiscItem{},
		&OwnedResource{},
	)
	if err != nil {
		panic("failed to migrate database " + err.Error())
	}
	// Count number of debug names
	var count int64
	GormDB.Model(&DebugName{}).Count(&count)
	if count == 0 {
		logger.LogEvent("ORM", "Populating", "Debug names table is empty, populating...", logger.LOG_LEVEL_INFO)
		// Execute python3 script from the _tools folder
		cmd := exec.Command("python3", "insert_packet_names.py")
		cmd.Dir = "_tools"
		if err := cmd.Run(); err != nil {
			logger.LogEvent("ORM", "Populating", "Failed to populate debug names table", logger.LOG_LEVEL_ERROR)
		} else {
			logger.LogEvent("ORM", "Populating", "Debug names table populated!", logger.LOG_LEVEL_INFO)
		}
	}
}
