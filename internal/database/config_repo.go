package database

import (
	"database/sql"
	"fmt"
	"time"
)

type GuildConfig struct {
	WelcomeEnabled   bool
	WelcomeChannelID string
	WelcomeMessage   string
	UpdatedAt        time.Time
}

func UpsertGuildWelcomeConfig(db *sql.DB, guildID, channelID, message string) error {
	_, err := db.Exec(`INSERT INTO guild_configs (guild_id, welcome_enabled, welcome_channel_id, welcome_message, updated_at)
		VALUES (?, 1, ?, ?, datetime('now'))
		ON CONFLICT(guild_id) DO UPDATE SET
			welcome_enabled = 1,
			welcome_channel_id = excluded.welcome_channel_id,
			welcome_message = excluded.welcome_message,
			updated_at = datetime('now')`,
		guildID, channelID, message,
	)
	if err != nil {
		return fmt.Errorf("upsert welcome config for guild %s: %w", guildID, err)
	}
	return nil
}

func SetWelcomeEnabled(db *sql.DB, guildID string, enabled bool) error {
	val := 0
	if enabled {
		val = 1
	}
	result, err := db.Exec(`UPDATE guild_configs SET welcome_enabled = ?, updated_at = datetime('now') WHERE guild_id = ?`,
		val, guildID,
	)
	if err != nil {
		return fmt.Errorf("set welcome enabled for guild %s: %w", guildID, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no welcome config found for guild %s", guildID)
	}
	return nil
}

func GetGuildConfig(db *sql.DB, guildID string) (*GuildConfig, error) {
	var c GuildConfig
	var enabled int
	var updatedAt string
	err := db.QueryRow(`SELECT welcome_enabled, welcome_channel_id, welcome_message, updated_at
		FROM guild_configs WHERE guild_id = ?`, guildID).Scan(&enabled, &c.WelcomeChannelID, &c.WelcomeMessage, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get guild config %s: %w", guildID, err)
	}
	c.WelcomeEnabled = enabled == 1
	c.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &c, nil
}
