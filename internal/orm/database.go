package orm

import (
	"os"

	"github.com/ggmolly/belfast/internal/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	GormDB *gorm.DB
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
		&AdminUser{},
		&AdminSession{},
		&WebAuthnCredential{},
		&AuthChallenge{},
		&AdminAuditLog{},
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
		&CommanderStory{},
		&CommanderAttire{},
		&CommanderLivingAreaCover{},
		&CommanderItem{},
		&CommanderSurvey{},
		// &CommanderLimitItem{},
		&CommanderMiscItem{},
		&CommanderBuff{},
		&OwnedResource{},
		&Like{},
		&EquipCodeReport{},
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
