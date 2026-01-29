package orm

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/rng"
)

func TestCommanderBeforeSaveCapsLevel(t *testing.T) {
	commander := Commander{Level: 130}
	if err := commander.BeforeSave(GormDB); err != nil {
		t.Fatalf("before save: %v", err)
	}
	if commander.Level != 120 {
		t.Fatalf("expected level capped at 120, got %d", commander.Level)
	}
}

func TestCommanderCreateBuild(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Build{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})

	originalRng := shipRng
	shipRng = rng.NewLockedRandFromSeed(1)
	defer func() { shipRng = originalRng }()

	for _, rarity := range []uint32{2, 3, 4, 5} {
		ship := Ship{TemplateID: rarity + 1000, Name: "Ship", EnglishName: "Ship", RarityID: rarity, Star: 1, Type: 1, Nationality: 1, BuildTime: 60, PoolID: uint32Ptr(1)}
		if err := GormDB.Create(&ship).Error; err != nil {
			t.Fatalf("seed ship: %v", err)
		}
	}
	commander := Commander{CommanderID: 60, AccountID: 60, Name: "Builder"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.BuildsMap = make(map[uint32]*Build)
	var running int
	build, buildTime, err := commander.CreateBuild(1, &running)
	if err != nil {
		t.Fatalf("create build: %v", err)
	}
	if build == nil || buildTime != 60 {
		t.Fatalf("unexpected build result")
	}
	if running != 1 {
		t.Fatalf("expected running build count 1, got %d", running)
	}
	if commander.BuildsMap[build.ID] == nil {
		t.Fatalf("expected build in map")
	}
}

