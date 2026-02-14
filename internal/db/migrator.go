package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migration struct {
	Version  int64
	Name     string
	Filename string
	SQL      string
	Checksum [32]byte
	// NoTransaction disables wrapping the migration in a transaction.
	// Use sparingly; Postgres supports transactional DDL for most statements.
	NoTransaction bool
}

type MigratorOptions struct {
	SchemaName string
}

var migrationFilenamePattern = regexp.MustCompile(`^(\d+)_([a-zA-Z0-9][a-zA-Z0-9_-]*)\.sql$`)

const migrationAdvisoryLockClassID int32 = 0x62666d67
const migrationAdvisoryLockAcquireTimeout = 10 * time.Second
const migrationStatementTimeout = 5 * time.Minute
const migrationAdvisoryLockTimeout = 5 * time.Second
const migrationResetTimeout = 5 * time.Second

func LoadEmbeddedMigrations() ([]Migration, error) {
	entries, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, ErrMigrationsUnavailable
	}
	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		base := filepath.Base(entry)
		match := migrationFilenamePattern.FindStringSubmatch(base)
		if match == nil {
			return nil, fmt.Errorf("invalid migration filename %q (expected NNNN_name.sql)", base)
		}
		version, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version %q: %w", match[1], err)
		}
		data, err := fs.ReadFile(migrationsFS, entry)
		if err != nil {
			return nil, err
		}
		sqlText := string(data)
		checksum := sha256.Sum256(data)
		migrations = append(migrations, Migration{
			Version:       version,
			Name:          match[2],
			Filename:      base,
			SQL:           sqlText,
			Checksum:      checksum,
			NoTransaction: hasNoTransactionDirective(sqlText),
		})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version == migrations[i-1].Version {
			return nil, fmt.Errorf("duplicate migration version %d", migrations[i].Version)
		}
	}
	return migrations, nil
}

func hasNoTransactionDirective(sqlText string) bool {
	// Minimal directive: a line containing `+migrate NoTransaction`.
	// Matches common formats like:
	//   -- +migrate NoTransaction
	//   --+migrate NoTransaction
	for _, line := range strings.Split(sqlText, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") {
			continue
		}
		if strings.Contains(trimmed, "+migrate") && strings.Contains(strings.ToLower(trimmed), "notransaction") {
			return true
		}
	}
	return false
}

