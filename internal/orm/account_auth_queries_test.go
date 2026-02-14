package orm

import (
	"strconv"
	"strings"
	"time"

	"errors"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var accountAuthQueriesOnce sync.Once

func initAccountAuthQueriesDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	accountAuthQueriesOnce.Do(func() {
		InitDatabase()
	})
}

func clearAccountAuthQueryTables(t *testing.T) {
	t.Helper()
	tables := []string{
		"account_permission_overrides",
		"account_roles",
		"web_authn_credentials",
		"auth_challenges",
		"sessions",
		"audit_logs",
		"accounts",
	}
	for _, table := range tables {
		if _, err := db.DefaultStore.Pool.Exec(t.Context(), "DELETE FROM "+table); err != nil {
			t.Fatalf("clear table %s: %v", table, err)
		}
	}
}

func TestAccountAuthQueriesNotFoundOnZeroRows(t *testing.T) {
	initAccountAuthQueriesDB(t)
	clearAccountAuthQueryTables(t)

	now := time.Now().UTC()

	err := UpdateAccountUsername("missing-account", "admin", "admin", now)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected UpdateAccountUsername to return db.ErrNotFound, got %v", err)
	}

	err = UpdateAccountDisabledAt("missing-account", &now, now)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected UpdateAccountDisabledAt to return db.ErrNotFound, got %v", err)
	}

	err = UpdateAccountPassword("missing-account", "hash", "argon2id", now, now)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected UpdateAccountPassword to return db.ErrNotFound, got %v", err)
	}

	err = DeleteAccountByID("missing-account")
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected DeleteAccountByID to return db.ErrNotFound, got %v", err)
	}
}

func TestListAdminAccountsCountsOnlyAdmins(t *testing.T) {
	initAccountAuthQueriesDB(t)
	clearAccountAuthQueryTables(t)

	now := time.Now().UTC().Truncate(time.Second)
	admin1 := "admin-one"
	admin2 := "admin-two"
	member := "player-user"
	for i, acct := range []struct {
		id      string
		isAdmin bool
		name    string
	}{
		{id: "admin-a", isAdmin: true, name: admin1},
		{id: "admin-b", isAdmin: true, name: admin2},
		{id: "user-c", isAdmin: false, name: member},
	} {
		name := acct.name
		row := Account{
			ID:                 acct.id,
			Username:           &name,
			UsernameNormalized: &name,
			PasswordHash:       "hash" + strconv.Itoa(i),
			PasswordAlgo:       "algo",
			PasswordUpdatedAt:  now,
			IsAdmin:            acct.isAdmin,
			CreatedAt:          now.Add(time.Second * time.Duration(i)),
			UpdatedAt:          now.Add(time.Second * time.Duration(i)),
		}
		if err := CreateAccount(&row); err != nil {
			t.Fatalf("seed account %s: %v", acct.id, err)
		}
	}

	accounts, total, err := ListAdminAccounts(0, 20)
	if err != nil {
		t.Fatalf("list admin accounts: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total admins 2, got %d", total)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 admin rows, got %d", len(accounts))
	}
	for _, acct := range accounts {
		if !acct.IsAdmin {
			t.Fatalf("expected admin account in list, got %q", *acct.Username)
		}
	}
}

