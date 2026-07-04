package middleware

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func PrefixGuildOnly(next PrefixHandler) PrefixHandler {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.GuildID == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "This command can only be used in a server.")
			if err != nil {
				slog.Error("error sending guild-only message", "error", err)
			}
			return
		}
		next(s, m)
	}
}
