package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/toriato/katia"
	"github.com/toriato/katia-plugin-dcaptcha/model"
	"gorm.io/gorm"
)

var patternToken = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

type Worker struct {
	latestGuestbookID int64
}

func (worker *Worker) do(bot *katia.Bot, plugin *katia.Plugin) error {
	log := plugin.Logger.WithField("on", "worker")

	res, err := http.Get("https://gallog.dcinside.com/dcaptcha/guestbook")
	if err != nil {
		return err
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	if doc.Text() == "" {
		return errors.New("gallog.dcinside.com returns blank page")
	}

	var latestGuestbookID int64

	doc.Find("#gb_comments > li").Each(func(i int, s *goquery.Selection) {
		// 이미 처리한 글이라면 넘어가기
		id, _ := strconv.ParseInt(s.AttrOr("data-no", ""), 10, 64)
		if worker.latestGuestbookID >= id {
			return
		}

		// 현재 루프 기준으로 가장 최근에 올라온 방명록이라면 아이디 보관하기
		if latestGuestbookID < id {
			latestGuestbookID = id
		}

		// 방명록 내용이 토큰 패턴과 일치하지 않는다면 넘어가기
		content := s.Find(".memo").Text()
		if !patternToken.MatchString(content) {
			return
		}

		// 캐시에서 토큰이 존재하는지 확인하기
		var token *model.Token

		cache := bot.Get("cache").(*cache.Cache)
		if value, exists := cache.Get(content); exists {
			token = value.(*model.Token)
		} else {
			log.Debugf("%s is invalid token", content)
			return
		}

		context := map[string]string{}
		userID := strconv.FormatInt(token.User, 10)
		guildID := strconv.FormatInt(token.Guild, 10)

		// 사용자 정보 파싱하기
		{
			ref := s.Find(".writer_info")
			src := ref.Find(".writer_nikcon img").AttrOr("onclick", "")
			parts := strings.SplitN(src, "'", 3)
			context["username"] = parts[1][1:] // 앞에 붙은 슬래시 제외하기
			context["nickname"] = ref.Find(".nickname").AttrOr("title", "")
		}

		// 가입한 사용자가 아니라면 넘어가기
		if context["username"] == "" {
			return
		}

		log.WithFields(
			logrus.Fields{
				"user":    userID,
				"guild":   guildID,
				"context": context,
			},
		).Infof("%s synced with %s", userID, context["username"])

		// 캐시에서 토큰 제거하기
		cache.Delete(userID)
		cache.Delete(content)

		// 길드 별 액션 가져오기
		// TODO: 길드 별 액션 캐시하기
		var actions []model.Action

		db := bot.Get("db").(*gorm.DB)
		if err := db.Where("guild = ?", token.Guild).Find(&actions).Error; err != nil {
			log.Error(err)
			return
		}

		// 액션이 없다면 넘어가기
		if len(actions) < 1 {
			return
		}

		session := bot.Session

		// 액션 실행하기
		for _, action := range actions {
			if action.LogChannel != nil {
				channelID := strconv.FormatInt(*action.LogChannel, 10)
				embed := &discordgo.MessageEmbed{
					Description: fmt.Sprintf(
						"<@!%s> 님이 성공적으로 인증됐습니다\n ┗ [[갤로그]](https://gallog.dcinside.com/%s) [[구글]](https://www.google.com/search?q=site%%3Adcinside.com+%%22%s%%22)",
						userID,
						context["username"],
						context["username"]),
					Fields: []*discordgo.MessageEmbedField{
						{Inline: true, Name: "닉네임", Value: context["nickname"]},
						{Inline: true, Name: "아이디", Value: context["username"]},
					},
				}
				if _, err := session.ChannelMessageSendEmbed(channelID, embed); err != nil {
					log.
						WithFields(logrus.Fields{
							"action":  "LogChannel",
							"channel": channelID,
						}).
						Error(err)
				}
			}
			if action.RoleGrant != nil {
				roleID := strconv.FormatInt(*action.RoleGrant, 10)
				if err := session.GuildMemberRoleAdd(guildID, userID, roleID); err != nil {
					log.WithField("action", "RoleGrant").Error(err)
				}
			}
			if action.RoleRevoke != nil {
				roleID := strconv.FormatInt(*action.RoleRevoke, 10)
				if err := session.GuildMemberRoleRemove(guildID, userID, roleID); err != nil {
					log.WithField("action", "RoleRevoke").Error(err)
				}
			}
			if action.Name != nil {
				args := []string{}
				for key, value := range context {
					args = append(args, "{"+key+"}", value)
				}

				nickname := strings.NewReplacer(args...).Replace(*action.Name)
				if err := session.GuildMemberNickname(guildID, userID, nickname); err != nil {
					log.WithField("action", "Name").Error(err)
				}
			}
			if action.NameReset != nil {
				if err := session.GuildMemberNickname(guildID, userID, ""); err != nil {
					log.WithField("action", "NameReset").Error(err)
				}
			}
			if action.Kick != nil {
				if err := session.GuildMemberDelete(guildID, userID); err != nil {
					log.WithField("action", "Kick").Error(err)
				}
			}
		}
	})

	if worker.latestGuestbookID < latestGuestbookID {
		worker.latestGuestbookID = latestGuestbookID
	}

	return nil
}
