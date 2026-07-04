package commands

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mridhop/go-discord-bot/internal/database"
)

func StickyMessageHandler(db *sql.DB) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		sticky, err := database.GetStickyMessage(db, m.ChannelID)
		if err != nil {
			slog.Error("failed to get sticky message", "channel_id", m.ChannelID, "error", err)
			return
		}
		if sticky == nil {
			return
		}

		if sticky.CooldownSeconds > 0 {
			elapsed := time.Since(sticky.UpdatedAt)
			if elapsed < time.Duration(sticky.CooldownSeconds)*time.Second {
				return
			}
		}

		if sticky.LastMessageID != "" {
			if delErr := s.ChannelMessageDelete(m.ChannelID, sticky.LastMessageID); delErr != nil {
				slog.Warn("failed to delete old sticky message", "message_id", sticky.LastMessageID, "error", delErr)
			}
		}

		var payload sendMessagePayload
		if err := json.Unmarshal([]byte(sticky.Content), &payload); err != nil {
			slog.Error("failed to parse sticky message JSON", "channel_id", m.ChannelID, "error", err)
			return
		}

		var msgComponents []discordgo.MessageComponent
		if len(payload.Components) > 0 {
			var compErr error
			msgComponents, compErr = convertComponents(payload.Components)
			if compErr != nil {
				slog.Error("failed to convert sticky components", "channel_id", m.ChannelID, "error", compErr)
				return
			}
		}

		msg := &discordgo.MessageSend{
			Content: payload.Content,
		}
		for _, e := range payload.Embeds {
			e := e
			msg.Embeds = append(msg.Embeds, convertEmbed(&e))
		}
		msg.Components = msgComponents

		sent, err := s.ChannelMessageSendComplex(m.ChannelID, msg)
		if err != nil {
			slog.Error("failed to send sticky message", "channel_id", m.ChannelID, "error", err)
			return
		}

		if err := database.UpdateStickyLastMessageID(db, m.ChannelID, sent.ID); err != nil {
			slog.Warn("failed to update sticky last_message_id", "error", err)
		}
	}
}
