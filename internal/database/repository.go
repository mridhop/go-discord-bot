package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func UpsertGuild(db *sql.DB, g *discordgo.Guild) error {
	_, err := db.Exec(`INSERT OR REPLACE INTO guilds (guild_id, name, owner_id, member_count, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'))`,
		g.ID, g.Name, g.OwnerID, g.MemberCount,
	)
	if err != nil {
		return fmt.Errorf("upsert guild %s: %w", g.ID, err)
	}
	return nil
}

func UpsertUser(db *sql.DB, m *discordgo.Member) error {
	username := ""
	globalName := ""
	avatar := ""
	bot := 0
	if m.User != nil {
		username = m.User.Username
		globalName = m.User.GlobalName
		avatar = m.User.Avatar
		if m.User.Bot {
			bot = 1
		}
	}
	_, err := db.Exec(`INSERT OR REPLACE INTO users (user_id, username, global_name, avatar, bot, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))`,
		m.User.ID, username, globalName, avatar, bot,
	)
	if err != nil {
		return fmt.Errorf("upsert user %s: %w", m.User.ID, err)
	}
	return nil
}

func UpsertChannel(db *sql.DB, guildID string, c *discordgo.Channel) error {
	_, err := db.Exec(`INSERT OR REPLACE INTO channels (channel_id, guild_id, name, type, position, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))`,
		c.ID, guildID, c.Name, int(c.Type), c.Position,
	)
	if err != nil {
		return fmt.Errorf("upsert channel %s: %w", c.ID, err)
	}
	return nil
}

func UpsertRole(db *sql.DB, guildID string, r *discordgo.Role) error {
	managed := 0
	if r.Managed {
		managed = 1
	}
	_, err := db.Exec(`INSERT OR REPLACE INTO roles (role_id, guild_id, name, color, position, managed, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
		r.ID, guildID, r.Name, r.Color, r.Position, managed,
	)
	if err != nil {
		return fmt.Errorf("upsert role %s: %w", r.ID, err)
	}
	return nil
}

func UpsertAllGuildMembers(db *sql.DB, s *discordgo.Session, guildID string) (int, error) {
	total := 0
	after := ""
	for {
		members, err := s.GuildMembers(guildID, after, 1000)
		if err != nil {
			return total, fmt.Errorf("fetch members (after %s): %w", after, err)
		}
		if len(members) == 0 {
			break
		}
		for _, m := range members {
			if err := UpsertUser(db, m); err != nil {
				slog.Warn("failed to upsert user", "user_id", m.User.ID, "error", err)
				continue
			}
			total++
		}
		if len(members) < 1000 {
			break
		}
		after = members[len(members)-1].User.ID
	}
	return total, nil
}
