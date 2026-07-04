package database

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type Migration struct {
	Version int
	Name    string
	Up      string
}

func Migrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_schema_migrations",
			Up: `CREATE TABLE IF NOT EXISTS schema_migrations (
			version   INTEGER NOT NULL,
			name      TEXT    NOT NULL,
			applied_at TEXT   NOT NULL DEFAULT (datetime('now')),
			PRIMARY KEY (version, name)
		)`,
		},
		{
			Version: 1,
			Name:    "create_users",
			Up: `CREATE TABLE IF NOT EXISTS users (
				user_id TEXT PRIMARY KEY,
				created_at TEXT NOT NULL DEFAULT (datetime('now'))
			)`,
		},
		{
			Version: 1,
			Name:    "create_guilds",
			Up: `CREATE TABLE IF NOT EXISTS guilds (
				guild_id TEXT PRIMARY KEY,
				created_at TEXT NOT NULL DEFAULT (datetime('now'))
			)`,
		},
		{
			Version: 1,
			Name:    "create_channels",
			Up: `CREATE TABLE IF NOT EXISTS channels (
				channel_id TEXT PRIMARY KEY,
				created_at TEXT NOT NULL DEFAULT (datetime('now'))
			)`,
		},
		{
			Version: 1,
			Name:    "create_messages",
			Up: `CREATE TABLE IF NOT EXISTS messages (
				message_id TEXT PRIMARY KEY,
				created_at TEXT NOT NULL DEFAULT (datetime('now'))
			)`,
		},
		{
			Version: 1,
			Name:    "create_roles",
			Up: `CREATE TABLE IF NOT EXISTS roles (
				role_id TEXT PRIMARY KEY,
				created_at TEXT NOT NULL DEFAULT (datetime('now'))
			)`,
		},
		{
			Version: 2,
			Name:    "add_guild_columns",
			Up: `ALTER TABLE guilds ADD COLUMN name TEXT NOT NULL DEFAULT '';
			ALTER TABLE guilds ADD COLUMN owner_id TEXT NOT NULL DEFAULT '';
			ALTER TABLE guilds ADD COLUMN member_count INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE guilds ADD COLUMN updated_at TEXT NOT NULL DEFAULT (datetime('now'));`,
		},
		{
			Version: 2,
			Name:    "add_user_columns",
			Up: `ALTER TABLE users ADD COLUMN username TEXT NOT NULL DEFAULT '';
			ALTER TABLE users ADD COLUMN global_name TEXT NOT NULL DEFAULT '';
			ALTER TABLE users ADD COLUMN avatar TEXT NOT NULL DEFAULT '';
			ALTER TABLE users ADD COLUMN bot INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE users ADD COLUMN updated_at TEXT NOT NULL DEFAULT (datetime('now'));`,
		},
		{
			Version: 2,
			Name:    "add_channel_columns",
			Up: `ALTER TABLE channels ADD COLUMN guild_id TEXT NOT NULL DEFAULT '';
			ALTER TABLE channels ADD COLUMN name TEXT NOT NULL DEFAULT '';
			ALTER TABLE channels ADD COLUMN type INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE channels ADD COLUMN position INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE channels ADD COLUMN updated_at TEXT NOT NULL DEFAULT (datetime('now'));`,
		},
		{
			Version: 2,
			Name:    "add_role_columns",
			Up: `ALTER TABLE roles ADD COLUMN guild_id TEXT NOT NULL DEFAULT '';
			ALTER TABLE roles ADD COLUMN name TEXT NOT NULL DEFAULT '';
			ALTER TABLE roles ADD COLUMN color INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE roles ADD COLUMN position INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE roles ADD COLUMN managed INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE roles ADD COLUMN updated_at TEXT NOT NULL DEFAULT (datetime('now'));`,
		},
		{
			Version: 3,
			Name:    "create_sticky_messages",
			Up: `CREATE TABLE IF NOT EXISTS sticky_messages (
				channel_id      TEXT PRIMARY KEY,
				content         TEXT NOT NULL,
				last_message_id TEXT NOT NULL DEFAULT '',
				created_at      TEXT NOT NULL DEFAULT (datetime('now')),
				updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
			)`,
		},
		{
			Version: 4,
			Name:    "create_sticky_cooldowns",
			Up: `CREATE TABLE IF NOT EXISTS sticky_cooldowns (
				id      INTEGER PRIMARY KEY,
				seconds INTEGER NOT NULL,
				label   TEXT NOT NULL UNIQUE
			);
			INSERT OR IGNORE INTO sticky_cooldowns (id, seconds, label) VALUES
				(0, 0, 'off'),
				(1, 1, '1s'),
				(2, 5, '5s'),
				(3, 15, '15s'),
				(4, 30, '30s'),
				(5, 60, '1m');`,
		},
		{
			Version: 4,
			Name:    "add_sticky_cooldown_id",
			Up: `ALTER TABLE sticky_messages ADD COLUMN cooldown_id INTEGER NOT NULL DEFAULT 0`,
		},
	}
}

func Migrate(db *sql.DB) error {
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return err
	}

	migrations := Migrations()

	for _, m := range migrations {
		applied, err := isApplied(db, m.Version, m.Name)
		if err != nil {
			return fmt.Errorf("check migration %d (%s): %w", m.Version, m.Name, err)
		}
		if applied {
			continue
		}

		slog.Info("applying migration", "version", m.Version, "name", m.Name)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %w", m.Version, err)
		}
		defer tx.Rollback()

		if _, err := tx.Exec(m.Up); err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", m.Version, m.Name, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			m.Version, m.Name,
		); err != nil {
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}
	}

	return nil
}

func ensureSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version   INTEGER NOT NULL,
		name      TEXT    NOT NULL,
		applied_at TEXT   NOT NULL DEFAULT (datetime('now')),
		PRIMARY KEY (version, name)
	)`)
	return err
}

func isApplied(db *sql.DB, version int, name string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ? AND name = ?)", version, name).Scan(&exists)
	return exists, err
}
