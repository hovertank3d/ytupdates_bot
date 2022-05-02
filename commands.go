package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type (
	command struct {
		name    string
		handler commandHandler
	}
	commandHandler func(update tgbotapi.Update) string
)

var commands = [...]command{
	{"ping", ping},
	{"channelid", channelid},
	{"chatid", chatid},
}

func execCommand(update tgbotapi.Update) string {

	for _, cmd := range commands {
		if cmd.name == update.Message.Command() {
			return cmd.handler(update)
		}
	}
	return ""
}

func ping(update tgbotapi.Update) string {

	return "still alive"
}

func channelid(update tgbotapi.Update) string {

	if update.Message.CommandArguments() == "" {
		return "да."
	}

	return getChannelId(yt_service, update.Message.CommandArguments())
}

func chatid(update tgbotapi.Update) string {

	return fmt.Sprintf("%d", update.Message.Chat.ID)
}
