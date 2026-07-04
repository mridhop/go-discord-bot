package commands

import (
	"database/sql"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/mridhop/go-discord-bot/internal/database"
)

func GuildDeleteHandler(db *sql.DB) func(s *discordgo.Session, g *discordgo.GuildDelete) {
	return func(s *discordgo.Session, g *discordgo.GuildDelete) {
		if g.BeforeDelete != nil {
			slog.Info("bot removed from guild, cleaning up", "guild_id", g.ID, "guild_name", g.BeforeDelete.Name)
		} else {
			slog.Info("bot removed from guild, cleaning up", "guild_id", g.ID)
		}
		if err := database.DeleteGuild(db, g.ID); err != nil {
			slog.Error("failed to delete guild data", "guild_id", g.ID, "error", err)
		}
	}
}
