package middleware

import (
	"github.com/bwmarrin/discordgo"
)

type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

type Middleware func(CommandHandler) CommandHandler

func Chain(h CommandHandler, mw ...Middleware) CommandHandler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}
