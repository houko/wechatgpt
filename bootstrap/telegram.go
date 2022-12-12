package bootstrap

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"github.com/wechatgpt/wechatbot/config"
	"github.com/wechatgpt/wechatbot/handler/telegram"
	"time"
)

func StartTelegramBot() {
	getConfig := config.GetConfig()
	if getConfig == nil {
		return
	}
	botConfig := getConfig.ChatGpt
	bot, err := tgbotapi.NewBotAPI(botConfig.Telegram)
	if err != nil {
		return
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)

	updates := bot.GetUpdatesChan(u)
	time.Sleep(time.Millisecond * 500)
	for len(updates) != 0 {
		<-updates
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		text := update.Message.Text
		chatID := update.Message.Chat.ID
		responseMsg := telegram.Handle(text)
		if responseMsg == nil {
			continue
		}
		msg := tgbotapi.NewMessage(chatID, *responseMsg)
		send, err := bot.Send(msg)
		if err != nil {
			log.Errorf("发送消息出错:%s", err.Error())
			return
		}
		fmt.Println(send.Text)
	}
}
