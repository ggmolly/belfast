package orm

import (
	"sync"
	"testing"

	"gorm.io/gorm"
)

var localAccountTestOnce sync.Once

func initLocalAccountTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	localAccountTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&LocalAccount{}).Error; err != nil {
		t.Fatalf("clear local accounts: %v", err)
	}
}

func TestLocalAccountCreate(t *testing.T) {
	initLocalAccountTest(t)
	entry := LocalAccount{Arg2: 123456, Account: "user", Password: "pass", MailBox: "mail"}
	if err := GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("create local account failed: %v", err)
	}
	var stored LocalAccount
	if err := GormDB.Where("account = ?", "user").First(&stored).Error; err != nil {
		t.Fatalf("fetch local account failed: %v", err)
	}
	if stored.Arg2 != 123456 {
		t.Fatalf("expected arg2 123456, got %d", stored.Arg2)
	}
}

func TestLocalAccountDuplicateAccount(t *testing.T) {
	initLocalAccountTest(t)
	entry1 := LocalAccount{Arg2: 123457, Account: "user", Password: "pass", MailBox: "mail"}
	entry2 := LocalAccount{Arg2: 123458, Account: "user", Password: "pass2", MailBox: "mail2"}
	GormDB.Create(&entry1)
	if err := GormDB.Create(&entry2).Error; err == nil {
		t.Fatalf("expected duplicate account to fail")
	}
}

func TestLocalAccountDuplicateArg2(t *testing.T) {
	initLocalAccountTest(t)
	entry1 := LocalAccount{Arg2: 123459, Account: "user1", Password: "pass", MailBox: "mail"}
	entry2 := LocalAccount{Arg2: 123459, Account: "user2", Password: "pass2", MailBox: "mail2"}
	GormDB.Create(&entry1)
	if err := GormDB.Create(&entry2).Error; err == nil {
		t.Fatalf("expected duplicate arg2 to fail")
	}
}
