package commands

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

var EmbedDemoCommand = &discordgo.ApplicationCommand{
	Name:        "embed-demo",
	Description: "Sends an embed with interactive buttons",
}

func EmbedDemoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Interactive Embed Demo",
					Description: "This is an embed with buttons. Click one of the buttons below.",
					Color:       0x5865F2,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Confirm",
							Value:  "Click the green button to confirm.",
							Inline: true,
						},
						{
							Name:   "Cancel",
							Value:  "Click the red button to cancel.",
							Inline: true,
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Embeds & Components Demo",
					},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Confirm",
							Style:    discordgo.SuccessButton,
							CustomID: "embed_demo_confirm",
						},
						discordgo.Button{
							Label:    "Cancel",
							Style:    discordgo.DangerButton,
							CustomID: "embed_demo_cancel",
						},
					},
				},
			},
		},
	})
	if err != nil {
		slog.Error("error responding to interaction", "error", err)
	}
}

func EmbedDemoConfirmHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.Member.User
	if user == nil {
		user = i.User
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Confirmed!",
					Description: fmt.Sprintf("%s confirmed the action.", user.Mention()),
					Color:       0x57F287,
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Embeds & Components Demo",
					},
				},
			},
			Components: []discordgo.MessageComponent{},
		},
	})
	if err != nil {
		slog.Error("error responding to component interaction", "error", err)
	}
}

func EmbedDemoCancelHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.Member.User
	if user == nil {
		user = i.User
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Cancelled",
					Description: fmt.Sprintf("%s cancelled the action.", user.Mention()),
					Color:       0xED4245,
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Embeds & Components Demo",
					},
				},
			},
			Components: []discordgo.MessageComponent{},
		},
	})
	if err != nil {
		slog.Error("error responding to component interaction", "error", err)
	}
}
