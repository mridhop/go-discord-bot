package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type sendMessagePayload struct {
	Content string            `json:"content"`
	Embed   *sendMessageEmbed `json:"embed"`
}

type sendMessageEmbed struct {
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Color       int             `json:"color,omitempty"`
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

var SendMessageCommand = &discordgo.ApplicationCommand{
	Name:        "send-message",
	Description: "Sends a message with optional embed using a JSON payload",
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

	if payload.Content == "" && payload.Embed == nil {
		respondEphemeral(s, i, "Message must have `content` or an `embed`.")
		return
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
	if payload.Embed != nil {
		msg.Embeds = []*discordgo.MessageEmbed{convertEmbed(payload.Embed)}
	}

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
		Color:       e.Color,
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
