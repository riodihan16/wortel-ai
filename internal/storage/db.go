package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Current schema version. Bump when adding migrations.
const currentVersion = 3

// DB wraps the SQLite handle.
type DB struct {
	inner *sql.DB
	Path  string
}

// DefaultPath returns %LOCALAPPDATA%\WortelAI\wortel.db.
// Falls back to user cache dir when LOCALAPPDATA is unset.
func DefaultPath() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		var err error
		base, err = os.UserCacheDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(base, "WortelAI", "wortel.db"), nil
}

// Open creates parent dirs, opens SQLite, enables FK/WAL, runs migrations.
func Open(path string) (*DB, error) {
	if path == "" {
		return nil, fmt.Errorf("empty db path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db := &DB{inner: sqlDB, Path: path}
	for _, p := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.inner.Exec(p); err != nil {
			_ = db.inner.Close()
			return nil, fmt.Errorf("pragma: %w", err)
		}
	}
	if err := db.migrate(); err != nil {
		_ = db.inner.Close()
		return nil, err
	}
	return db, nil
}

// Close releases the handle.
func (d *DB) Close() error { return d.inner.Close() }

func (d *DB) migrate() error {
	var version int
	_, err := d.inner.Exec(`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("migration bootstrap: %w", err)
	}
	err = d.inner.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)
	if err != nil {
		return fmt.Errorf("migration version check: %w", err)
	}
	for version < currentVersion {
		next := version + 1
		stmts, ok := migrations[next]
		if !ok {
			return fmt.Errorf("missing migration %d", next)
		}
		tx, err := d.inner.Begin()
		if err != nil {
			return fmt.Errorf("migration %d begin: %w", next, err)
		}
		for _, s := range stmts {
			if _, err := tx.Exec(s); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("migration %d: %w", next, err)
			}
		}
		if _, err := tx.Exec(
			`INSERT INTO schema_version (version, applied_at) VALUES (?, ?)`,
			next, time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %d stamp: %w", next, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migration %d commit: %w", next, err)
		}
		version = next
	}
	return nil
}

// migrations maps version -> statements applied atomically.
var migrations = map[int][]string{
	1: {
		`CREATE TABLE provider_profiles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			base_url TEXT NOT NULL,
			model TEXT NOT NULL,
			credential_ref TEXT NOT NULL UNIQUE,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	},
	2: {
		`CREATE TABLE conversations (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'complete',
			provider_id TEXT,
			model TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX idx_messages_conv ON messages(conversation_id, created_at)`,
	},
	// v3: multiple models per provider. `models` is a JSON array of model
	// ids; `active_model` picks the one used for chat. Existing rows keep
	// their single `model` and are seeded into `models` as a one-item list
	// with `active_model` set to it, so pre-v3 data stays usable.
	3: {
		`ALTER TABLE provider_profiles ADD COLUMN models TEXT NOT NULL DEFAULT '[]'`,
		`ALTER TABLE provider_profiles ADD COLUMN active_model TEXT NOT NULL DEFAULT ''`,
		`UPDATE provider_profiles
			SET models = json_array(model), active_model = model
			WHERE models = '[]' OR models = ''`,
	},
}