func RunMigrations(ctx context.Context, db *sql.DB, opts MigratorOptions) error {
	lockConn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer lockConn.Close()

	acquireCtx, cancel := context.WithTimeout(ctx, migrationAdvisoryLockAcquireTimeout)
	defer cancel()

	if err := setMigrationSessionTimeouts(acquireCtx, lockConn, migrationAdvisoryLockTimeout, migrationStatementTimeout, false); err != nil {
		return err
	}

	lockObjectID := migrationAdvisoryLockObjectID(opts.SchemaName)
	if _, err := lockConn.ExecContext(acquireCtx, `SELECT pg_advisory_lock($1, $2)`, migrationAdvisoryLockClassID, lockObjectID); err != nil {
		return err
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), migrationResetTimeout)
		defer cancel()
		_, _ = lockConn.ExecContext(unlockCtx, `SELECT pg_advisory_unlock($1, $2)`, migrationAdvisoryLockClassID, lockObjectID)
	}()

	migrations, err := LoadEmbeddedMigrations()
	if err != nil {
		return err
	}
	if err := ensureSchemaMigrationsTable(ctx, db, opts.SchemaName); err != nil {
		return err
	}
	applied, err := loadAppliedMigrations(ctx, db, opts.SchemaName)
	if err != nil {
		return err
	}
	for _, m := range migrations {
		if appliedChecksum, ok := applied[m.Version]; ok {
			if appliedChecksum != m.Checksum {
				return fmt.Errorf("migration %d already applied but checksum changed (%s)", m.Version, m.Filename)
			}
			continue
		}
		if err := applyMigration(ctx, db, opts.SchemaName, m); err != nil {
			return err
		}
	}
	return nil
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB, schemaName string) error {
	table := qualifiedName(schemaName, "schema_migrations")
	// applied_at is stored for operator debugging; migrations are ordered by version.
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
  version BIGINT PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  checksum BYTEA NOT NULL
)
`, table))
	return err
}

func loadAppliedMigrations(ctx context.Context, db *sql.DB, schemaName string) (map[int64][32]byte, error) {
	table := qualifiedName(schemaName, "schema_migrations")
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`SELECT version, checksum FROM %s`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64][32]byte)
	for rows.Next() {
		var version int64
		var checksum []byte
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, err
		}
		if len(checksum) != 32 {
			return nil, fmt.Errorf("invalid checksum length %d for migration %d", len(checksum), version)
		}
		var sum [32]byte
		copy(sum[:], checksum)
		applied[version] = sum
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return applied, nil
}

func applyMigration(ctx context.Context, db *sql.DB, schemaName string, m Migration) error {
	if strings.TrimSpace(m.SQL) == "" {
		// Empty migration is still recorded to lock-in version ordering.
		return recordMigration(ctx, db, schemaName, m)
	}
	if m.NoTransaction {
		conn, err := db.Conn(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()

		if err := setMigrationSessionTimeouts(ctx, conn, 0, migrationStatementTimeout, false); err != nil {
			return err
		}
		defer func() {
			resetCtx, cancel := context.WithTimeout(context.Background(), migrationResetTimeout)
			defer cancel()
			_, _ = conn.ExecContext(resetCtx, `RESET statement_timeout`)
		}()

		if err := setSearchPath(ctx, conn, schemaName); err != nil {
			return err
		}
		defer func() {
			if strings.TrimSpace(schemaName) == "" {
				return
			}
			resetCtx, cancel := context.WithTimeout(context.Background(), migrationResetTimeout)
			defer cancel()
			_, _ = conn.ExecContext(resetCtx, `RESET search_path`)
		}()

		statements := splitSQLStatements(m.SQL)
		for _, statement := range statements {
			if _, err := conn.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply migration %d (%s): %w", m.Version, m.Filename, err)
			}
		}
		return recordMigrationOnConn(ctx, conn, schemaName, m)
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := setLocalSearchPath(ctx, tx, schemaName); err != nil {
		return err
	}
	if err := setMigrationStatementTimeout(ctx, tx, migrationStatementTimeout); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return fmt.Errorf("apply migration %d (%s): %w", m.Version, m.Filename, err)
	}
	if err := recordMigrationTx(ctx, tx, schemaName, m); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func migrationAdvisoryLockObjectID(schemaName string) int32 {
	value := strings.TrimSpace(schemaName)
	if value == "" {
		value = "public"
	}
	sum := sha256.Sum256([]byte(value))
	return int32(binary.BigEndian.Uint32(sum[:4]))
}

func recordMigration(ctx context.Context, db *sql.DB, schemaName string, m Migration) error {
	// Keep the record insert outside the DDL transaction if the migration opted out.
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := setLocalSearchPath(ctx, tx, schemaName); err != nil {
		return err
	}
	if err := recordMigrationTx(ctx, tx, schemaName, m); err != nil {
		return err
	}
	return tx.Commit()
}

func recordMigrationOnConn(ctx context.Context, conn *sql.Conn, schemaName string, m Migration) error {
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := setLocalSearchPath(ctx, tx, schemaName); err != nil {
		return err
	}
	if err := recordMigrationTx(ctx, tx, schemaName, m); err != nil {
		return err
	}
	return tx.Commit()
}

func recordMigrationTx(ctx context.Context, tx *sql.Tx, schemaName string, m Migration) error {
	table := qualifiedName(schemaName, "schema_migrations")
	_, err := tx.ExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (version, name, applied_at, checksum) VALUES ($1, $2, $3, $4) ON CONFLICT (version) DO NOTHING`,
		table,
	), m.Version, m.Name, time.Now().UTC(), m.Checksum[:])
	return err
}

func setSearchPath(ctx context.Context, execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, schemaName string) error {
	if strings.TrimSpace(schemaName) == "" {
		return nil
	}
	// SET affects the session/connection; used only for non-transactional migrations.
	_, err := execer.ExecContext(ctx, `SET search_path TO `+quoteIdent(schemaName))
	return err
}

