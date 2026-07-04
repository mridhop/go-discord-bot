package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type editMessagePayload struct {
	Content    *string                   `json:"content"`
	Embeds     []*sendMessageEmbed       `json:"embeds,omitempty"`
	Components []sendMessageComponentRow `json:"components,omitempty"`
}

var EditMessageCommand = &discordgo.ApplicationCommand{
	Name:                     "edit-message",
	Description:              "Edits a bot message by its ID using a JSON payload",
	DefaultMemberPermissions: int64Ptr(discordgo.PermissionManageMessages),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message-id",
			Description: "ID of the bot message to edit",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "json",
			Description: "JSON payload with content, embeds, and components",
			Required:    true,
		},
	},
}

func EditMessageHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	msgIDOpt := data.GetOption("message-id")
	jsonOpt := data.GetOption("json")
	if msgIDOpt == nil || jsonOpt == nil {
		return
	}

	msgID := msgIDOpt.StringValue()
	raw := jsonOpt.StringValue()

	var payload editMessagePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	if payload.Content == nil && payload.Embeds == nil && payload.Components == nil {
		respondEphemeral(s, i, "Payload must have `content`, `embeds`, or `components`.")
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

	msgEdit := discordgo.NewMessageEdit(i.ChannelID, msgID)

	if payload.Content != nil {
		msgEdit.SetContent(*payload.Content)
	}

	if payload.Embeds != nil {
		embeds := make([]*discordgo.MessageEmbed, len(payload.Embeds))
		for idx, e := range payload.Embeds {
			embeds[idx] = convertEmbed(e)
		}
		msgEdit.SetEmbeds(embeds)
	}

	if payload.Components != nil {
		comps, compErr := convertComponents(payload.Components)
		if compErr != nil {
			editMsg := fmt.Sprintf("Invalid components: %v", compErr)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &editMsg})
			return
		}
		msgEdit.Components = &comps
	}

	edited, err := s.ChannelMessageEditComplex(msgEdit)

	editMsg := "Message edited."
	if err != nil {
		slog.Error("error editing message", "error", err)
		editMsg = fmt.Sprintf("Failed to edit message: %v", err)
	} else if edited != nil {
		editMsg = fmt.Sprintf("Message edited: https://discord.com/channels/%s/%s/%s", i.GuildID, i.ChannelID, edited.ID)
	}

	_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &editMsg,
	})
	if editErr != nil {
		slog.Error("error editing deferred response", "error", editErr)
	}
}
