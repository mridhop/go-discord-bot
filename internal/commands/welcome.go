package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mridhop/go-discord-bot/internal/database"
	"github.com/mridhop/go-discord-bot/internal/middleware"
)

var WelcomeCommand = &discordgo.ApplicationCommand{
	Name:                     "welcome",
	Description:              "Manage welcome messages for new members",
	DefaultMemberPermissions: int64Ptr(discordgo.PermissionManageGuild),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "set",
			Description: "Set the welcome message for this channel",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "json",
					Description: "JSON payload with content and embeds (supports {user}, {user_name}, {time} etc.)",
					Required:    true,
				},
			},
		},
		{
			Name:        "toggle",
			Description: "Toggle welcome messages on or off",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
	},
}

func WelcomeHandler(db *sql.DB) middleware.CommandHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ApplicationCommandData()
		options := data.Options
		if len(options) == 0 {
			return
		}

		switch options[0].Name {
		case "set":
			handleWelcomeSet(s, i, db, options[0])
		case "toggle":
			handleWelcomeToggle(s, i, db)
		}
	}
}

func handleWelcomeSet(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, opt *discordgo.ApplicationCommandInteractionDataOption) {
	subOpts := opt.Options
	if len(subOpts) == 0 {
		return
	}
	raw := subOpts[0].StringValue()

	var payload sendMessagePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	if payload.Content == "" && len(payload.Embeds) == 0 {
		respondEphemeral(s, i, "Message must have `content` or `embeds`.")
		return
	}

	if err := database.UpsertGuildWelcomeConfig(db, i.GuildID, i.ChannelID, raw); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to save welcome message: %v", err))
		return
	}

	respondEphemeral(s, i, "Welcome message set and enabled in this channel.")
}

func handleWelcomeToggle(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	cfg, err := database.GetGuildConfig(db, i.GuildID)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to check welcome config: %v", err))
		return
	}
	if cfg == nil {
		respondEphemeral(s, i, "No welcome message is set for this server. Use `/welcome set` first.")
		return
	}

	newState := !cfg.WelcomeEnabled
	if err := database.SetWelcomeEnabled(db, i.GuildID, newState); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to update: %v", err))
		return
	}

	stateStr := "disabled"
	if newState {
		stateStr = "enabled"
	}
	respondEphemeral(s, i, fmt.Sprintf("Welcome messages %s.", stateStr))
}