func TestListAdminAccountsIncludesRoleAdmins(t *testing.T) {
	initAccountAuthQueriesDB(t)
	clearAccountAuthQueryTables(t)

	if err := EnsureAuthzDefaults(); err != nil {
		t.Fatalf("ensure authz defaults: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	seed := []struct {
		id      string
		isAdmin bool
		name    string
		role    string
	}{
		{id: "admin-flag", isAdmin: true, name: "flag-admin"},
		{id: "admin-role", isAdmin: false, name: "role-admin", role: "admin"},
		{id: "member", isAdmin: false, name: "member", role: "player"},
	}
	for i, acct := range seed {
		name := acct.name
		row := Account{
			ID:                 acct.id,
			Username:           &name,
			UsernameNormalized: &name,
			PasswordHash:       "hash" + strconv.Itoa(i),
			PasswordAlgo:       "algo",
			PasswordUpdatedAt:  now,
			IsAdmin:            acct.isAdmin,
			CreatedAt:          now.Add(time.Second * time.Duration(i)),
			UpdatedAt:          now.Add(time.Second * time.Duration(i)),
		}
		if err := CreateAccount(&row); err != nil {
			t.Fatalf("seed account %s: %v", acct.id, err)
		}
		if acct.role != "" {
			if err := AssignRoleByName(acct.id, acct.role); err != nil {
				t.Fatalf("assign role %s to %s: %v", acct.role, acct.id, err)
			}
		}
	}

	accounts, total, err := ListAdminAccounts(0, 20)
	if err != nil {
		t.Fatalf("list admin accounts: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total admins 2, got %d", total)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 admin rows, got %d", len(accounts))
	}

	ids := map[string]bool{}
	for _, acct := range accounts {
		ids[acct.ID] = true
	}
	if !ids["admin-flag"] || !ids["admin-role"] {
		t.Fatalf("expected flag and role admins in list, got ids=%v", ids)
	}
	if ids["member"] {
		t.Fatalf("did not expect non-admin account in list")
	}
}

func TestGetAccountByIDNullableFieldsRoundTrip(t *testing.T) {
	initAccountAuthQueriesDB(t)
	clearAccountAuthQueryTables(t)

	now := time.Now().UTC().Truncate(time.Second)
	account := Account{
		ID:                "nullable-account",
		PasswordHash:      "hash",
		PasswordAlgo:      "algo",
		PasswordUpdatedAt: now,
		IsAdmin:           false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := CreateAccount(&account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	loaded, err := GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if loaded.Username != nil || loaded.UsernameNormalized != nil {
		t.Fatalf("expected nil usernames, got username=%v normalized=%v", loaded.Username, loaded.UsernameNormalized)
	}
	if loaded.CommanderID != nil {
		t.Fatalf("expected nil commander id, got %v", *loaded.CommanderID)
	}
	if loaded.DisabledAt != nil || loaded.LastLoginAt != nil {
		t.Fatalf("expected nil disabled/last login, got disabled=%v last_login=%v", loaded.DisabledAt, loaded.LastLoginAt)
	}
}

func TestWebAuthnCredentialNullableFieldsRoundTrip(t *testing.T) {
	initAccountAuthQueriesDB(t)
	clearAccountAuthQueryTables(t)

	now := time.Now().UTC().Truncate(time.Second)
	username := "credential-user"
	account := Account{
		ID:                 "credential-account",
		Username:           &username,
		UsernameNormalized: &username,
		PasswordHash:       "hash",
		PasswordAlgo:       "algo",
		PasswordUpdatedAt:  now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := CreateAccount(&account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	if _, err := db.DefaultStore.Pool.Exec(t.Context(), `
INSERT INTO web_authn_credentials (
	id,
	user_id,
	credential_id,
	public_key,
	sign_count,
	transports,
	aaguid,
	attestation_fmt,
	resident_key,
	backup_eligible,
	backup_state,
	created_at,
	last_used_at,
	label,
	rp_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL, NULL, $10, NULL, NULL, $11)
`, "cred-1", account.ID, "credential-id-1", []byte("pub"), int64(7), []byte("[]"), "aaguid", "none", "", now, "rp.example"); err != nil {
		t.Fatalf("seed webauthn credential: %v", err)
	}

	loaded, err := ListWebAuthnCredentialsByUserID(account.ID)
	if err != nil {
		t.Fatalf("list webauthn credentials: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(loaded))
	}
	if loaded[0].BackupEligible != nil || loaded[0].BackupState != nil || loaded[0].LastUsedAt != nil || loaded[0].Label != nil {
		t.Fatalf("expected nullable fields to remain nil, got eligible=%v state=%v last_used=%v label=%v", loaded[0].BackupEligible, loaded[0].BackupState, loaded[0].LastUsedAt, loaded[0].Label)
	}
}

func TestAccountAuthQueriesRejectOutOfRangeUint32Mappings(t *testing.T) {
	initAccountAuthQueriesDB(t)
	clearAccountAuthQueryTables(t)

	now := time.Now().UTC().Truncate(time.Second)

	t.Run("account commander id", func(t *testing.T) {
		if _, err := db.DefaultStore.Pool.Exec(t.Context(), `
INSERT INTO accounts (
	id,
	username,
	username_normalized,
	commander_id,
	password_hash,
	password_algo,
	password_updated_at,
	is_admin,
	disabled_at,
	last_login_at,
	web_authn_user_handle,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL, NULL, $9, $10, $11)
`, "bad-commander", "bad", "bad", int64(-1), "hash", "algo", now, false, []byte("handle-bad-commander"), now, now); err != nil {
			t.Fatalf("seed account with invalid commander id: %v", err)
		}

		_, err := GetAccountByID("bad-commander")
		if err == nil {
			t.Fatalf("expected conversion error for invalid commander id")
		}
		if !strings.Contains(err.Error(), "out of uint32 range") {
			t.Fatalf("expected range error, got %v", err)
		}
	})

	t.Run("webauthn sign count", func(t *testing.T) {
		username := "signcount-user"
		account := Account{
			ID:                 "signcount-account",
			Username:           &username,
			UsernameNormalized: &username,
			PasswordHash:       "hash",
			PasswordAlgo:       "algo",
			PasswordUpdatedAt:  now,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := CreateAccount(&account); err != nil {
			t.Fatalf("create account: %v", err)
		}

		if _, err := db.DefaultStore.Pool.Exec(t.Context(), `
INSERT INTO web_authn_credentials (
	id,
	user_id,
	credential_id,
	public_key,
	sign_count,
	transports,
	aaguid,
	attestation_fmt,
	resident_key,
	backup_eligible,
	backup_state,
	created_at,
	last_used_at,
	label,
	rp_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL, NULL, $10, NULL, NULL, $11)
`, "bad-signcount", account.ID, "credential-bad-signcount", []byte("pub"), int64(-1), []byte("[]"), "", "", "", now, "rp.example"); err != nil {
			t.Fatalf("seed webauthn credential with invalid sign_count: %v", err)
		}

		_, err := ListWebAuthnCredentialsByUserID(account.ID)
		if err == nil {
			t.Fatalf("expected conversion error for invalid sign_count")
		}
		if !strings.Contains(err.Error(), "out of uint32 range") {
			t.Fatalf("expected range error, got %v", err)
		}
	})
}

func TestUint32FromInt64Checked(t *testing.T) {
	if got, err := Uint32FromInt64Checked(0); err != nil || got != 0 {
		t.Fatalf("expected 0 without error, got value=%d err=%v", got, err)
	}
	if got, err := Uint32FromInt64Checked(4294967295); err != nil || got != 4294967295 {
		t.Fatalf("expected max uint32 without error, got value=%d err=%v", got, err)
	}
	if _, err := Uint32FromInt64Checked(-1); err == nil {
		t.Fatalf("expected error for negative value")
	}
	if _, err := Uint32FromInt64Checked(4294967296); err == nil {
		t.Fatalf("expected error for value above uint32 max")
	}
}
