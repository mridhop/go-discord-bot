package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/mridhop/go-discord-bot/internal/database"
	"github.com/mridhop/go-discord-bot/internal/middleware"
)

var StickyCommand = &discordgo.ApplicationCommand{
	Name:        "sticky",
	Description: "Manage a sticky message that always stays at the bottom of the channel",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "set",
			Description: "Set the sticky message for this channel",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "JSON payload with content, embeds, and components",
					Required:    true,
				},
			},
		},
		{
			Name:        "cooldown",
			Description: "Set how often the sticky message reposts",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "value",
					Description: "Cooldown between sticky reposts",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "off", Value: "off"},
						{Name: "1s (1 second)", Value: "1s"},
						{Name: "5s (5 seconds)", Value: "5s"},
						{Name: "15s (15 seconds)", Value: "15s"},
						{Name: "30s (30 seconds)", Value: "30s"},
						{Name: "1m (1 minute)", Value: "1m"},
					},
				},
			},
		},
		{
			Name:        "remove",
			Description: "Remove the sticky message from this channel",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
	},
}

func StickyHandler(db *sql.DB) middleware.CommandHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ApplicationCommandData()
		options := data.Options
		if len(options) == 0 {
			return
		}

		switch options[0].Name {
		case "set":
			handleStickySet(s, i, db, options[0])
		case "cooldown":
			handleStickyCooldown(s, i, db, options[0])
		case "remove":
			handleStickyRemove(s, i, db)
		}
	}
}

func handleStickySet(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, opt *discordgo.ApplicationCommandInteractionDataOption) {
	msgOpt := opt.Options
	if len(msgOpt) == 0 {
		return
	}
	raw := msgOpt[0].StringValue()

	var payload sendMessagePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	if payload.Content == "" && len(payload.Embeds) == 0 && len(payload.Components) == 0 {
		respondEphemeral(s, i, "Message must have `content`, `embeds`, or `components`.")
		return
	}

	var msgComponents []discordgo.MessageComponent
	if len(payload.Components) > 0 {
		var compErr error
		msgComponents, compErr = convertComponents(payload.Components)
		if compErr != nil {
			respondEphemeral(s, i, fmt.Sprintf("Invalid components: %v", compErr))
			return
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("error deferring interaction", "error", err)
		return
	}

	edit := func(msg string) {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		if editErr != nil {
			slog.Error("error editing deferred response", "error", editErr)
		}
	}

	prev, err := database.GetStickyMessage(db, i.ChannelID)
	if err != nil {
		edit(fmt.Sprintf("Failed to check existing sticky: %v", err))
		return
	}

	if prev != nil && prev.LastMessageID != "" {
		if delErr := s.ChannelMessageDelete(i.ChannelID, prev.LastMessageID); delErr != nil {
			slog.Warn("failed to delete old sticky message", "message_id", prev.LastMessageID, "error", delErr)
		}
	}

	if err := database.UpsertStickyMessage(db, i.ChannelID, raw); err != nil {
		edit(fmt.Sprintf("Failed to save sticky message: %v", err))
		return
	}

	msg := &discordgo.MessageSend{
		Content: payload.Content,
	}
	for _, e := range payload.Embeds {
		e := e
		msg.Embeds = append(msg.Embeds, convertEmbed(&e))
	}
	msg.Components = msgComponents

	sent, err := s.ChannelMessageSendComplex(i.ChannelID, msg)
	if err != nil {
		edit(fmt.Sprintf("Failed to send sticky message: %v", err))
		return
	}

	if err := database.UpdateStickyLastMessageID(db, i.ChannelID, sent.ID); err != nil {
		slog.Warn("failed to update sticky last_message_id", "error", err)
	}

	edit("Sticky message set for this channel.")
}

func handleStickyCooldown(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, opt *discordgo.ApplicationCommandInteractionDataOption) {
	subOpts := opt.Options
	if len(subOpts) == 0 {
		return
	}
	label := subOpts[0].StringValue()

	cooldownID, err := database.GetStickyCooldownByLabel(db, label)
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}

	prev, err := database.GetStickyMessage(db, i.ChannelID)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to check sticky message: %v", err))
		return
	}
	if prev == nil {
		respondEphemeral(s, i, "No sticky message is set for this channel. Use `/sticky set` first.")
		return
	}

	if err := database.UpdateStickyCooldown(db, i.ChannelID, cooldownID); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to update cooldown: %v", err))
		return
	}

	respondEphemeral(s, i, fmt.Sprintf("Sticky cooldown set to `%s`.", label))
}

func handleStickyRemove(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	prev, err := database.GetStickyMessage(db, i.ChannelID)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to check sticky message: %v", err))
		return
	}
	if prev == nil {
		respondEphemeral(s, i, "No sticky message is set for this channel.")
		return
	}

	if prev.LastMessageID != "" {
		if delErr := s.ChannelMessageDelete(i.ChannelID, prev.LastMessageID); delErr != nil {
			slog.Warn("failed to delete sticky message", "message_id", prev.LastMessageID, "error", delErr)
		}
	}

	if err := database.DeleteStickyMessage(db, i.ChannelID); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to remove sticky message: %v", err))
		return
	}

	respondEphemeral(s, i, "Sticky message removed from this channel.")
}
