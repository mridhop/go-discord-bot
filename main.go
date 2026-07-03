package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/mridhop/go-discord-bot/internal/commands"
	"github.com/mridhop/go-discord-bot/internal/config"
	"github.com/mridhop/go-discord-bot/internal/logger"
	"github.com/mridhop/go-discord-bot/internal/middleware"
	"github.com/mridhop/go-discord-bot/internal/router"
)

func main() {
	godotenv.Load()

	logger.Setup()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	dg, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		slog.Error("error creating Discord session", "error", err)
		os.Exit(1)
	}

	dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged |
		discordgo.IntentGuildMembers |
		discordgo.IntentMessageContent

	r := router.New()
	r.Register(commands.PingCommand, middleware.Chain(commands.Ping, middleware.Recover, middleware.GuildOnly))
	dg.AddHandler(r.Handle)

	if cfg.AppID != "" {
		r.Sync(dg, cfg.AppID, cfg.GuildID)
	}

	if err = dg.Open(); err != nil {
		slog.Error("error opening connection", "error", err)
		os.Exit(1)
	}
	defer dg.Close()

	if cfg.GuildID != "" {
		slog.Info("bot is online", "mode", "guild (dev)")
	} else {
		slog.Info("bot is online", "mode", "global")
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	slog.Info("shutting down...")
}
