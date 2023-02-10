package config

import (
	"fmt"
	"os"
	"strings"
	"github.com/spf13/viper"
	"strconv"
)

var config *Config

type Config struct {
	ChatGpt ChatGptConfig `json:"chatgpt" mapstructure:"chatgpt" yaml:"chatgpt"`
}

type ChatGptConfig struct {
	Token         string  `json:"token,omitempty"  mapstructure:"token,omitempty"  yaml:"token,omitempty"`
	Wechat        *string `json:"wechat,omitempty" mapstructure:"wechat,omitempty" yaml:"wechat,omitempty"`
	WechatKeyword *string `json:"wechat_keyword"   mapstructure:"wechat_keyword"   yaml:"wechat_keyword"`
	Model         *string `json:"model,omitempty"  mapstructure:"model"            yaml:"model"`
	MaxLen        *int    `json:"maxlen,omitempty" mapstructure:"maxlen"           yaml:"maxlen"`
	Telegram      *string `json:"telegram"         mapstructure:"telegram"         yaml:"telegram"`
	TgWhitelist   *string `json:"tg_whitelist"     mapstructure:"tg_whitelist"     yaml:"tg_whitelist"`
	TgKeyword     *string `json:"tg_keyword"       mapstructure:"tg_keyword"       yaml:"tg_keyword"`
}

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./local")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&config); err != nil {
		return err
	}
	return nil
}

func GetWechat() *string {
	wechat := getEnv("wechat")

	if wechat != nil {
		return wechat
	}
	if config == nil {
		return nil
	}
	if wechat == nil {
		wechat = config.ChatGpt.Wechat
	}
	return wechat
}

func GetWechatKeyword() *string {
	keyword := getEnv("wechat_keyword")

	if keyword != nil {
		return keyword
	}
	if config == nil {
		return nil
	}
	if keyword == nil {
		keyword = config.ChatGpt.WechatKeyword
	}
	return keyword
}

func GetModelType() *string {
	keyword := getEnv("model")

	if keyword != nil {
		return keyword
	}
	if config == nil {
		return nil
	}
	if keyword == nil {
		keyword = config.ChatGpt.Model
	}
	return keyword
}

func GetMaxLen() *int {
	keyword := getEnv("maxlen")

	if keyword != nil {
		maxlen, _ := strconv.Atoi(*keyword)
		return &maxlen
	}
	if config == nil {
		return nil
	}
	if keyword == nil {
		return config.ChatGpt.MaxLen
	}
	return nil
}

func GetTelegram() *string {
	tg := getEnv("telegram")
	fmt.Println(tg)
	if tg != nil {
		return tg
	}
	if config == nil {
		return nil
	}
	if tg == nil {
		tg = config.ChatGpt.Telegram
	}
	return tg
}

func GetTelegramKeyword() *string {
	tgKeyword := getEnv("tg_keyword")

	if tgKeyword != nil {
		return tgKeyword
	}
	if config == nil {
		return nil
	}
	if tgKeyword == nil {
		tgKeyword = config.ChatGpt.TgKeyword
	}
	return tgKeyword
}

func GetTelegramWhitelist() *string {
	tgWhitelist := getEnv("tg_whitelist")

	if tgWhitelist != nil {
		return tgWhitelist
	}
	if config == nil {
		return nil
	}
	if tgWhitelist == nil {
		tgWhitelist = config.ChatGpt.TgWhitelist
	}
	return tgWhitelist
}

func GetOpenAiApiKey() *string {
	apiKey := getEnv("token")

	if apiKey != nil {
		return apiKey
	}

	if config == nil {
		return nil
	}
	if apiKey == nil {
		apiKey = &config.ChatGpt.Token
	}
	return apiKey
}

func getEnv(key string) *string {
	value := os.Getenv(key)
	if len(value) == 0 {
		value = os.Getenv(strings.ToUpper(key))
	}

	if len(value) > 0 {
		return &value
	}

	if config == nil {
		return nil
	}

	if len(value) > 0 {
		return &value
	} else if config.ChatGpt.WechatKeyword != nil {
		value = *config.ChatGpt.WechatKeyword
	}
	return nil
}
