package command

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/toriato/katia"
)

type CommandOptionCreate struct {
	Label   string `name:"label" desc:"버튼 내용"`
	Content string `name:"content" desc:"메세지 내용"`
	Color   string `name:"color" desc:"메세지 색상"`
}

var CommandCreate = &katia.Command{
	Name:        "create",
	Description: "사용자가 인증을 시작할 수 있는 메세지 상자를 만듭니다",
	Options:     CommandOptionCreate{},
	OnExecute: func(ctx katia.CommandContext) interface{} {
		if ctx.Interaction.Member.Permissions&discordgo.PermissionAdministrator == 0 {
			return katia.ErrInteractionForbidden
		}

		options := ctx.Options.(CommandOptionCreate)

		button := discordgo.Button{CustomID: "dcaptcha-create", Label: "인증하기"}
		embed := &discordgo.MessageEmbed{
			Description: "아래 인증하기 버튼을 눌러 디시인사이드 계정을 인증해주세요",
			Color:       0x4b59a7,
		}

		if options.Label != "" {
			button.Label = options.Label
		}

		if options.Content != "" {
			embed.Description = options.Content
		}

		if options.Color != "" {
			color, err := strconv.ParseInt(options.Color, 16, 32)
			if err != nil {
				return "색상 값은 16진수여야만 합니다 (예: FFA0A0)"
			}

			embed.Color = int(color)
		}

		return &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{button},
				},
			},
		}
	},
}
