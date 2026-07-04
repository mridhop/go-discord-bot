package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

func PrefixRecover(next PrefixHandler) PrefixHandler {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic in prefix command handler",
					"panic", r,
					"stack", string(debug.Stack()),
				)
			}
		}()
		next(s, m)
	}
}
