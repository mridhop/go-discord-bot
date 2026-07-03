package router

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/mridhop/go-discord-bot/internal/middleware"
)

type commandEntry struct {
	command *discordgo.ApplicationCommand
	handler middleware.CommandHandler
}

type InteractionRouter struct {
	commands   map[string]commandEntry
	components map[string]middleware.CommandHandler
}

func New() *InteractionRouter {
	return &InteractionRouter{
		commands:   make(map[string]commandEntry),
		components: make(map[string]middleware.CommandHandler),
	}
}

func (r *InteractionRouter) Register(cmd *discordgo.ApplicationCommand, h middleware.CommandHandler) {
	r.commands[cmd.Name] = commandEntry{command: cmd, handler: h}
}

func (r *InteractionRouter) RegisterComponent(customID string, h middleware.CommandHandler) {
	r.components[customID] = h
}

func (r *InteractionRouter) Sync(s *discordgo.Session, appID string, guildID string) {
	cmds := make([]*discordgo.ApplicationCommand, 0, len(r.commands))
	for _, entry := range r.commands {
		cmds = append(cmds, entry.command)
	}

	registered, err := s.ApplicationCommandBulkOverwrite(appID, guildID, cmds)
	if err != nil {
		slog.Error("failed to register commands", "error", err)
		return
	}

	slog.Info("commands registered", "count", len(registered))
}

func (r *InteractionRouter) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		entry, ok := r.commands[data.Name]
		if !ok {
			return
		}
		entry.handler(s, i)

	case discordgo.InteractionMessageComponent:
		data := i.MessageComponentData()
		h, ok := r.components[data.CustomID]
		if !ok {
			return
		}
		h(s, i)
	}
}
