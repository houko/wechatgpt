package wechat

import (
	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"
)

var Handler = openwechat.NewMessageMatchDispatcher()

func init() {
	Handler.SetAsync(true)
	Handler.OnText(RawTextMessageHandler)
	Handler.OnImage(RawImageMessageHandler)
}

func RawTextMessageHandler(ctx *openwechat.MessageContext) {
	msg, err := WrapMessage(ctx.Message)
	if err != nil {
		log.Errorf("Failed to wrap message: %v", err)
		return
	}
	log.Debugf("Received Text Msg : %v", msg.Content)
	err = TextMessageHandler(msg)
	if err != nil {
		log.Errorf("Failed to handle message: %v", err)
		return
	}
}

func RawImageMessageHandler(ctx *openwechat.MessageContext) {
	msg, err := WrapMessage(ctx.Message)
	if err != nil {
		log.Errorf("Failed to wrap message: %v", err)
		return
	}
	_, err = msg.ReplyText("请问您需要了解关于这张图片的什么问题？或者继续发送更多图片让我分析吧！")
	if err != nil {
		log.Errorf("Failed to reply message: %v", err)
		return
	}
}
