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
