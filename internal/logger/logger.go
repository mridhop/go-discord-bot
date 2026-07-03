package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
)

func Setup() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	discordgo.Logger = func(msgL, caller int, format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		sourcePair := slog.String("source", "discordgo")

		switch msgL {
		case discordgo.LogError:
			slog.Error(msg, sourcePair)
		case discordgo.LogWarning:
			slog.Warn(msg, sourcePair)
		default:
			slog.Debug(msg, sourcePair)
		}
	}
}
