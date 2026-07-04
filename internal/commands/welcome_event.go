package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mridhop/go-discord-bot/internal/database"
)

func WelcomeMemberHandler(db *sql.DB) func(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
		if e.Member.User.Bot {
			return
		}

		cfg, err := database.GetGuildConfig(db, e.GuildID)
		if err != nil {
			slog.Error("failed to get guild config for welcome", "guild_id", e.GuildID, "error", err)
			return
		}
		if cfg == nil || !cfg.WelcomeEnabled {
			return
		}

		raw := cfg.WelcomeMessage

		guildName := ""
		memberCount := ""
		if g, err := s.State.Guild(e.GuildID); err == nil {
			guildName = g.Name
			memberCount = fmt.Sprintf("%d", g.MemberCount)
		}

		user := e.Member.User
		// loc := time.FixedZone("GMT+7", 7*60*60)
		timeStr := time.Now().Format(time.RFC3339)

		replacements := map[string]string{
			"{user}":             fmt.Sprintf("<@%s>", user.ID),
			"{user_name}":        user.Username,
			"{user_global_name}": user.GlobalName,
			"{user_photo}":       user.AvatarURL("256"),
			"{time}":             timeStr,
			"{server}":           guildName,
			"{member_count}":     memberCount,
		}

		for placeholder, value := range replacements {
			raw = strings.ReplaceAll(raw, placeholder, value)
		}

		var payload sendMessagePayload
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			slog.Error("failed to parse welcome message json after replacements", "guild_id", e.GuildID, "error", err)
			return
		}

		msg := &discordgo.MessageSend{
			Content: payload.Content,
		}
		for _, emb := range payload.Embeds {
			emb := emb
			msg.Embeds = append(msg.Embeds, convertEmbed(&emb))
		}
		if len(payload.Components) > 0 {
			if components, err := convertComponents(payload.Components); err == nil {
				msg.Components = components
			}
		}

		_, err = s.ChannelMessageSendComplex(cfg.WelcomeChannelID, msg)
		if err != nil {
			slog.Error("failed to send welcome message", "guild_id", e.GuildID, "channel_id", cfg.WelcomeChannelID, "error", err)
		}
	}
}
