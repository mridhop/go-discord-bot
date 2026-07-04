package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

var GetMessageJSONCommand = &discordgo.ApplicationCommand{
	Name:                     "get-message-as-json",
	Description:              "Gets a bot message as a JSON payload for editing",
	DefaultMemberPermissions: int64Ptr(discordgo.PermissionManageMessages),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message-id",
			Description: "ID of the bot message to fetch",
			Required:    true,
		},
	},
}

type getMessageOutput struct {
	Content    string            `json:"content"`
	Embeds     []json.RawMessage `json:"embeds,omitempty"`
	Components []json.RawMessage `json:"components,omitempty"`
}

func GetMessageJSONHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	msgIDOpt := data.GetOption("message-id")
	if msgIDOpt == nil {
		return
	}
	msgID := msgIDOpt.StringValue()

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

	editMsg := func(edit *discordgo.WebhookEdit) {
		_, editErr := s.InteractionResponseEdit(i.Interaction, edit)
		if editErr != nil {
			slog.Error("error editing deferred response", "error", editErr)
		}
	}

	strPtr := func(s string) *string { return &s }

	msg, err := s.ChannelMessage(i.ChannelID, msgID)
	if err != nil {
		editMsg(&discordgo.WebhookEdit{
			Content: strPtr(fmt.Sprintf("Failed to fetch message: %v", err)),
		})
		return
	}

	if msg.Author == nil || msg.Author.ID != s.State.User.ID {
		editMsg(&discordgo.WebhookEdit{
			Content: strPtr("That message does not belong to this bot."),
		})
		return
	}

	output := getMessageOutput{
		Content: msg.Content,
	}

	for _, embed := range msg.Embeds {
		eJSON, err := json.Marshal(embed)
		if err != nil {
			editMsg(&discordgo.WebhookEdit{
				Content: strPtr(fmt.Sprintf("Failed to encode embed: %v", err)),
			})
			return
		}
		output.Embeds = append(output.Embeds, eJSON)
	}

	for _, comp := range msg.Components {
		cJSON, err := json.Marshal(comp)
		if err != nil {
			editMsg(&discordgo.WebhookEdit{
				Content: strPtr(fmt.Sprintf("Failed to encode component: %v", err)),
			})
			return
		}
		output.Components = append(output.Components, cJSON)
	}

	payload, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		editMsg(&discordgo.WebhookEdit{
			Content: strPtr(fmt.Sprintf("Failed to encode message: %v", err)),
		})
		return
	}

	editMsg(&discordgo.WebhookEdit{
		Content: strPtr("Here is the JSON payload:"),
		Files: []*discordgo.File{
			{
				Name:        "message.json",
				ContentType: "application/json",
				Reader:      bytes.NewReader(payload),
			},
		},
	})
}
