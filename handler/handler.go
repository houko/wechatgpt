package handler

import (
	"github.com/eatmoreapple/openwechat"
	"github.com/prometheus/common/log"
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
	if msg.IsSendByGroup() {
		err := handlers[GroupHandler].handle(msg)
		if err != nil {
			log.Errorf("handle error: %s\n", err.Error())
			return
		}
	}
}
