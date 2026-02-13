package orm

import (
	"context"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var localAccountTestOnce sync.Once

func initLocalAccountTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	localAccountTestOnce.Do(func() {
		InitDatabase()
	})
	clearTable(t, &LocalAccount{})
}

func TestLocalAccountCreate(t *testing.T) {
	initLocalAccountTest(t)
	entry := LocalAccount{Arg2: 123456, Account: "user", Password: "pass", MailBox: "mail"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO local_accounts (arg2, account, password, mail_box) VALUES ($1, $2, $3, $4)`, int64(entry.Arg2), entry.Account, entry.Password, entry.MailBox); err != nil {
		t.Fatalf("create local account failed: %v", err)
	}
	var stored LocalAccount
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT arg2, account, password, mail_box FROM local_accounts WHERE account = $1`, "user").Scan(&stored.Arg2, &stored.Account, &stored.Password, &stored.MailBox); err != nil {
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
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO local_accounts (arg2, account, password, mail_box) VALUES ($1, $2, $3, $4)`, int64(entry1.Arg2), entry1.Account, entry1.Password, entry1.MailBox); err != nil {
		t.Fatalf("create first local account failed: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO local_accounts (arg2, account, password, mail_box) VALUES ($1, $2, $3, $4)`, int64(entry2.Arg2), entry2.Account, entry2.Password, entry2.MailBox); err == nil {
		t.Fatalf("expected duplicate account to fail")
	}
}

func TestLocalAccountDuplicateArg2(t *testing.T) {
	initLocalAccountTest(t)
	entry1 := LocalAccount{Arg2: 123459, Account: "user1", Password: "pass", MailBox: "mail"}
	entry2 := LocalAccount{Arg2: 123459, Account: "user2", Password: "pass2", MailBox: "mail2"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO local_accounts (arg2, account, password, mail_box) VALUES ($1, $2, $3, $4)`, int64(entry1.Arg2), entry1.Account, entry1.Password, entry1.MailBox); err != nil {
		t.Fatalf("create first local account failed: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO local_accounts (arg2, account, password, mail_box) VALUES ($1, $2, $3, $4)`, int64(entry2.Arg2), entry2.Account, entry2.Password, entry2.MailBox); err == nil {
		t.Fatalf("expected duplicate arg2 to fail")
	}
}
