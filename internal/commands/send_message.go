package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type sendMessagePayload struct {
	Content    string                    `json:"content"`
	Embeds     []sendMessageEmbed        `json:"embeds"`
	Components []sendMessageComponentRow `json:"components"`
}

type embedColor int

func (c *embedColor) UnmarshalJSON(data []byte) error {
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*c = embedColor(i)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("color must be an integer or hex string")
	}

	s = strings.TrimPrefix(s, "#")
	val, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		return fmt.Errorf("invalid hex color: %s", s)
	}
	*c = embedColor(int(val))
	return nil
}

type sendMessageEmbed struct {
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Color       embedColor      `json:"color,omitempty"`
	Timestamp   string          `json:"timestamp,omitempty"`
	Footer      *embedFooter    `json:"footer,omitempty"`
	Thumbnail   *embedMedia     `json:"thumbnail,omitempty"`
	Image       *embedMedia     `json:"image,omitempty"`
	Author      *embedAuthor    `json:"author,omitempty"`
	Fields      []*embedField   `json:"fields,omitempty"`
}

type embedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

type embedMedia struct {
	URL string `json:"url"`
}

type embedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type sendMessageComponentRow struct {
	Type       int               `json:"type"`
	Components []json.RawMessage `json:"components"`
}

type sendMessageButton struct {
	Type     int                        `json:"type"`
	Style    int                        `json:"style"`
	Label    string                     `json:"label"`
	CustomID string                     `json:"custom_id,omitempty"`
	URL      string                     `json:"url,omitempty"`
	Disabled bool                       `json:"disabled,omitempty"`
	Emoji    *sendMessageComponentEmoji `json:"emoji,omitempty"`
}

type sendMessageSelectMenu struct {
	Type        int                        `json:"type"`
	CustomID    string                     `json:"custom_id"`
	Placeholder string                     `json:"placeholder,omitempty"`
	MinValues   *int                       `json:"min_values,omitempty"`
	MaxValues   int                        `json:"max_values,omitempty"`
	Disabled    bool                       `json:"disabled,omitempty"`
	Options     []sendMessageSelectOption  `json:"options,omitempty"`
}

type sendMessageSelectOption struct {
	Label       string                     `json:"label"`
	Value       string                     `json:"value"`
	Description string                     `json:"description,omitempty"`
	Emoji       *sendMessageComponentEmoji `json:"emoji,omitempty"`
	Default     bool                       `json:"default,omitempty"`
}

type sendMessageComponentEmoji struct {
	Name     string `json:"name,omitempty"`
	ID       string `json:"id,omitempty"`
	Animated bool   `json:"animated,omitempty"`
}

var SendMessageCommand = &discordgo.ApplicationCommand{
	Name:                     "send-message",
	Description:              "Sends a message with optional embed using a JSON payload",
	DefaultMemberPermissions: int64Ptr(discordgo.PermissionManageMessages),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "json",
			Description: "JSON payload with content and embed fields",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message-id",
			Description: "ID of the message to reply to",
			Required:    false,
		},
	},
}

func SendMessageHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	opt := data.GetOption("json")
	if opt == nil {
		return
	}
	raw := opt.StringValue()

	var payload sendMessagePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	if payload.Content == "" && len(payload.Embeds) == 0 && len(payload.Components) == 0 {
		respondEphemeral(s, i, "Message must have `content`, an `embed`, or `components`.")
		return
	}

	var msgComponents []discordgo.MessageComponent
	if len(payload.Components) > 0 {
		var compErr error
		msgComponents, compErr = convertComponents(payload.Components)
		if compErr != nil {
			respondEphemeral(s, i, fmt.Sprintf("Invalid components: %v", compErr))
			return
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("error deferring interaction", "error", err)
		return
	}

	msg := &discordgo.MessageSend{
		Content: payload.Content,
	}
	for _, e := range payload.Embeds {
		e := e
		msg.Embeds = append(msg.Embeds, convertEmbed(&e))
	}

	msg.Components = msgComponents

	if replyID := data.GetOption("message-id"); replyID != nil {
		msg.Reference = &discordgo.MessageReference{
			MessageID: replyID.StringValue(),
			ChannelID: i.ChannelID,
			GuildID:   i.GuildID,
		}
	}

	_, err = s.ChannelMessageSendComplex(i.ChannelID, msg)

	editMsg := "Message sent."
	if err != nil {
		slog.Error("error sending standalone message", "error", err)
		editMsg = fmt.Sprintf("Failed to send message: %v", err)
	}

	_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &editMsg,
	})
	if editErr != nil {
		slog.Error("error editing deferred response", "error", editErr)
	}
}

