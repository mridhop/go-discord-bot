package commands

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/mridhop/go-discord-bot/internal/database"
	"github.com/mridhop/go-discord-bot/internal/middleware"
)

var SyncServerCommand = &discordgo.ApplicationCommand{
	Name:                     "sync-server",
	Description:              "Sync all guild members, channels, and roles to the database",
	DefaultMemberPermissions: int64Ptr(discordgo.PermissionAdministrator),
}

func SyncServerHandler(db *sql.DB) middleware.CommandHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			slog.Error("failed to defer interaction", "error", err)
			return
		}

		guildID := i.GuildID

		guild, err := s.Guild(guildID)
		if err != nil {
			editResponse(s, i, fmt.Sprintf("Failed to fetch guild: %v", err))
			return
		}
		if err := database.UpsertGuild(db, guild); err != nil {
			slog.Error("failed to upsert guild", "error", err)
		}

		memberCount, err := database.UpsertAllGuildMembers(db, s, guildID)
		if err != nil {
			slog.Error("failed to upsert members", "error", err)
		}

		if err := database.DeleteChannelsByGuild(db, guildID); err != nil {
			slog.Error("failed to purge channels", "error", err)
		}

		channels, err := s.GuildChannels(guildID)
		if err != nil {
			editResponse(s, i, fmt.Sprintf("Failed to fetch channels: %v", err))
			return
		}
		channelCount := 0
		for _, c := range channels {
			if err := database.UpsertChannel(db, guildID, c); err != nil {
				slog.Warn("failed to upsert channel", "channel_id", c.ID, "error", err)
				continue
			}
			channelCount++
		}

		if err := database.DeleteRolesByGuild(db, guildID); err != nil {
			slog.Error("failed to purge roles", "error", err)
		}

		roles, err := s.GuildRoles(guildID)
		if err != nil {
			editResponse(s, i, fmt.Sprintf("Failed to fetch roles: %v", err))
			return
		}
		roleCount := 0
		for _, r := range roles {
			if err := database.UpsertRole(db, guildID, r); err != nil {
				slog.Warn("failed to upsert role", "role_id", r.ID, "error", err)
				continue
			}
			roleCount++
		}

		msg := fmt.Sprintf("Server synced.\n- Guild: %s\n- Members: %d\n- Channels: %d\n- Roles: %d",
			guild.Name, memberCount, channelCount, roleCount)
		editResponse(s, i, msg)
	}
}

func editResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		slog.Error("failed to edit interaction response", "error", err)
	}
}
