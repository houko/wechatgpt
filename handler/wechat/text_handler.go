package wechat

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"wechatbot/config"
	"wechatbot/openai"
	"wechatbot/utils"

	"github.com/eatmoreapple/openwechat"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

func TextMessageHandler(msg *Message) error {
	sender, err := msg.Sender()
	group := openwechat.Group{User: sender}
	log.Infof("Received Group %v Text Msg : %v", group.NickName, msg.Content)

	wechat := config.GetWechatKeyword()
	requestText := msg.Content
	if wechat != "" {
		content, key := utils.ContainsI(requestText, wechat)
		if len(key) == 0 {
			return nil
		}

		splitItems := strings.Split(content, key)
		if len(splitItems) < 2 {
			return nil
		}

		requestText = strings.TrimSpace(splitItems[1])
	}

	log.Infof("问题：%s, typ: %v", requestText, msg.typ)
	reply := ""
	if msg.typ == VisionMessageText {
		reply, err = openai.GlobalSession.Completions(sender.UserName, requestText, lo.Map(msg.related, func(m *Message, i int) string {
			return m.imagePathIfPicture
		}))
		messageCache.Delete(sender.UserName)
	} else if msg.typ == ImageGenMessage {
		reply, err = openai.GlobalSession.ImageGeneration(sender.UserName, requestText)
	} else if msg.typ == TextMessage {
		reply, err = openai.GlobalSession.Completions(sender.UserName, requestText, nil)
	}

	if err != nil {
		log.Errorf("Failed to get reply: %v", err)
		// 一次只能回复4000个字
		if len(reply) > 4000 {
			for i := 0; i < len(reply); i += 4000 {
				if i+4000 > len(reply) {
					_, err = msg.ReplyText(reply[i:])
				} else {
					_, err = msg.ReplyText(reply[i : i+4000])
				}
				if err != nil {
					return errors.Wrap(err, "failed to reply message")
				}
			}
		}

		text, err := msg.ReplyText(fmt.Sprintf("bot error: %s", err.Error()))
		return errors.Wrap(err, fmt.Sprintf("failed to reply message: %v", text))
	}

	// 如果在提问的时候没有包含？,AI会自动在开头补充个？看起来很奇怪
	if strings.HasPrefix(reply, "?") {
		reply = strings.Replace(reply, "?", "", -1)
	}

	if strings.HasPrefix(reply, "？") {
		reply = strings.Replace(reply, "？", "", -1)
	}

	// 微信不支持markdown格式，所以把反引号直接去掉
	if strings.Contains(reply, "`") {
		reply = strings.Replace(reply, "`", "", -1)
	}

	if reply != "" {
		_, err = msg.ReplyText(reply)
		if err != nil {
			return errors.Wrap(err, "failed to reply message")
		}
	}

	return nil
}
