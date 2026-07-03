package middleware

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func GuildOnly(next CommandHandler) CommandHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.GuildID == "" {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "This command can only be used in a server.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})

			if err != nil {
				slog.Error("error responding to interaction", "error", err, "interaction", i)
			}

			return
		}
		next(s, i)
	}
}
