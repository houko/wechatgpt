package telegram

import (
	"strings"

	"wechatbot/openai"

	log "github.com/sirupsen/logrus"
)

func Handle(sender string, msg string) string {
	requestText := strings.TrimSpace(msg)
	reply, err := openai.GlobalSession.Completions(sender, requestText, nil)
	if err != nil {
		log.Error(err)
	}
	return reply
}
