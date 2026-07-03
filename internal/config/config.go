package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppID        string
	BotToken     string
	GuildID      string
	DatabasePath string
}

func Load() (*Config, error) {
	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		databasePath = "bot.db"
	}

	cfg := &Config{
		AppID:        os.Getenv("APP_ID"),
		BotToken:     os.Getenv("DISCORD_BOT_TOKEN"),
		GuildID:      os.Getenv("DISCORD_GUILD_ID"),
		DatabasePath: databasePath,
	}

	if cfg.BotToken == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN is required")
	}

	return cfg, nil
}
