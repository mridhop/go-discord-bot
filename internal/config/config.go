package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppID    string
	BotToken string
	GuildID  string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppID:    os.Getenv("APP_ID"),
		BotToken: os.Getenv("DISCORD_BOT_TOKEN"),
		GuildID:  os.Getenv("DISCORD_GUILD_ID"),
	}

	if cfg.BotToken == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN is required")
	}

	return cfg, nil
}