func convertEmbed(e *sendMessageEmbed) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       e.Title,
		Description: e.Description,
		URL:         e.URL,
		Color:       int(e.Color),
		Timestamp:   e.Timestamp,
	}

	if e.Footer != nil {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text:    e.Footer.Text,
			IconURL: e.Footer.IconURL,
		}
	}
	if e.Thumbnail != nil {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: e.Thumbnail.URL,
		}
	}
	if e.Image != nil {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: e.Image.URL,
		}
	}
	if e.Author != nil {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    e.Author.Name,
			URL:     e.Author.URL,
			IconURL: e.Author.IconURL,
		}
	}
	if len(e.Fields) > 0 {
		embed.Fields = make([]*discordgo.MessageEmbedField, len(e.Fields))
		for i, f := range e.Fields {
			embed.Fields[i] = &discordgo.MessageEmbedField{
				Name:   f.Name,
				Value:  f.Value,
				Inline: f.Inline,
			}
		}
	}

	return embed
}

func convertComponents(rows []sendMessageComponentRow) ([]discordgo.MessageComponent, error) {
	result := make([]discordgo.MessageComponent, 0, len(rows))
	for _, row := range rows {
		ar := &discordgo.ActionsRow{
			Components: make([]discordgo.MessageComponent, 0, len(row.Components)),
		}
		for _, raw := range row.Components {
			var typePeek struct {
				Type int `json:"type"`
			}
			if err := json.Unmarshal(raw, &typePeek); err != nil {
				return nil, fmt.Errorf("failed to parse component: %w", err)
			}
			switch typePeek.Type {
			case 2:
				var btn sendMessageButton
				if err := json.Unmarshal(raw, &btn); err != nil {
					return nil, fmt.Errorf("failed to parse button: %w", err)
				}
				ar.Components = append(ar.Components, convertButton(&btn))
			case 3, 5, 6, 7, 8:
				var sel sendMessageSelectMenu
				if err := json.Unmarshal(raw, &sel); err != nil {
					return nil, fmt.Errorf("failed to parse select menu: %w", err)
				}
				ar.Components = append(ar.Components, convertSelectMenu(&sel))
			default:
				return nil, fmt.Errorf("unknown component type: %d", typePeek.Type)
			}
		}
		result = append(result, ar)
	}
	return result, nil
}

func convertButton(b *sendMessageButton) *discordgo.Button {
	btn := &discordgo.Button{
		Style:    discordgo.ButtonStyle(b.Style),
		Label:    b.Label,
		CustomID: b.CustomID,
		URL:      b.URL,
		Disabled: b.Disabled,
	}
	if b.Emoji != nil {
		btn.Emoji = &discordgo.ComponentEmoji{
			Name:     b.Emoji.Name,
			ID:       b.Emoji.ID,
			Animated: b.Emoji.Animated,
		}
	}
	return btn
}

func convertSelectMenu(s *sendMessageSelectMenu) *discordgo.SelectMenu {
	menu := &discordgo.SelectMenu{
		MenuType:    discordgo.SelectMenuType(s.Type),
		CustomID:    s.CustomID,
		Placeholder: s.Placeholder,
		MinValues:   s.MinValues,
		MaxValues:   s.MaxValues,
		Disabled:    s.Disabled,
	}
	for _, opt := range s.Options {
		o := discordgo.SelectMenuOption{
			Label:       opt.Label,
			Value:       opt.Value,
			Description: opt.Description,
			Default:     opt.Default,
		}
		if opt.Emoji != nil {
			o.Emoji = &discordgo.ComponentEmoji{
				Name:     opt.Emoji.Name,
				ID:       opt.Emoji.ID,
				Animated: opt.Emoji.Animated,
			}
		}
		menu.Options = append(menu.Options, o)
	}
	return menu
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("error sending ephemeral response", "error", err)
	}
}
