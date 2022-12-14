package bootstrap

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"github.com/wechatgpt/wechatbot/config"
	"github.com/wechatgpt/wechatbot/handler/telegram"
	"os"
	"time"
)

func StartTelegramBot() {
	telegramKey := os.Getenv("telegram")
	if len(telegramKey) == 0 {
		getConfig := config.GetConfig()
		if getConfig == nil {
			return
		}
		botConfig := getConfig.ChatGpt
		telegramKey = botConfig.Telegram
		log.Info("读取本地本置文件中的telegram token:", telegramKey)
	} else {
		log.Info("找到环境变量: telegram token:", telegramKey)
	}
	bot, err := tgbotapi.NewBotAPI(telegramKey)
	if err != nil {
		return
	}

	bot.Debug = false
	log.Info("Authorized on account: ", bot.Self.UserName)
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
