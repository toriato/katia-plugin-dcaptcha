package command

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/toriato/katia"
	"github.com/toriato/katia-plugin-dcaptcha/model"
	"gorm.io/gorm"
)

func stoi(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

type CommandOptionCreateAction struct {
	LogChannel *discordgo.Channel `name:"log-channel" desc:"채널에 기록 보내기"`
	RoleGrant  *discordgo.Role    `name:"role-grant" desc:"역할 추가하기"`
	RoleRevoke *discordgo.Role    `name:"role-revoke" desc:"역할 제거하기"`
	Name       string             `name:"name" desc:"이름 설정하기"`
	NameReset  bool               `name:"name-reset" desc:"이름 초기화하기"`
	Kick       bool               `name:"kick" desc:"추방하기"`
}

var CommandCreateAction = &katia.Command{
	Name:        "create-action",
	Description: "사용자가 인증에 성공 했을 때 실행할 작업을 추가합니다",
	Options:     CommandOptionCreateAction{},
	OnExecute: func(ctx katia.CommandContext) interface{} {
		if ctx.Interaction.Member.Permissions&discordgo.PermissionAdministrator == 0 {
			return katia.ErrInteractionForbidden
		}

		if len(ctx.Interaction.ApplicationCommandData().Options) == 0 {
			return "실행할 행동이 없습니다"
		}

		options := ctx.Options.(CommandOptionCreateAction)
		action := model.Action{
			Guild:     stoi(ctx.Interaction.GuildID),
			CreatedBy: stoi(ctx.Interaction.Member.User.ID),
		}
		t := true

		if options.LogChannel != nil {
			i := stoi(options.LogChannel.ID)
			action.LogChannel = &i
		}
		if options.RoleGrant != nil {
			i := stoi(options.RoleGrant.ID)
			action.RoleGrant = &i
		}
		if options.RoleRevoke != nil {
			i := stoi(options.RoleRevoke.ID)
			action.RoleRevoke = &i
		}
		if options.Name != "" {
			action.Name = &options.Name
		}
		if options.NameReset {
			action.NameReset = &t
		}
		if options.Kick {
			action.Kick = &t
		}

		db := ctx.Bot.Get("db").(*gorm.DB)
		if err := db.Create(&action).Error; err != nil {
			return err
		}

		return "성공적으로 새 이벤트를 만들었습니다"
	},
}
