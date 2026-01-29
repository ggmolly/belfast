package orm

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/consts"
)

func TestCompensationCRUDAndExpiry(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Compensation{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 80, AccountID: 80, Name: "Comp"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	comp := Compensation{CommanderID: commander.CommanderID, Title: "T", Text: "Body", ExpiresAt: time.Now().Add(time.Hour)}
	if err := comp.Create(); err != nil {
		t.Fatalf("create compensation: %v", err)
	}
	comp.Text = "Updated"
	if err := comp.Update(); err != nil {
		t.Fatalf("update compensation: %v", err)
	}
	if comp.IsExpired(time.Now()) {
		t.Fatalf("expected not expired")
	}
	if !(&Compensation{ExpiresAt: time.Time{}}).IsExpired(time.Now()) {
		t.Fatalf("expected zero expiry to be expired")
	}
	if err := comp.Delete(); err != nil {
		t.Fatalf("delete compensation: %v", err)
	}
}

func TestCompensationCollectAttachmentsAndSummary(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Compensation{})
	clearTable(t, &CompensationAttachment{})
	clearTable(t, &Commander{})
	clearTable(t, &Ship{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &CommanderItem{})
	clearTable(t, &OwnedResource{})
	clearTable(t, &OwnedSkin{})

	commander := Commander{CommanderID: 81, AccountID: 81, Name: "Comp"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	commander.CommanderItemsMap = make(map[uint32]*CommanderItem)
	commander.OwnedShipsMap = make(map[uint32]*OwnedShip)
	commander.OwnedSkinsMap = make(map[uint32]*OwnedSkin)

	ship := Ship{TemplateID: 10001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	comp := Compensation{CommanderID: commander.CommanderID, Title: "T", Text: "Body", ExpiresAt: time.Now().Add(time.Hour)}
	comp.Attachments = []CompensationAttachment{
		{Type: consts.DROP_TYPE_RESOURCE, ItemID: 1, Quantity: 5},
		{Type: consts.DROP_TYPE_ITEM, ItemID: 200, Quantity: 2},
		{Type: consts.DROP_TYPE_SHIP, ItemID: ship.TemplateID, Quantity: 1},
		{Type: consts.DROP_TYPE_SKIN, ItemID: 300, Quantity: 1},
		{Type: 999, ItemID: 0, Quantity: 0},
	}
	if err := comp.Create(); err != nil {
		t.Fatalf("create compensation: %v", err)
	}
	attachments, err := comp.CollectAttachments(&commander)
	if err != nil {
		t.Fatalf("collect attachments: %v", err)
	}
	if len(attachments) != 5 {
		t.Fatalf("expected 5 attachments, got %d", len(attachments))
	}
	if !comp.AttachFlag {
		t.Fatalf("expected attach flag set")
	}

	now := time.Now()
	count, maxTimestamp := CompensationSummary([]Compensation{comp}, now)
	if count != 0 {
		t.Fatalf("expected no uncollected compensations")
	}
	if maxTimestamp == 0 {
		t.Fatalf("expected max timestamp")
	}
}

func TestLoadCommanderCompensations(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Compensation{})
	clearTable(t, &CompensationAttachment{})

	comp := Compensation{CommanderID: 82, Title: "T", Text: "Body", ExpiresAt: time.Now().Add(time.Hour)}
	comp.Attachments = []CompensationAttachment{{Type: 1, ItemID: 1, Quantity: 1}}
	if err := comp.Create(); err != nil {
		t.Fatalf("create compensation: %v", err)
	}
	loaded, err := LoadCommanderCompensations(82)
	if err != nil {
		t.Fatalf("load compensations: %v", err)
	}
	if len(loaded) != 1 || len(loaded[0].Attachments) != 1 {
		t.Fatalf("expected attachments loaded")
	}
}
