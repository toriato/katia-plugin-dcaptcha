package component

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"github.com/toriato/katia"
	"github.com/toriato/katia-plugin-dcaptcha/model"
)

func stoi(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

var ComponentCreate = &katia.Component{
	Name: "dcaptcha-create",
	OnExecute: func(ctx katia.ComponentContext) interface{} {
		member := ctx.Interaction.Member
		cache := ctx.Bot.Get("cache").(*cache.Cache)

		tokenValue := &model.Token{
			User:  stoi(member.User.ID),
			Guild: stoi(ctx.Interaction.GuildID),
		}

		// 토큰 캐시가 존재한다면 무시하기
		if _, ok := cache.Get(member.User.ID); ok {
			return katia.Error{Message: "인증 토큰을 만든 뒤 5분 후에 다시 인증할 수 있습니다"}
		}

		// 토큰 만들어 캐시하기
		var token string
		{
			b := make([]byte, 8)
			rand.Read(b)
			token = fmt.Sprintf("%x", b)
		}

		cache.Set(member.User.ID, tokenValue, 5*time.Minute)
		cache.Set(token, tokenValue, 5*time.Minute)

		return &discordgo.InteractionResponseData{
			Flags:   1 << 6,
			Content: "`" + token + "`",
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: "새로운 인증을 요청했습니다\n" +
						"디시인사이드에 로그인한 뒤 방명록에 인증 토큰을 올려주세요\n" +
						"인증 토큰은 5분간 유효합니다",
					Color: 0x4b59a7,
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "갤로그 방명록 열기",
							Style: discordgo.LinkButton,
							URL:   "https://gallog.dcinside.com/dcaptcha/guestbook"},
					},
				},
			},
		}

		return true
	},
}
