package orm

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/ggmolly/belfast/internal/logger"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	GormDB          *gorm.DB
	packetNameRegex = regexp.MustCompile(`^(CS|SC)_(\d+)`)
)

const (
	RootPacketDir = "protobuf"
)

func InitDatabase() bool {
	if os.Getenv("MODE") == "test" {
		GormDB = initSqlite("file::memory:?cache=shared")
		return seedDatabase(true)
	}
	if err := os.MkdirAll("data", 0o755); err != nil {
		panic("failed to create data directory " + err.Error())
	}
	GormDB = initSqlite("data/belfast.db")
	return seedDatabase(false)
}

func initSqlite(dsn string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic("failed to connect database " + err.Error())
	}
	return db
}

func seedDatabase(skipSeed bool) bool {
	err := GormDB.AutoMigrate(
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
		&Fleet{},
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
		&Like{},
	)
	if err != nil {
		panic("failed to migrate database " + err.Error())
	}
	if skipSeed {
		logger.LogEvent("ORM", "Init", "Skipping database seeding in test mode", logger.LOG_LEVEL_INFO)
		return true
	}
	// Pre-populate debug names table, user will be able to rename them later
	var count int64
	GormDB.Model(&DebugName{}).Count(&count)
	if count == 0 {
		logger.LogEvent("ORM", "Populating", "Debug names table is empty, populating...", logger.LOG_LEVEL_INFO)
		tx := GormDB.Begin()
		files, err := os.ReadDir(RootPacketDir)
		if err != nil {
			panic("failed to read directory " + err.Error())
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			// Match group 1 is the direction (CS/SC), group 2 is the packet ID
			matches := packetNameRegex.FindStringSubmatch(file.Name())
			if len(matches) != 3 {
				continue
			}
			packetID, _ := strconv.Atoi(matches[2])
			if err := tx.Save(&DebugName{
				ID:   packetID,
				Name: fmt.Sprintf("%s_%d", matches[1], packetID),
			}).Error; err != nil {
				panic("failed to save debug name " + err.Error())
			}
		}
		if err := tx.Commit().Error; err != nil {
			panic("failed to commit transaction " + err.Error())
		}
	}
	// Pre-populate the server table (if empty)
	GormDB.Model(&Server{}).Count(&count)
	if count == 0 {
		tx := GormDB.Begin()
		logger.LogEvent("ORM", "Populating", "Adding default server entry...", logger.LOG_LEVEL_INFO)
		// Create server states
		tx.Save(&ServerState{
			ID:          1,
			Color:       "success",
			Description: "Online",
		})
		tx.Save(&ServerState{
			ID:          2,
			Color:       "neutral",
			Description: "Offline",
		})
		tx.Save(&ServerState{
			ID:          3,
			Color:       "primary",
			Description: "Busy",
		})
		tx.Save(&ServerState{
			ID:          4,
			Color:       "accent",
			Description: "Full",
		})
		tx.Commit()
		tx = GormDB.Begin()
		// Create default servers
		tx.Save(&Server{
			ID:      1,
			Name:    "Belfast",
			IP:      "localhost",
			Port:    80,
			StateID: proto.Uint32(1),
		})
		tx.Save(&Server{
			ID:      2,
			Name:    "github.com/ggmolly/belfast",
			IP:      "localhost",
			Port:    80,
			StateID: proto.Uint32(2),
		})
		tx.Save(&Server{
			ID:      3,
			Name:    "https://belfast.mana.rip",
			IP:      "localhost",
			Port:    80,
			StateID: proto.Uint32(2),
		})

		logger.LogEvent("ORM", "Populating", "Adding rarities...", logger.LOG_LEVEL_INFO)
		tx.Save(&Rarity{
			ID:   2,
			Name: "Common",
		})
		tx.Save(&Rarity{
			ID:   3,
			Name: "Rare",
		})
		tx.Save(&Rarity{
			ID:   4,
			Name: "Elite",
		})
		tx.Save(&Rarity{
			ID:   5,
			Name: "Super Rare",
		})
		tx.Save(&Rarity{
			ID:   6,
			Name: "Ultra Rare",
		})
		tx.Commit()
	}
	return count == 0
}
