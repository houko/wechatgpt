package bootstrap

import (
	"wechatbot/handler/wechat"

	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"
)

func StartWebChat() {
	log.Info("Start WebChat Bot")
	bot := openwechat.DefaultBot(openwechat.Desktop)
	bot.MessageHandler = wechat.Handler.AsMessageHandler()
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()
	err := bot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption())
	if err != nil {
		log.Fatal(err)
		return
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	if err != nil {
		log.Fatal(err)
		return
	}

	friends, err := self.Friends()
	for i, friend := range friends {
		log.Println(i, friend)
	}

	groups, err := self.Groups()
	for i, group := range groups {
		log.Println(i, group)
	}

	err = bot.Block()
	if err != nil {
		log.Fatal(err)
		return
	}
}
