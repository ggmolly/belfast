package orm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	GormDB *gorm.DB
)

func InitDatabase() bool {
	if os.Getenv("MODE") == "test" {
		GormDB = initSqlite("file::memory:?cache=shared")
		return seedDatabase(true)
	}
	cfg := config.Current()
	driver := strings.ToLower(strings.TrimSpace(cfg.DB.Driver))
	if driver == "" {
		driver = "sqlite"
	}
	schemaName := strings.TrimSpace(cfg.DB.SchemaName)
	if schemaName != "" && driver != "sqlite" && driver != "sqlite3" {
		if err := validateIdentifier(schemaName); err != nil {
			panic(err.Error())
		}
	}
	if driver == "postgres" && schemaName != "" && schemaName != strings.ToLower(schemaName) {
		panic("database.schema_name must be lowercase for postgres")
	}

	switch driver {
	case "sqlite", "sqlite3":
		dsn := strings.TrimSpace(cfg.DB.Path)
		if dsn == "" {
			dsn = "data/belfast.db"
		}
		if err := ensureParentDir(dsn); err != nil {
			panic("failed to create database directory " + err.Error())
		}
		GormDB = initSqlite(dsn)
	case "postgres", "postgresql", "pg":
		dsn := strings.TrimSpace(cfg.DB.DSN)
		if dsn == "" {
			panic("database.dsn is required when database.driver=postgres")
		}
		GormDB = initPostgres(dsn, schemaName)
		if err := ensurePostgresSchema(GormDB, schemaName); err != nil {
			panic("failed to ensure postgres schema " + err.Error())
		}
	case "mysql":
		dsn := strings.TrimSpace(cfg.DB.DSN)
		if dsn == "" {
			panic("database.dsn is required when database.driver=mysql")
		}
		if schemaName != "" {
			updatedDSN, err := ensureMySQLDatabase(dsn, schemaName)
			if err != nil {
				panic("failed to ensure mysql database " + err.Error())
			}
			dsn = updatedDSN
		}
		GormDB = initMySQL(dsn)
	default:
		panic(fmt.Sprintf("unsupported database.driver %q", driver))
	}
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

func initPostgres(dsn string, schemaName string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:    true,
		NamingStrategy: schemaNamingStrategy(schemaName),
	})
	if err != nil {
		panic("failed to connect database " + err.Error())
	}
	return db
}

func initMySQL(dsn string) *gorm.DB {
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic("failed to connect database " + err.Error())
	}
	return db
}

func ensurePostgresSchema(db *gorm.DB, schemaName string) error {
	if schemaName == "" {
		return nil
	}
	return db.Exec("CREATE SCHEMA IF NOT EXISTS \"" + schemaName + "\"").Error
}

func ensureMySQLDatabase(dsn string, databaseName string) (string, error) {
	parsed, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}
	parsed.DBName = databaseName
	updatedDSN := parsed.FormatDSN()

	admin := *parsed
	admin.DBName = ""
	adminDSN := admin.FormatDSN()
	adminDB := initMySQL(adminDSN)
	defer func() {
		sqlDB, err := adminDB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}()

	if err := adminDB.Exec("CREATE DATABASE IF NOT EXISTS `" + databaseName + "`").Error; err != nil {
		return "", err
	}
	return updatedDSN, nil
}

