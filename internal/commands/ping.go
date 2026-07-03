package commands

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

var PingCommand = &discordgo.ApplicationCommand{
	Name:        "ping",
	Description: "Replies with Pong!",
}

func Ping(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})

	if err != nil {
		slog.Error("error responding to interaction", "error", err, "interaction", i)
	}
}
