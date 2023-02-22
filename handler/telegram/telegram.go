package telegram

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wechatgpt/wechatbot/openai"
)

func Handle(msg string) *string {
	requestText := strings.TrimSpace(msg)
	reply, err := openai.Completions(requestText)
	if err != nil {
		log.Println(err)
	}
	return reply
}