func schemaNamingStrategy(schemaName string) schema.NamingStrategy {
	if schemaName == "" {
		return schema.NamingStrategy{}
	}
	return schema.NamingStrategy{TablePrefix: schemaName + "."}
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

var identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateIdentifier(value string) error {
	if value == "" {
		return nil
	}
	if !identifierPattern.MatchString(value) {
		return fmt.Errorf("invalid database.schema_name %q (expected [a-zA-Z_][a-zA-Z0-9_]*)", value)
	}
	return nil
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
		&Compensation{},
		&CompensationAttachment{},
		&ExchangeCode{},
		&ExchangeCodeRedeem{},
		&ShopOffer{},
		&ShoppingStreetState{},
		&ShoppingStreetGood{},
		&Message{},
		&GuildChatMessage{},
		&Fleet{},
		&ExerciseFleet{},
		&ArenaShopState{},
		&GuildShopState{},
		&GuildShopGood{},
		&MedalShopState{},
		&MedalShopGood{},
		&MiniGameShopState{},
		&MiniGameShopGood{},
		&ActivityFleet{},
		// Debug stuff
		&DebugName{},
		&Debug{},
		// Accounts
		&Account{},
		&Session{},
		&Role{},
		&Permission{},
		&RolePermission{},
		&AccountRole{},
		&AccountPermissionOverride{},
		&WebAuthnCredential{},
		&AuthChallenge{},
		&AuditLog{},
		&UserRegistrationChallenge{},
		// Commander related stuff
		&YostarusMap{},
		&LocalAccount{},
		&DeviceAuthMap{},
		&OwnedShip{},
		&OwnedSpWeapon{},
		&OwnedShipEquipment{},
		&OwnedShipTransform{},
		&OwnedShipStrength{},
		&OwnedSkin{},
		&OwnedEquipment{},
		&Punishment{},
		&Commander{},
		&CommanderCommonFlag{},
		&CommanderMedalDisplay{},
		&CommanderStory{},
		&CommanderAttire{},
		&CommanderLivingAreaCover{},
		&CommanderItem{},
		&CommanderSurvey{},
		&CommanderTrophyProgress{},
		&CommanderStoreupAwardProgress{},
		// &CommanderLimitItem{},
		&CommanderMiscItem{},
		&CommanderBuff{},
		&CommanderFurniture{},
		&CommanderSoundStory{},
		&MonthShopPurchase{},
		&OwnedResource{},
		&Like{},
		&EquipCodeShare{},
		&EquipCodeReport{},
		&EquipCodeLike{},
		&SecondaryPasswordState{},
		&RandomFlagShip{},
		&OwnedShipShadowSkin{},
		&JuustagramGroup{},
		&JuustagramChatGroup{},
		&JuustagramReply{},
		&JuustagramTemplate{},
		&JuustagramNpcTemplate{},
		&JuustagramLanguage{},
		&JuustagramShipGroupTemplate{},
		&JuustagramMessageState{},
		&JuustagramPlayerDiscuss{},
		&Dorm3dApartment{},
		&CommanderTB{},
		&CommanderAppreciationState{},
		&ActivityPermanentState{},
		&EventCollection{},
		&SurveyState{},
		&RefluxState{},
		&RemasterState{},
		&RemasterProgress{},
		&ChapterState{},
		&SubmarineExpeditionState{},
		&CommanderDormState{},
		&CommanderDormFloorLayout{},
		&CommanderDormTheme{},
		&BackyardCustomThemeTemplate{},
		&BackyardPublishedThemeVersion{},
		&BackyardThemeLike{},
		&BackyardThemeCollection{},
		&BackyardThemeInform{},
		&ChapterProgress{},
		&ChapterDrop{},
		&EscortState{},
		&BattleSession{},
		// Skin restrictions
		&GlobalSkinRestriction{},
		&GlobalSkinRestrictionWindow{},
		// Game data
		&Weapon{},
		&Equipment{},
		&Skill{},
		&RequisitionShip{},
		&ConfigEntry{},
	)
	if err != nil {
		panic("failed to migrate database " + err.Error())
	}
	if err := EnsureAuthzDefaults(); err != nil {
		panic("failed to seed authz defaults " + err.Error())
	}
	if skipSeed {
		logger.LogEvent("ORM", "Init", "Skipping database seeding in test mode", logger.LOG_LEVEL_INFO)
		return true
	}
	var count int64
	GormDB.Model(&Rarity{}).Count(&count)
	if count == 0 {
		logger.LogEvent("ORM", "Populating", "Adding rarities...", logger.LOG_LEVEL_INFO)
		GormDB.Save(&Rarity{
			ID:   2,
			Name: "Common",
		})
		GormDB.Save(&Rarity{
			ID:   3,
			Name: "Rare",
		})
		GormDB.Save(&Rarity{
			ID:   4,
			Name: "Elite",
		})
		GormDB.Save(&Rarity{
			ID:   5,
			Name: "Super Rare",
		})
		GormDB.Save(&Rarity{
			ID:   6,
			Name: "Ultra Rare",
		})
		return true
	}
	return false
}
