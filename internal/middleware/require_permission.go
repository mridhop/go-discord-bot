package middleware

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func RequirePermission(required int64) Middleware {
	return func(next CommandHandler) CommandHandler {
		return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member != nil && i.Member.Permissions&required == 0 && i.Member.Permissions&discordgo.PermissionAdministrator == 0 {
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You don't have permission to use this command.",
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
}
