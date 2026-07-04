package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

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

func DeleteChannelsByGuild(db *sql.DB, guildID string) error {
	_, err := db.Exec(`DELETE FROM channels WHERE guild_id = ?`, guildID)
	if err != nil {
		return fmt.Errorf("delete channels for guild %s: %w", guildID, err)
	}
	return nil
}

func DeleteRolesByGuild(db *sql.DB, guildID string) error {
	_, err := db.Exec(`DELETE FROM roles WHERE guild_id = ?`, guildID)
	if err != nil {
		return fmt.Errorf("delete roles for guild %s: %w", guildID, err)
	}
	return nil
}

func UpsertStickyMessage(db *sql.DB, channelID, content string) error {
	_, err := db.Exec(`INSERT INTO sticky_messages (channel_id, content, last_message_id, updated_at)
		VALUES (?, ?, '', datetime('now'))
		ON CONFLICT(channel_id) DO UPDATE SET
			content = excluded.content,
			last_message_id = '',
			updated_at = datetime('now')`,
		channelID, content,
	)
	if err != nil {
		return fmt.Errorf("upsert sticky message for channel %s: %w", channelID, err)
	}
	return nil
}

type StickyMessage struct {
	Content         string
	LastMessageID   string
	CooldownSeconds int
	UpdatedAt       time.Time
}

func GetStickyMessage(db *sql.DB, channelID string) (*StickyMessage, error) {
	var m StickyMessage
	var updatedAt string
	err := db.QueryRow(`SELECT sm.content, sm.last_message_id, sc.seconds, sm.updated_at
		FROM sticky_messages sm
		JOIN sticky_cooldowns sc ON sc.id = sm.cooldown_id
		WHERE sm.channel_id = ?`, channelID).Scan(&m.Content, &m.LastMessageID, &m.CooldownSeconds, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sticky message for channel %s: %w", channelID, err)
	}
	m.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAt)
	if err != nil {
		m.UpdatedAt = time.Time{}
	}
	return &m, nil
}

func GetStickyCooldownByLabel(db *sql.DB, label string) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM sticky_cooldowns WHERE label = ?`, label).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid cooldown label: %s", label)
	}
	if err != nil {
		return 0, fmt.Errorf("get cooldown id for label %s: %w", label, err)
	}
	return id, nil
}

func UpdateStickyCooldown(db *sql.DB, channelID string, cooldownID int) error {
	_, err := db.Exec(`UPDATE sticky_messages SET cooldown_id = ?, updated_at = datetime('now') WHERE channel_id = ?`,
		cooldownID, channelID,
	)
	if err != nil {
		return fmt.Errorf("update sticky cooldown for channel %s: %w", channelID, err)
	}
	return nil
}

func UpdateStickyLastMessageID(db *sql.DB, channelID, messageID string) error {
	_, err := db.Exec(`UPDATE sticky_messages SET last_message_id = ?, updated_at = datetime('now') WHERE channel_id = ?`,
		messageID, channelID,
	)
	if err != nil {
		return fmt.Errorf("update sticky last_message_id for channel %s: %w", channelID, err)
	}
	return nil
}

func DeleteStickyMessage(db *sql.DB, channelID string) error {
	_, err := db.Exec(`DELETE FROM sticky_messages WHERE channel_id = ?`, channelID)
	if err != nil {
		return fmt.Errorf("delete sticky message for channel %s: %w", channelID, err)
	}
	return nil
}

func DeleteGuild(db *sql.DB, guildID string) error {
	if err := DeleteChannelsByGuild(db, guildID); err != nil {
		return err
	}
	if err := DeleteRolesByGuild(db, guildID); err != nil {
		return err
	}
	_, err := db.Exec(`DELETE FROM guilds WHERE guild_id = ?`, guildID)
	if err != nil {
		return fmt.Errorf("delete guild %s: %w", guildID, err)
	}
	return nil
}
