package main

import (
	"time"

	"github.com/toriato/katia"
	"github.com/toriato/katia-plugin-dcaptcha/command"
	"github.com/toriato/katia-plugin-dcaptcha/component"
	"github.com/toriato/katia-plugin-dcaptcha/model"
	"gorm.io/gorm"
)

var Plugin = katia.Plugin{
	Name:        "katia-plugin-dcaptcha",
	Description: "디시인사이드를 통한 서버 인증을 제공합니다",
	Author:      "Sangha Lee <totoriato@gmail.com>",
	Version:     [3]int{0, 1, 0},
	Depends: []string{
		"katia-plugin-database",
		"katia-plugin-cache"},

	OnEnable: func(bot *katia.Bot, plugin *katia.Plugin) error {
		command.CommandCreate.Plugin = plugin
		command.CommandCreateAction.Plugin = plugin
		component.ComponentCreate.Plugin = plugin

		bot.RegisterCommand(command.CommandCreate)
		bot.RegisterCommand(command.CommandCreateAction)
		bot.RegisterComponent(component.ComponentCreate)

		db := bot.Get("db").(*gorm.DB)

		if err := db.AutoMigrate(&model.Action{}); err != nil {
			return err
		}

		worker := Worker{}
		go func() {
			for {
				if err := worker.do(bot, plugin); err != nil {
					plugin.Logger.Error(err)
				}

				time.Sleep(5 * time.Second)
			}
		}()

		return nil
	},
}
