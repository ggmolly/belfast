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
