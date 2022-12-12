package telegram

import (
	log "github.com/sirupsen/logrus"
	"github.com/wechatgpt/wechatbot/config"
	"github.com/wechatgpt/wechatbot/openai"
	"strings"
)

func Handle(msg string) *string {
	appConfig := config.GetConfig()
	if appConfig == nil {
		return nil
	}
	requestText := strings.TrimSpace(msg)
	reply, err := openai.Completions(requestText, appConfig.ChatGpt.Token)
	if err != nil {
		log.Println(err)
	}
	return reply
}
