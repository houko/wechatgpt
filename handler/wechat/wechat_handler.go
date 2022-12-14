package wechat

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/wechatgpt/wechatbot/config"
	"github.com/wechatgpt/wechatbot/openai"
	"log"
	"strings"
)

var _ MessageHandlerInterface = (*GroupMessageHandler)(nil)

type GroupMessageHandler struct {
}

func (gmh *GroupMessageHandler) handle(msg *openwechat.Message) error {
	if !msg.IsText() {
		return nil
	}
	return gmh.ReplyText(msg)
}

func NewGroupMessageHandler() MessageHandlerInterface {
	return &GroupMessageHandler{}
}

func (gmh *GroupMessageHandler) ReplyText(msg *openwechat.Message) error {

	sender, err := msg.Sender()
	group := openwechat.Group{User: sender}
	log.Printf("Received Group %v Text Msg : %v", group.NickName, msg.Content)

	keyword := "chatgpt"
	appConfig := config.GetConfig()
	if appConfig != nil {
		keyword = appConfig.ChatGpt.Keyword
	}

	if !strings.Contains(msg.Content, keyword) {
		return nil
	}
	splitItems := strings.Split(msg.Content, keyword)
	if len(splitItems) < 2 {
		return nil
	}
	requestText := strings.TrimSpace(splitItems[1])
	reply, err := openai.Completions(requestText)
	if err != nil {
		log.Println(err)
		text, err := msg.ReplyText(fmt.Sprintf("bot error: %s", err.Error()))
		if err != nil {
			return err
		}
		log.Println(text)
		return err
	}
	// 如果在提问的时候没有包含？,AI会自动在开头补充个？看起来很奇怪
	result := *reply
	if strings.HasPrefix(result, "?") {
		result = strings.Replace(result, "?", "", -1)
	}
	if strings.HasPrefix(result, "？") {
		result = strings.Replace(result, "？", "", -1)
	}
	// 微信不支持markdown格式，所以把反引号直接去掉
	if strings.Contains(result, "`") {
		result = strings.Replace(result, "`", "", -1)
	}

	if reply != nil {
		_, err = msg.ReplyText(*reply)
		if err != nil {
			log.Println(err)
		}
		return err
	}

	return nil
}