func setMigrationSessionTimeouts(ctx context.Context, execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, lockTimeout time.Duration, statementTimeout time.Duration, local bool) error {
	localKeyword := ""
	if local {
		localKeyword = "LOCAL "
	}
	if lockTimeout > 0 {
		if _, err := execer.ExecContext(ctx, `SET `+localKeyword+`lock_timeout = '`+formatPgDuration(lockTimeout)+`'`); err != nil {
			return err
		}
	}
	if statementTimeout > 0 {
		if _, err := execer.ExecContext(ctx, `SET `+localKeyword+`statement_timeout = '`+formatPgDuration(statementTimeout)+`'`); err != nil {
			return err
		}
	}
	return nil
}

func setMigrationStatementTimeout(ctx context.Context, tx *sql.Tx, timeout time.Duration) error {
	return setMigrationSessionTimeouts(ctx, tx, 0, timeout, true)
}

func formatPgDuration(d time.Duration) string {
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func setLocalSearchPath(ctx context.Context, tx *sql.Tx, schemaName string) error {
	if strings.TrimSpace(schemaName) == "" {
		return nil
	}
	// SET LOCAL keeps the change scoped to the current transaction.
	_, err := tx.ExecContext(ctx, `SET LOCAL search_path TO `+quoteIdent(schemaName))
	return err
}

func qualifiedName(schemaName string, name string) string {
	schema := strings.TrimSpace(schemaName)
	if schema == "" {
		return quoteIdent(name)
	}
	return quoteIdent(schema) + "." + quoteIdent(name)
}

func quoteIdent(value string) string {
	if value == "" {
		return `""`
	}
	// Basic Postgres identifier quoting.
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func splitSQLStatements(sqlText string) []string {
	var statements []string
	var builder strings.Builder

	inSingleQuote := false
	inDoubleQuote := false
	inLineComment := false
	inBlockComment := false
	dollarTag := ""

	flushStatement := func() {
		statement := strings.TrimSpace(builder.String())
		if statement != "" {
			statements = append(statements, statement)
		}
		builder.Reset()
	}

	for i := 0; i < len(sqlText); i++ {
		ch := sqlText[i]
		next := byte(0)
		if i+1 < len(sqlText) {
			next = sqlText[i+1]
		}

		if inLineComment {
			builder.WriteByte(ch)
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}

		if inBlockComment {
			builder.WriteByte(ch)
			if ch == '*' && next == '/' {
				builder.WriteByte(next)
				i++
				inBlockComment = false
			}
			continue
		}

		if dollarTag != "" {
			if strings.HasPrefix(sqlText[i:], dollarTag) {
				builder.WriteString(dollarTag)
				i += len(dollarTag) - 1
				dollarTag = ""
				continue
			}
			builder.WriteByte(ch)
			continue
		}

		if !inSingleQuote && !inDoubleQuote {
			if ch == '-' && next == '-' {
				builder.WriteByte(ch)
				builder.WriteByte(next)
				i++
				inLineComment = true
				continue
			}

			if ch == '/' && next == '*' {
				builder.WriteByte(ch)
				builder.WriteByte(next)
				i++
				inBlockComment = true
				continue
			}

			if ch == '$' {
				j := i + 1
				for j < len(sqlText) {
					if sqlText[j] == '$' {
						dollarTag = sqlText[i : j+1]
						builder.WriteString(dollarTag)
						i = j
						break
					}
					if (sqlText[j] >= 'a' && sqlText[j] <= 'z') || (sqlText[j] >= 'A' && sqlText[j] <= 'Z') || (sqlText[j] >= '0' && sqlText[j] <= '9') || sqlText[j] == '_' {
						j++
						continue
					}
					break
				}
				if dollarTag != "" {
					continue
				}
			}

			if ch == ';' {
				flushStatement()
				continue
			}
		}

		if ch == '\'' && !inDoubleQuote {
			if inSingleQuote && next == '\'' {
				builder.WriteByte(ch)
				builder.WriteByte(next)
				i++
				continue
			}
			inSingleQuote = !inSingleQuote
		}

		if ch == '"' && !inSingleQuote {
			if inDoubleQuote && next == '"' {
				builder.WriteByte(ch)
				builder.WriteByte(next)
				i++
				continue
			}
			inDoubleQuote = !inDoubleQuote
		}

		builder.WriteByte(ch)
	}

	flushStatement()
	return statements
}

var ErrMigrationsUnavailable = errors.New("no embedded migrations found")
