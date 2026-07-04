package router

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mridhop/go-discord-bot/internal/middleware"
)

type PrefixRouter struct {
	prefix   string
	commands map[string]middleware.PrefixHandler
}

func NewPrefixRouter(prefix string) *PrefixRouter {
	return &PrefixRouter{
		prefix:   prefix,
		commands: make(map[string]middleware.PrefixHandler),
	}
}

func (r *PrefixRouter) Register(name string, h middleware.PrefixHandler) {
	r.commands[name] = h
}

func (r *PrefixRouter) Handle() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if !strings.HasPrefix(m.Content, r.prefix) {
			return
		}

		content := strings.TrimPrefix(m.Content, r.prefix)
		parts := strings.Fields(content)
		if len(parts) == 0 {
			return
		}

		name := strings.ToLower(parts[0])
		handler, ok := r.commands[name]
		if !ok {
			return
		}

		if handler == nil {
			slog.Warn("prefix command handler is nil", "command", name)
			return
		}

		handler(s, m)
	}
}
