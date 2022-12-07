package wechat

import (
	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"
)

type MessageHandlerInterface interface {
	handle(*openwechat.Message) error
	ReplyText(*openwechat.Message) error
}

type Type string

const (
	GroupHandler = "group"
)

var handlers map[Type]MessageHandlerInterface

func init() {
	handlers = make(map[Type]MessageHandlerInterface)
	handlers[GroupHandler] = NewGroupMessageHandler()
}

func Handler(msg *openwechat.Message) {
	err := handlers[GroupHandler].handle(msg)
	if err != nil {
		log.Errorf("handle error: %s\n", err.Error())
		return
	}
}
