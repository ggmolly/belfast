package orm

import (
	"sync"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/consts"
	"gorm.io/gorm"
)

var mailTestOnce sync.Once

func initMailTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	mailTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Mail{}).Error; err != nil {
		t.Fatalf("clear mail: %v", err)
	}
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&MailAttachment{}).Error; err != nil {
		t.Fatalf("clear mail attachments: %v", err)
	}
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Commander{}).Error; err != nil {
		t.Fatalf("clear commanders: %v", err)
	}
}

func TestMailCreate(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1001,
		AccountID:   100,
		Name:        "Test Commander",
		Level:       1,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:          1,
		ReceiverID:  commander.CommanderID,
		Read:        false,
		Title:       "Welcome to Belfast!",
		Body:        "Thank you for joining our server.",
		IsImportant: false,
	}

	if err := mail.Create(); err != nil {
		t.Fatalf("create mail: %v", err)
	}

	if mail.ID != 1 {
		t.Fatalf("expected id 1, got %d", mail.ID)
	}
	if mail.Title != "Welcome to Belfast!" {
		t.Fatalf("expected title 'Welcome to Belfast!', got %s", mail.Title)
	}
}

func TestMailFind(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1002,
		AccountID:   101,
		Name:        "Test Commander 2",
		Level:       1,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:         2,
		ReceiverID: commander.CommanderID,
		Read:       false,
		Title:      "Test Mail",
		Body:       "Test body",
	}
	mail.Create()

	var found Mail
	if err := GormDB.First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if found.ID != 2 {
		t.Fatalf("expected id 2, got %d", found.ID)
	}
	if found.Title != "Test Mail" {
		t.Fatalf("expected title 'Test Mail', got %s", found.Title)
	}
}

func TestMailDelete(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1003,
		AccountID:   102,
		Name:        "Test Commander 3",
		Level:       1,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:         3,
		ReceiverID: commander.CommanderID,
		Read:       false,
		Title:      "Delete Test",
		Body:       "To be deleted",
	}
	mail.Create()

	if err := mail.Delete(); err != nil {
		t.Fatalf("delete mail: %v", err)
	}

	var found Mail
	err := GormDB.First(&found, mail.ID).Error
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestMailCollectAttachments(t *testing.T) {
	initMailTest(t)

	commander := Commander{CommanderID: 1004, AccountID: 103, Name: "Collect", Level: 1}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	commander.CommanderItemsMap = make(map[uint32]*CommanderItem)
	commander.OwnedShipsMap = make(map[uint32]*OwnedShip)
	commander.OwnedSkinsMap = make(map[uint32]*OwnedSkin)

	ship := Ship{TemplateID: 11001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	mail := Mail{ID: 4, ReceiverID: commander.CommanderID, Title: "Collect", Body: "Body"}
	mail.Attachments = []MailAttachment{
		{Type: consts.DROP_TYPE_RESOURCE, ItemID: 1, Quantity: 5},
		{Type: consts.DROP_TYPE_ITEM, ItemID: 200, Quantity: 2},
		{Type: consts.DROP_TYPE_SHIP, ItemID: ship.TemplateID, Quantity: 1},
		{Type: consts.DROP_TYPE_SKIN, ItemID: 300, Quantity: 1},
		{Type: 999, ItemID: 0, Quantity: 0},
	}
	if err := mail.Create(); err != nil {
		t.Fatalf("create mail: %v", err)
	}
	attachments, err := mail.CollectAttachments(&commander)
	if err != nil {
		t.Fatalf("collect attachments: %v", err)
	}
	if len(attachments) != 5 {
		t.Fatalf("expected 5 attachments, got %d", len(attachments))
	}
	if !mail.AttachmentsCollected {
		t.Fatalf("expected attachments collected")
	}
}

func TestMailUpdate(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1004,
		AccountID:   103,
		Name:        "Test Commander 4",
		Level:       1,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:         4,
		ReceiverID: commander.CommanderID,
		Read:       false,
		Title:      "Original Title",
		Body:       "Original body",
	}
	mail.Create()

	mail.Title = "Updated Title"
	mail.Body = "Updated body"

	if err := mail.Update(); err != nil {
		t.Fatalf("update mail: %v", err)
	}

	var found Mail
	if err := GormDB.First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find updated mail: %v", err)
	}

	if found.Title != "Updated Title" {
		t.Fatalf("expected title 'Updated Title', got %s", found.Title)
	}
	if found.Body != "Updated body" {
		t.Fatalf("expected body 'Updated body', got %s", found.Body)
	}
}

func TestMailSetRead(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1005,
		AccountID:   104,
		Name:        "Test Commander 5",
		Level:       1,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:         5,
		ReceiverID: commander.CommanderID,
		Read:       false,
		Title:      "Unread Mail",
		Body:       "You have unread mail",
	}
	mail.Create()

	if err := mail.SetRead(true); err != nil {
		t.Fatalf("set mail as read: %v", err)
	}

	var found Mail
	if err := GormDB.First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if !found.Read {
		t.Fatalf("expected mail to be read")
	}

	if err := mail.SetRead(false); err != nil {
		t.Fatalf("set mail as unread: %v", err)
	}

	var found2 Mail
	if err := GormDB.First(&found2, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if found2.Read {
		t.Fatalf("expected mail to be unread")
	}
}

