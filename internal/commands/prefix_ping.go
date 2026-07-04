package commands

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func PrefixPing(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if err != nil {
		slog.Error("error sending prefix ping response", "error", err)
	}
}
