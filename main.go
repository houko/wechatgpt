package main

import (
	"github.com/wechatgpt/wechatbot/bootstrap"
	"github.com/wechatgpt/wechatbot/config"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	bootstrap.StartTelegramBot()
	bootstrap.StartWebChat()
}