func TestCommanderAddShipAndTx(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedShip{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})

	shipA := Ship{TemplateID: 4001, Name: "ShipA", EnglishName: "ShipA", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	shipB := Ship{TemplateID: 4002, Name: "ShipB", EnglishName: "ShipB", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := GormDB.Create(&shipA).Error; err != nil {
		t.Fatalf("seed shipA: %v", err)
	}
	if err := GormDB.Create(&shipB).Error; err != nil {
		t.Fatalf("seed shipB: %v", err)
	}
	commander := Commander{CommanderID: 61, AccountID: 61, Name: "Owner"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.OwnedShipsMap = make(map[uint32]*OwnedShip)
	if _, err := commander.AddShip(shipA.TemplateID); err != nil {
		t.Fatalf("add ship: %v", err)
	}
	if len(commander.Ships) != 1 {
		t.Fatalf("expected 1 owned ship")
	}
	if commander.OwnedShipsMap[commander.Ships[0].ID] == nil {
		t.Fatalf("expected ship in map")
	}

	tx := GormDB.Begin()
	if _, err := commander.AddShipTx(tx, shipB.TemplateID); err != nil {
		tx.Rollback()
		t.Fatalf("add ship tx: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit tx: %v", err)
	}
}

func TestCommanderConsumeItemTxAndSaveTx(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderItem{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 62, AccountID: 62, Name: "Items"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	item := CommanderItem{CommanderID: commander.CommanderID, ItemID: 9001, Count: 2}
	if err := GormDB.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
	commander.CommanderItemsMap = map[uint32]*CommanderItem{9001: &item}

	tx := GormDB.Begin()
	if err := commander.ConsumeItemTx(tx, 9001, 1); err != nil {
		tx.Rollback()
		t.Fatalf("consume item tx: %v", err)
	}
	commander.Name = "Updated"
	if err := commander.SaveTx(tx); err != nil {
		tx.Rollback()
		t.Fatalf("save tx: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit tx: %v", err)
	}
}

func TestCommanderGetItemResourceAndCounts(t *testing.T) {
	commander := Commander{}
	commander.CommanderItemsMap = map[uint32]*CommanderItem{1: {ItemID: 1, Count: 3}}
	commander.OwnedResourcesMap = map[uint32]*OwnedResource{4: {ResourceID: 4, Amount: 7}}

	if _, err := commander.GetItem(1); err != nil {
		t.Fatalf("expected item present")
	}
	if _, err := commander.GetItem(2); err == nil {
		t.Fatalf("expected missing item error")
	}
	if _, err := commander.GetResource(4); err != nil {
		t.Fatalf("expected resource present")
	}
	if _, err := commander.GetResource(5); err == nil {
		t.Fatalf("expected missing resource error")
	}
	if commander.GetResourceCount(14) != 7 {
		t.Fatalf("expected alias resource count")
	}
}

func TestCommanderPunishAndRevoke(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Punishment{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 63, AccountID: 63, Name: "Punish"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	if err := commander.Punish(nil, false); err != nil {
		t.Fatalf("punish: %v", err)
	}
	if err := commander.RevokeActivePunishment(); err != nil {
		t.Fatalf("revoke punish: %v", err)
	}
	var count int64
	if err := GormDB.Model(&Punishment{}).Where("punished_id = ?", commander.CommanderID).Count(&count).Error; err != nil {
		t.Fatalf("count punishments: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected punishments removed")
	}
}

func TestCommanderCommitAndBuildRange(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Build{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 64, AccountID: 64, Name: "Commit"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.Name = "Commit Updated"
	if err := commander.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	for i := 0; i < 3; i++ {
		build := Build{BuilderID: commander.CommanderID, ShipID: 1, PoolID: 1, FinishesAt: time.Now()}
		if err := GormDB.Create(&build).Error; err != nil {
			t.Fatalf("seed build: %v", err)
		}
	}
	builds, err := commander.GetBuildRange(0, 1)
	if err != nil {
		t.Fatalf("get build range: %v", err)
	}
	if len(builds) != 2 {
		t.Fatalf("expected 2 builds, got %d", len(builds))
	}
}

func TestCommanderBumpLastLogin(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 65, AccountID: 65, Name: "Login"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	if err := commander.BumpLastLogin(); err != nil {
		t.Fatalf("bump last login: %v", err)
	}
	if commander.LastLogin.IsZero() {
		t.Fatalf("expected last login set")
	}
}

func TestCommanderGetSecretaries(t *testing.T) {
	commander := Commander{}
	pos0 := uint32(0)
	pos1 := uint32(1)
	commander.Ships = []OwnedShip{
		{ID: 1, IsSecretary: true, SecretaryPosition: &pos1},
		{ID: 2, IsSecretary: true, SecretaryPosition: &pos0},
		{ID: 3, IsSecretary: false},
	}
	secretaries := commander.GetSecretaries()
	if len(secretaries) != 2 {
		t.Fatalf("expected 2 secretaries, got %d", len(secretaries))
	}
	if secretaries[0].ID != 2 {
		t.Fatalf("expected secretary order by position")
	}
}

func TestCommanderGiveSkinAndExpiry(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedSkin{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 66, AccountID: 66, Name: "Skins"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.OwnedSkinsMap = make(map[uint32]*OwnedSkin)
	if err := commander.GiveSkin(7001); err != nil {
		t.Fatalf("give skin: %v", err)
	}
	if commander.OwnedSkinsMap[7001] == nil {
		t.Fatalf("expected skin in map")
	}
	if err := commander.GiveSkin(7001); err != nil {
		t.Fatalf("expected duplicate give skin to succeed")
	}

	old := time.Now().Add(1 * time.Hour)
	owned := OwnedSkin{CommanderID: commander.CommanderID, SkinID: 8001, ExpiresAt: &old}
	if err := GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned skin: %v", err)
	}
	commander.OwnedSkinsMap[8001] = &owned
	newExpiry := time.Now().Add(2 * time.Hour)
	if err := commander.GiveSkinWithExpiry(8001, &newExpiry); err != nil {
		t.Fatalf("give skin expiry: %v", err)
	}
	if commander.OwnedSkinsMap[8001].ExpiresAt == nil || !commander.OwnedSkinsMap[8001].ExpiresAt.Equal(newExpiry) {
		t.Fatalf("expected expiry updated")
	}
}

func TestCommanderMailboxAndSendMail(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Mail{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 67, AccountID: 67, Name: "Mail"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	mail := Mail{Title: "Hello", Body: "World"}
	if err := commander.SendMail(&mail); err != nil {
		t.Fatalf("send mail: %v", err)
	}
	if mail.ReceiverID != commander.CommanderID {
		t.Fatalf("expected receiver id set")
	}
	if commander.MailsMap == nil || commander.MailsMap[mail.ID] == nil {
		t.Fatalf("expected mail in map")
	}
	if err := commander.CleanMailbox(); err != nil {
		t.Fatalf("clean mailbox: %v", err)
	}
	var count int64
	if err := GormDB.Model(&Mail{}).Where("receiver_id = ?", commander.CommanderID).Count(&count).Error; err != nil {
		t.Fatalf("count mails: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected mailbox cleared")
	}
}

func TestCommanderDestroyAndRetireShips(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedShip{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})
	clearTable(t, &CommanderItem{})
	clearTable(t, &OwnedResource{})

	commander := Commander{CommanderID: 68, AccountID: 68, Name: "Dock"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.OwnedShipsMap = make(map[uint32]*OwnedShip)
	commander.CommanderItemsMap = make(map[uint32]*CommanderItem)

	ship := Ship{TemplateID: 9001, Name: "Ship", EnglishName: "Ship", RarityID: 3, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	owned := OwnedShip{ID: 1, OwnerID: commander.CommanderID, ShipID: ship.TemplateID, Ship: ship}
	if err := GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	commander.Ships = []OwnedShip{owned}
	commander.OwnedShipsMap[owned.ID] = &commander.Ships[0]

	shipIDs := []uint32{owned.ID}
	if err := commander.RetireShips(&shipIDs); err != nil {
		t.Fatalf("retire ships: %v", err)
	}
	var count int64
	if err := GormDB.Model(&OwnedShip{}).Where("owner_id = ?", commander.CommanderID).Count(&count).Error; err != nil {
		t.Fatalf("count owned ships: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected ships destroyed")
	}

	owned2 := OwnedShip{ID: 2, OwnerID: commander.CommanderID, ShipID: ship.TemplateID, Ship: Ship{Type: 99, RarityID: 3}}
	commander.OwnedShipsMap[2] = &owned2
	shipIDs = []uint32{2}
	if err := commander.RetireShips(&shipIDs); err == nil {
		t.Fatalf("expected error for unknown ship type")
	}
}

func TestCommanderProposeAndRoomAndSecretaries(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedShip{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})
	clearTable(t, &CommanderItem{})

	commander := Commander{CommanderID: 69, AccountID: 69, Name: "Proposer"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.OwnedShipsMap = make(map[uint32]*OwnedShip)
	commander.CommanderItemsMap = make(map[uint32]*CommanderItem)

	ship := Ship{TemplateID: 9101, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	owned := OwnedShip{ID: 3, OwnerID: commander.CommanderID, ShipID: ship.TemplateID, Ship: ship}
	if err := GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	commander.Ships = []OwnedShip{owned}
	commander.OwnedShipsMap[owned.ID] = &commander.Ships[0]

	ring := CommanderItem{CommanderID: commander.CommanderID, ItemID: 15006, Count: 1}
	if err := GormDB.Create(&ring).Error; err != nil {
		t.Fatalf("seed ring: %v", err)
	}
	commander.CommanderItemsMap[15006] = &ring

	if ok, err := commander.ProposeShip(owned.ID); err != nil || !ok {
		t.Fatalf("propose ship: %v", err)
	}
	if !commander.OwnedShipsMap[owned.ID].Propose {
		t.Fatalf("expected ship proposed")
	}
	if _, err := commander.ProposeShip(999); err == nil {
		t.Fatalf("expected error for missing ship")
	}

	if err := commander.UpdateRoom(42); err != nil {
		t.Fatalf("update room: %v", err)
	}

	pos := uint32(0)
	commander.Ships[0].IsSecretary = true
	commander.Ships[0].SecretaryPosition = &pos
	if err := commander.RemoveSecretaries(); err != nil {
		t.Fatalf("remove secretaries: %v", err)
	}
	if commander.Ships[0].IsSecretary {
		t.Fatalf("expected secretary removed")
	}
	if err := commander.UpdateSecretaries([]SecretaryUpdate{{ShipID: owned.ID, PhantomID: 5}}); err != nil {
		t.Fatalf("update secretaries: %v", err)
	}
	if !commander.Ships[0].IsSecretary || commander.Ships[0].SecretaryPhantomID != 5 {
		t.Fatalf("expected secretary updated")
	}
}

func TestCommanderCountersAndLike(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Like{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 70, AccountID: 70, Name: "Counts"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.ExchangeCount = 390
	if err := commander.IncrementExchangeCount(20); err != nil {
		t.Fatalf("increment exchange: %v", err)
	}
	if commander.ExchangeCount != 400 {
		t.Fatalf("expected exchange count capped at 400")
	}
	if err := commander.IncrementDrawCount(1); err != nil {
		t.Fatalf("increment draw 1: %v", err)
	}
	if err := commander.IncrementDrawCount(10); err != nil {
		t.Fatalf("increment draw 10: %v", err)
	}
	if err := commander.IncrementDrawCount(5); err != nil {
		t.Fatalf("increment draw other should be nil, got %v", err)
	}

	now := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	month := SupportRequisitionMonth(now)
	if month == 0 {
		t.Fatalf("expected month computed")
	}
	if !commander.EnsureSupportRequisitionMonth(now) {
		t.Fatalf("expected month update")
	}
	if commander.EnsureSupportRequisitionMonth(now) {
		t.Fatalf("expected no update for same month")
	}

	if err := commander.Like(123); err != nil {
		t.Fatalf("like: %v", err)
	}
}

func TestCommanderRetireShipErrors(t *testing.T) {
	commander := Commander{OwnedShipsMap: map[uint32]*OwnedShip{}}
	shipIDs := []uint32{999}
	if err := commander.RetireShips(&shipIDs); err == nil {
		t.Fatalf("expected error for missing ship")
	}
}

func TestCommanderGiveSkinWithExpiryCreate(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedSkin{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 71, AccountID: 71, Name: "Expiry"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	expiry := time.Now().Add(time.Hour)
	if err := commander.GiveSkinWithExpiry(9001, &expiry); err != nil {
		t.Fatalf("give skin with expiry: %v", err)
	}
}

func TestCommanderSendMailTx(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Mail{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 72, AccountID: 72, Name: "Mailer"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	mail := Mail{Title: "Tx", Body: "Body"}
	tx := GormDB.Begin()
	if err := commander.SendMailTx(tx, &mail); err != nil {
		tx.Rollback()
		t.Fatalf("send mail tx: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit mail tx: %v", err)
	}
}

func TestCommanderSetRandomFlags(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Commander{})
	clearTable(t, &CommanderCommonFlag{})

	commander := Commander{CommanderID: 73, AccountID: 73, Name: "Random"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	if err := UpdateCommanderRandomFlagShipEnabled(GormDB, commander.CommanderID, true); err != nil {
		t.Fatalf("update random flag: %v", err)
	}
	if err := UpdateCommanderRandomShipMode(GormDB, commander.CommanderID, 2); err != nil {
		t.Fatalf("update random ship mode: %v", err)
	}
	if err := UpdateCommanderRandomShipMode(GormDB, commander.CommanderID, 1); err != nil {
		t.Fatalf("update random ship mode clear: %v", err)
	}
	flags, err := ListCommanderCommonFlags(commander.CommanderID)
	if err != nil {
		t.Fatalf("list commander common flags: %v", err)
	}
	if len(flags) != 0 {
		t.Fatalf("expected flags cleared")
	}
	if err := SetCommanderCommonFlag(GormDB, commander.CommanderID, consts.RandomFlagShipMode); err != nil {
		t.Fatalf("set common flag: %v", err)
	}
}

func TestCommanderStoryAndLivingAreaAndAttire(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderStory{})
	clearTable(t, &CommanderLivingAreaCover{})
	clearTable(t, &CommanderAttire{})

	if err := AddCommanderStory(GormDB, 1, 10); err != nil {
		t.Fatalf("add story: %v", err)
	}
	ids, err := ListCommanderStoryIDs(1)
	if err != nil {
		t.Fatalf("list story ids: %v", err)
	}
	if len(ids) != 1 || ids[0] != 10 {
		t.Fatalf("unexpected story ids")
	}

	entry := CommanderLivingAreaCover{CommanderID: 1, CoverID: 5, IsNew: true}
	if err := UpsertCommanderLivingAreaCover(GormDB, entry); err != nil {
		t.Fatalf("upsert cover: %v", err)
	}
	if ok, err := CommanderHasLivingAreaCover(1, 5); err != nil || !ok {
		t.Fatalf("expected cover present")
	}
	covers, err := ListCommanderLivingAreaCovers(1)
	if err != nil || len(covers) != 1 {
		t.Fatalf("list covers: %v", err)
	}

	attire := CommanderAttire{CommanderID: 1, Type: 2, AttireID: 3}
	if err := UpsertCommanderAttire(GormDB, attire); err != nil {
		t.Fatalf("upsert attire: %v", err)
	}
	if ok, err := CommanderHasAttire(1, 2, 3, time.Now()); err != nil || !ok {
		t.Fatalf("expected attire present")
	}
	attires, err := ListCommanderAttires(1)
	if err != nil || len(attires) != 1 {
		t.Fatalf("list attires: %v", err)
	}
}

func TestCommanderBuffs(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderBuff{})

	now := time.Now().UTC()
	if err := UpsertCommanderBuff(1, 5, now.Add(time.Hour)); err != nil {
		t.Fatalf("upsert commander buff: %v", err)
	}
	buffs, err := ListCommanderBuffs(1)
	if err != nil || len(buffs) != 1 {
		t.Fatalf("list buffs: %v", err)
	}
	active, err := ListCommanderActiveBuffs(1, now)
	if err != nil || len(active) != 1 {
		t.Fatalf("list active buffs: %v", err)
	}
}

func TestCommanderCommonFlagClear(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderCommonFlag{})

	if err := SetCommanderCommonFlag(GormDB, 2, 9); err != nil {
		t.Fatalf("set common flag: %v", err)
	}
	flags, err := ListCommanderCommonFlags(2)
	if err != nil || len(flags) != 1 {
		t.Fatalf("list common flags: %v", err)
	}
	if err := ClearCommanderCommonFlag(GormDB, 2, 9); err != nil {
		t.Fatalf("clear common flag: %v", err)
	}
}

func TestCommanderAttireExpiry(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderAttire{})

	expired := time.Now().Add(-time.Hour)
	entry := CommanderAttire{CommanderID: 3, Type: 1, AttireID: 1, ExpiresAt: &expired}
	if err := UpsertCommanderAttire(GormDB, entry); err != nil {
		t.Fatalf("upsert attire: %v", err)
	}
	if ok, err := CommanderHasAttire(3, 1, 1, time.Now()); err != nil || ok {
		t.Fatalf("expected expired attire false")
	}
}

func TestCommanderConsumeItemTxErrors(t *testing.T) {
	commander := Commander{CommanderItemsMap: map[uint32]*CommanderItem{}}
	if err := commander.ConsumeItemTx(GormDB, 1, 1); err == nil {
		t.Fatalf("expected not enough items error")
	}
}

func TestCommanderProposeMissingRing(t *testing.T) {
	commander := Commander{CommanderID: 74}
	commander.OwnedShipsMap = map[uint32]*OwnedShip{1: {ID: 1}}
	commander.CommanderItemsMap = map[uint32]*CommanderItem{}
	if _, err := commander.ProposeShip(1); err == nil {
		t.Fatalf("expected missing ring error")
	}
}

func TestCommanderGiveSkinWithExpiryNoMap(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedSkin{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 75, AccountID: 75, Name: "Skin"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	if err := commander.GiveSkinWithExpiry(9100, nil); err != nil {
		t.Fatalf("give skin expiry nil: %v", err)
	}
}
