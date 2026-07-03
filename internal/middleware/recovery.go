package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

func Recover(next CommandHandler) CommandHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic in command handler",
					"panic", r,
					"stack", string(debug.Stack()),
				)
			}
		}()
		next(s, i)
	}
}
