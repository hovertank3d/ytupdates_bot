package main

import (
	"fmt"
	"log"
	"time"

	"io/ioutil"

	"github.com/BurntSushi/toml"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pborman/getopt/v2"
	"google.golang.org/api/youtube/v3"
)

type botConfig struct {
	Chats    []int64
	Apitoken string
	Youtube  youtubeConfig
}

var yt_service *youtube.Service

func loadConfig(path string) botConfig {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read config file: %v", err)
	}

	var config botConfig
	_, err = toml.Decode(string(b), &config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func potom_pridumaiu_nazvanie_xddd(bot *tgbotapi.BotAPI, config botConfig) {

	channels := loadChannelsInfo(yt_service, config.Youtube)

	for {
		updated_channels, videos := getNewVideos(channels)
		for i := 0; i < len(videos); i++ {
			for _, chat_id := range config.Chats {

				msg := tgbotapi.NewMessage(chat_id, fmt.Sprintf("%v published new [video](https://youtube.com/watch?v=%v)!", updated_channels[i].name, videos[i]))
				msg.ParseMode = tgbotapi.ModeMarkdown
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}
		}
		time.Sleep(config.Youtube.Cooldown)
	}
}

func main() {

	dir := getopt.StringLong("dir", 'd', "./", "working directory")

	getopt.Parse()

	config := loadConfig(*dir + "config.toml")

	bot, err := tgbotapi.NewBotAPI(config.Apitoken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	yt_service = initApi(*dir, config.Youtube)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	go potom_pridumaiu_nazvanie_xddd(bot, config)

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}
		if update.Message.IsCommand() {

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.Text = execCommand(update)
			if msg.Text != "" {
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.ReplyToMessageID = update.Message.MessageID
				if _, err := bot.Send(msg); err != nil {

					log.Panic(err)
				}
			}
		}
	}
}