func TestMailSetImportant(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1006,
		AccountID:   105,
		Name:        "Test Commander 6",
		Level:       1,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:          6,
		ReceiverID:  commander.CommanderID,
		Read:        false,
		Title:       "Normal Mail",
		Body:        "This is normal",
		IsImportant: false,
	}
	mail.Create()

	if err := mail.SetImportant(true); err != nil {
		t.Fatalf("set mail as important: %v", err)
	}

	var found Mail
	if err := GormDB.First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if !found.IsImportant {
		t.Fatalf("expected mail to be important")
	}
}

func TestMailSetArchived(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1007,
		AccountID:   106,
		Name:        "Test Commander 7",
		Level:       1,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:         7,
		ReceiverID: commander.CommanderID,
		Read:       false,
		Title:      "Archivable Mail",
		Body:       "This can be archived",
		IsArchived: false,
	}
	mail.Create()

	if err := mail.SetArchived(true); err != nil {
		t.Fatalf("set mail as archived: %v", err)
	}

	var found Mail
	if err := GormDB.First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if !found.IsArchived {
		t.Fatalf("expected mail to be archived")
	}
}

func TestMailCustomSender(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1008,
		AccountID:   107,
		Name:        "Test Commander 8",
		Level:       1,
	}
	GormDB.Create(&commander)

	sender := "Admin"
	mail := Mail{
		ID:           8,
		ReceiverID:   commander.CommanderID,
		Read:         false,
		Title:        "Admin Message",
		Body:         "This is from admin",
		CustomSender: &sender,
		IsImportant:  true,
	}
	mail.Create()

	var found Mail
	if err := GormDB.First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if found.CustomSender == nil || *found.CustomSender != "Admin" {
		t.Fatalf("expected custom sender 'Admin', got %v", found.CustomSender)
	}
}

func TestMailDate(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1009,
		AccountID:   108,
		Name:        "Test Commander 9",
		Level:       1,
	}
	GormDB.Create(&commander)

	now := time.Now()
	mail := Mail{
		ID:         9,
		ReceiverID: commander.CommanderID,
		Read:       false,
		Title:      "Timed Mail",
		Body:       "With timestamp",
		Date:       now,
	}
	mail.Create()

	var found Mail
	if err := GormDB.First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if found.Date.Before(now) {
		t.Fatalf("expected date to be after or equal to creation time")
	}
}

func TestMailAttachments(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1010,
		AccountID:   109,
		Name:        "Test Commander 10",
		Level:       1,
		Exp:         0,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:         10,
		ReceiverID: commander.CommanderID,
		Read:       false,
		Title:      "Mail with Attachments",
		Body:       "You got items!",
	}
	mail.Create()

	attachment := MailAttachment{
		ID:       1,
		MailID:   mail.ID,
		Type:     1,
		ItemID:   100,
		Quantity: 50,
	}
	GormDB.Create(&attachment)

	var found Mail
	if err := GormDB.Preload("Attachments").First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find mail with attachments: %v", err)
	}

	if len(found.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(found.Attachments))
	}
	if found.Attachments[0].ItemID != 100 {
		t.Fatalf("expected item id 100, got %d", found.Attachments[0].ItemID)
	}
}

func TestMailAttachmentsCollected(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1011,
		AccountID:   110,
		Name:        "Test Commander 11",
		Level:       1,
		Exp:         0,
	}
	GormDB.Create(&commander)

	mail := Mail{
		ID:                   11,
		ReceiverID:           commander.CommanderID,
		Read:                 false,
		Title:                "Mail to Collect",
		Body:                 "Collect these items",
		AttachmentsCollected: false,
	}
	mail.Create()

	var found Mail
	if err := GormDB.First(&found, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if found.AttachmentsCollected {
		t.Fatalf("expected attachments not collected initially")
	}

	mail.AttachmentsCollected = true
	mail.Update()

	var found2 Mail
	if err := GormDB.First(&found2, mail.ID).Error; err != nil {
		t.Fatalf("find mail: %v", err)
	}

	if !found2.AttachmentsCollected {
		t.Fatalf("expected attachments collected after update")
	}
}

func TestMailMultipleMails(t *testing.T) {
	initMailTest(t)

	commander := Commander{
		CommanderID: 1012,
		AccountID:   111,
		Name:        "Test Commander 12",
		Level:       1,
	}
	GormDB.Create(&commander)

	for i := 1; i <= 5; i++ {
		mail := Mail{
			ID:         uint32(20 + i),
			ReceiverID: commander.CommanderID,
			Read:       false,
			Title:      "Mail",
			Body:       "Body",
		}
		mail.Create()
	}

	var mails []Mail
	if err := GormDB.Where("receiver_id = ?", commander.CommanderID).Find(&mails).Error; err != nil {
		t.Fatalf("find mails: %v", err)
	}

	if len(mails) != 5 {
		t.Fatalf("expected 5 mails, got %d", len(mails))
	}
}
